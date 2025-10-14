// Copyright 2021 Tencent Galileo Authors
//
// Copyright 2021 Tencent OpenTelemetry Oteam
//
// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package profiles

import (
	"errors"
	"io"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
	// sonicStopProfiling 必须使用 unsafe
	_ "unsafe"

	"galiosight.ai/galio-sdk-go/components"
	"galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/processors/omp/profiles/delta"
)

//go:linkname sonicStopProfiling github.com/bytedance/sonic/internal/rt.StopProfiling
var sonicStopProfiling bool

// processor 性能数据处理器实现。
type processor struct {
	// exporter 性能数据导出器。
	exporter components.ProfilesExporter
	// cfg 配置。
	cfg *configs.Profiles
	// 开启增量收集的 profile 类型
	deltas map[model.ProfileType]delta.Profiler
	// 用于控制 CPU profiles 等待其他类型 profiles 完成收集后，最后完成收集，
	// 从而完整的收集 profiles 数据
	pendingProfiles sync.WaitGroup

	seq int64
	// 开启的 profiles 类型
	enabledProfiles map[model.ProfileType]bool

	// 确保 Shutdown 只执行一次
	stopOnce sync.Once
	// 退出时，控制 exporter 退出
	stopCh chan struct{}
	// stop 时，等待所有 goroutines 退出
	stopWait sync.WaitGroup

	// stats 自监控状态。
	stats *model.SelfMonitorStats
}

var _ components.ProfilesProcessor = (*processor)(nil)

// NewProcessor 构造日志处理器。
func NewProcessor(
	cfg *configs.Profiles,
	exporter components.ProfilesExporter,
) (components.ProfilesProcessor, error) {
	p := newProcessor(cfg, exporter)
	return p, nil
}

// Start 开始采集性能数据。
func (p *processor) Start() {
	p.run()
}

// Shutdown 停止采集新能数据
func (p *processor) Shutdown() {
	p.stopOnce.Do(
		func() {
			if p.exporter != nil {
				p.exporter.Shutdown()
			}
			p.stop()
		},
	)
}

func newProcessor(cfg *configs.Profiles, exporter components.ProfilesExporter) *processor {
	fixConfig(cfg)
	p := &processor{
		cfg:             cfg,
		exporter:        exporter,
		stats:           cfg.Stats,
		stopCh:          make(chan struct{}),
		deltas:          make(map[model.ProfileType]delta.Profiler),
		enabledProfiles: make(map[model.ProfileType]bool),
	}
	p.addProfileTypes(cfg.Processor.ProfileTypes)
	return p
}

func (p *processor) run() {
	profileEnabled := func(t model.ProfileType) bool {
		_, ok := p.enabledProfiles[t]
		return ok
	}

	if profileEnabled(model.MutexProfile) {
		runtime.SetMutexProfileFraction(int(p.cfg.Processor.MutexProfileFraction))
	}
	if profileEnabled(model.BlockProfile) {
		runtime.SetBlockProfileRate(int(p.cfg.Processor.BlockProfileRate))
	}
	sonicStopProfiling = true
	p.stopWait.Add(1)
	go func() {
		defer p.stopWait.Done()
		ticker := time.NewTicker(time.Duration(p.cfg.Processor.PeriodSeconds) * time.Second)
		defer ticker.Stop()
		p.collect(ticker.C)
	}()
}

func (p *processor) stop() {
	close(p.stopCh)
	p.stopWait.Wait()
	p.cfg.Log.Infof("[galileo]profiles processor stopped")
}

func (p *processor) addProfileTypes(types []string) {
	for _, pt := range types {
		if _, ok := p.enabledProfiles[model.ProfileType(pt)]; !ok {
			p.enabledProfiles[model.ProfileType(pt)] = true
			p.setDeltaProfiler(model.ProfileType(pt))
		}
	}
}

func (p *processor) setDeltaProfiler(pt model.ProfileType) {
	if d := profileCollectors[pt].DeltaValues; len(d) > 0 {
		p.deltas[pt] = delta.NewFastProfiler(d)
	}
}

// interruptibleSleep 休眠一段时间 d 或被 p.stopCh 中断
func (p *processor) interruptibleSleep(d time.Duration) {
	select {
	case <-p.stopCh:
	case <-time.After(d):
	}
}

func (p *processor) lookupProfile(name string, w io.Writer, debug int) error {
	prof := pprof.Lookup(name)
	if prof == nil {
		return errors.New("profile not found")
	}
	return prof.WriteTo(w, debug)
}

func (p *processor) collect(ticker <-chan time.Time) {
	var (
		// mu guards completed
		batchMutex sync.Mutex
		completed  []*model.Profile
		batchWait  sync.WaitGroup
	)

	for {
		p.seq++
		batch := &model.ProfilesBatch{
			Sequence: p.seq,
			Start:    time.Now().Unix(),
			Resource: &p.cfg.Resource,
		}

		completed = completed[:0]
		// 非 CPU profile, 通过 pendingProfiles 等待其完成
		profileTypes := p.enabledProfileTypes()
		for _, t := range profileTypes {
			if t != model.CPUProfile {
				p.pendingProfiles.Add(1)
			}
		}
		for _, t := range profileTypes {
			batchWait.Add(1)
			go func(t model.ProfileType) {
				defer batchWait.Done()
				if t != model.CPUProfile {
					defer p.pendingProfiles.Done()
				}
				profs, err := p.collectProfile(t)
				if err != nil {
					p.cfg.Log.Errorf(
						"[galileo]profiles processor getting %s profile failed: %v; seq: %v, skipping.", t, err, p.seq,
					)
				}
				batchMutex.Lock()
				defer batchMutex.Unlock()
				completed = append(completed, profs...)
			}(t)
		}
		batchWait.Wait()
		for _, prof := range completed {
			addProfile(batch, prof)
		}

		select {
		case <-ticker:
		case <-p.stopCh:
			return
		}

		batch.End = time.Now().Unix()
		// batch 导出 profile 数据到导出器.
		p.exporter.Export(batch)
	}
}

// enabledProfileTypes 按顺序返回 enabled profiles 的类型。
// CPU Profile 在第一位，因为多数用户都会关注 CPU profiles
func (p *processor) enabledProfileTypes() []model.ProfileType {
	order := []model.ProfileType{
		model.CPUProfile,
		model.HeapProfile,
		model.BlockProfile,
		model.MutexProfile,
		model.GoroutineProfile,
	}
	enabled := []model.ProfileType{}
	for _, t := range order {
		if _, ok := p.enabledProfiles[t]; ok {
			enabled = append(enabled, t)
		}
	}
	return enabled
}

func (p *processor) collectProfile(pt model.ProfileType) ([]*model.Profile, error) {
	t := getProfileCollector(pt)
	if t.Name == "unknown" {
		p.cfg.Log.Errorf("[galileo]profiles processor profile type %s is not supported", string(pt))
	}
	data, err := t.Collect(p)
	if err != nil {
		return nil, err
	}
	filename := t.Filename
	if p.cfg.Processor.EnableDeltaProfiles && len(t.DeltaValues) > 0 {
		filename = "delta-" + filename
	}
	return []*model.Profile{
		{
			Name: filename,
			Type: string(pt),
			Data: data,
		},
	}, nil
}

func addProfile(b *model.ProfilesBatch, p *model.Profile) {
	b.Profiles = append(b.Profiles, p)
}
