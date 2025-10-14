// Copyright 2024 Tencent Galileo Authors

package ocp

import (
	"context"
	"sync"
	"time"

	"galiosight.ai/galio-sdk-go/errs"
	"galiosight.ai/galio-sdk-go/model"
	selflog "galiosight.ai/galio-sdk-go/self/log"
	"github.com/pmezard/go-difflib/difflib"
	yaml "gopkg.in/yaml.v3"
)

var (
	// updater ocp 配置更新器，每个 target 对应一个 Updater。
	updaters = map[string]*Updater{}
	// 保护 updaters，避免并发读写
	rwMutex sync.RWMutex
)

// RegisterResource 注册 resource，注册后此 resource 的 ocp 配置会进行定时拉取以实现热更新。
// resource.Target 是 resource 的唯一主键。
// 每个 Target 只有第一次注册生效，需要 UnregisterResource 之后，才能再次注册。
func RegisterResource(resource *model.Resource, opts ...updaterOption) error {
	rwMutex.Lock()
	defer rwMutex.Unlock()
	_, exist := updaters[resource.Target]
	if exist {
		return errs.ErrResourceAlreadyRegistered
	}
	updaters[resource.Target] = newUpdater(resource, opts...)
	return nil
}

// UnregisterResource 注销注册的 resource，并停止其定时更新配置。
// resource.Target 是 resource 的唯一主键，通过 target 来注销。
func UnregisterResource(target string) error {
	rwMutex.Lock()
	defer rwMutex.Unlock()
	updater, exist := updaters[target]
	if !exist {
		return errs.ErrTargetNotExist
	}
	updater.Close()          // 停止协程
	delete(updaters, target) // 从 map 中移除
	return nil
}

// AddWatcher 添加配置观察者，当对应 target 的配置有更新时，会回调 watcher 将更新后的配置应用到对象中。
// 每个 target 可以添加多个 watcher.
// 必须先注册 RegisterResource，然后才能添加观察者 AddWatcher。
func AddWatcher(target string, watcher Watcher) error {
	rwMutex.RLock()
	defer rwMutex.RUnlock()
	u, exist := updaters[target]
	if !exist {
		return errs.ErrResourceAlreadyRegistered
	}
	u.registerWatcher(watcher)
	return nil
}

// GetUpdater 使用 target 获取配置更新器
func GetUpdater(target string) *Updater {
	rwMutex.RLock()
	defer rwMutex.RUnlock()
	return updaters[target]
}

// GalileoConfig 伽利略插件配置。
type GalileoConfig struct {
	Verbose string `yaml:"verbose" json:"verbose,omitempty"`
	// OcpAddr 服务器地址
	OcpAddr string `yaml:"ocp_addr" json:"ocp_addr,omitempty"`
	// Config 实际生效的配置，由远程配置合并本地配置而成，在调用接口时将 local 报上去，ocp 服务会将本地配置和远程配置合并后返回。
	Config model.GetConfigResponse `yaml:"config" json:"config"`
	// Resource 资源信息，在 SDK 运行期间不会改变。一般是不需要用户配置的，自动从框架配置中获取。
	Resource model.Resource `yaml:"resource" json:"resource"`
	// local 本地配置，初始化之后，不允许修改。此字段仅用于调用 ocp，其他地方不需要访问。
	local model.GetConfigResponse `yaml:"-" json:"-"`
	// 数据上报身份认证, 通过 http 头 X-Galileo-API-Key 报到后台用于鉴权
	APIKey string `yaml:"api_key" json:"api_key,omitempty"`
}

// Updater ocp 配置更新器，用于定时更新 ocp 配置。
type Updater struct {
	config GalileoConfig
	// 定时更新的时间
	duration time.Duration
	// 保护 config，避免并发读写
	rwMutex sync.RWMutex
	// 所有的观察者
	watchers []Watcher
	decoder  decoder
	ctx      context.Context
	cancel   context.CancelFunc
}

// Watcher 观察者，观察配置变化后，使用配置更新自己的状态。
// 通常观察者只读取配置，不应该修改配置，所以输入参数是只读的。
type Watcher interface {
	Watch(readOnlyConfig *GalileoConfig)
}

// newUpdater 创建 Updater 对象。注意 Updater 内部有一个协程，所以此对象不能随意创建。
func newUpdater(resource *model.Resource, opts ...updaterOption) *Updater {
	ctx, cancel := context.WithCancel(context.Background())

	updater := &Updater{
		config: GalileoConfig{
			Verbose:  "error",
			OcpAddr:  DefaultURL,
			Resource: *resource,
			Config:   *DefaultConfig(resource.TenantId),
		},
		duration: time.Minute,
		ctx:      ctx,
		cancel:   cancel,
	}

	for _, opt := range opts {
		opt(updater)
	}
	if updater.decoder != nil {
		_ = updater.decoder.Decode(&updater.config)
		fixResource(&updater.config)
	}
	updater.config.local = updater.config.Config
	selflog.SetLogLevel(updater.config.Verbose)
	updater.Update()          // 启动时更新 ocp 配置
	go updater.updateByTick() // 定时更新 ocp 配置
	return updater
}

// updateByTick 此方法在初始化时会被调用，有一个线程会执行在执行
func (u *Updater) updateByTick() {
	if u.config.OcpAddr == "" {
		return
	}
	tick := time.Tick(u.duration)
	for {
		select {
		case <-u.ctx.Done():
			return // 退出协程
		case <-tick:
			_ = u.Update()
		}
	}
}

// Close 停止 Updater 和其内部协程
func (u *Updater) Close() {
	u.cancel() // 取消上下文，停止协程
}

// Update 从 ocp 更新配置。通常由 Updater 内部的每分钟定时任务自动调用，当需要立即更新时，可以主动调用此方法。
func (u *Updater) Update() bool {
	config, err := GetOcpConfig(
		u.config.OcpAddr, &u.config.Resource, Local(&u.config.local), WithApiKey(u.config.APIKey),
	)
	if err != nil {
		selflog.Errorf(
			"err: %v, ocpAddr: %v, resource: %v, local: %v", err, u.config.OcpAddr,
			u.config.Resource, u.config.local,
		)
		return false
	}
	if u.config.local.Version > config.Version {
		// 兼容历史版本，ocp 会根据本地配置和 web 配置的版本来合并出最终配置，但是当前只对 trace 开放了 web 配置，
		// 并未对所有服务及所有配置项生效，所以在本地配置版本更高时，使用本地本地覆盖一下 ocp 返回来的配置，以保持本地配置优先。
		wrap := &GalileoConfig{
			Resource: u.config.Resource,
			Config:   *config,
		}
		u.decoder.Decode(wrap)
		fixResource(wrap)
		config = &wrap.Config
	}
	oldYaml := u.getConfigYAML()
	newYaml := toYAML(config)
	if oldYaml == newYaml {
		selflog.Infof("config no change, %v", u.config.Resource)
		return false
	}
	logDiffLines(oldYaml, newYaml)
	u.setConfig(config)
	u.notifyAllWatchers()
	return true
}

func fixResource(cfg *GalileoConfig) {
	cfg.Resource.Target = cfg.Resource.Platform + "." + cfg.Resource.ObjectName
	cfg.Config.Target = cfg.Resource.Target
}

// logDiffLines 输出 YAML 配置变化的行，方便定位问题。
func logDiffLines(oldGalileoYaml, newGalileoYaml string) {
	diff, err := difflib.GetUnifiedDiffString(
		difflib.UnifiedDiff{
			A:        difflib.SplitLines(oldGalileoYaml),
			B:        difflib.SplitLines(newGalileoYaml),
			FromFile: "oldGalileoYaml",
			FromDate: "",
			ToFile:   "newGalileoYaml",
			ToDate:   "",
			Context:  1,
		},
	)
	selflog.Infof("[galileo]updateGalileoConfig|err=%v,diff=\n%s,", err, diff)
}

func (u *Updater) notifyAllWatchers() {
	u.rwMutex.RLock()
	defer u.rwMutex.RUnlock()
	for i := range u.watchers {
		u.watchers[i].Watch(&u.config)
	}
}

// registerWatcher 注册观察器。需要热更新配置的对象，都需要注册到 Updater 中。
func (u *Updater) registerWatcher(w Watcher) {
	u.rwMutex.Lock()
	defer u.rwMutex.Unlock()
	u.watchers = append(u.watchers, w)
}

// setConfig 设置配置，此方法是线程安全的
func (u *Updater) setConfig(cfg *model.GetConfigResponse) {
	u.rwMutex.Lock()
	u.config.Config = *cfg
	u.config.Resource.TenantId = u.config.Config.TenantId
	u.rwMutex.Unlock()
}

// GetConfig 获取配置，此方法是线程安全的
func (u *Updater) GetConfig() *GalileoConfig {
	u.rwMutex.RLock()
	defer u.rwMutex.RUnlock()
	return &u.config
}

// getConfigYAML 获取配置的 YAML，此方法是线程安全的
func (u *Updater) getConfigYAML() string {
	u.rwMutex.RLock()
	defer u.rwMutex.RUnlock()
	return toYAML(&u.config.Config)
}

// toYAML 将 cfg 序列化成 YAML，此方法不是线程安全的，注意不要对 cfg 进行并发读写
func toYAML(cfg *model.GetConfigResponse) string {
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return ""
	}
	return string(out)
}

type updaterOption func(u *Updater)

// WithDuration 设置定时刷新的时间间隔
func WithDuration(duration time.Duration) updaterOption {
	return func(u *Updater) {
		u.duration = duration
	}
}

// WithOcpAddr 设置 ocp 地址
func WithOcpAddr(ocpAddr string) updaterOption {
	return func(u *Updater) {
		u.config.OcpAddr = ocpAddr
	}
}

type decoder interface {
	Decode(to *GalileoConfig) error
}

type DecodeFunc func(to *GalileoConfig) error

// Decode ...
func (fn DecodeFunc) Decode(to *GalileoConfig) error {
	return fn(to)
}

// WithLocalDecoder 设置本地配置
func WithLocalDecoder(local decoder) updaterOption {
	return func(u *Updater) {
		u.decoder = local
	}
}

// WithVerbose 设置调试级别
func WithVerbose(verbose string) updaterOption {
	return func(u *Updater) {
		u.config.Verbose = verbose
	}
}
