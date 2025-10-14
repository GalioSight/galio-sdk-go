// Copyright 2024 Tencent Galileo Authors

package ocp

import (
	"testing"
	"time"

	"galiosight.ai/galio-sdk-go/errs"
	"galiosight.ai/galio-sdk-go/model"
	"github.com/stretchr/testify/assert"
)

type mockWatcher struct {
	config *GalileoConfig
}

func (m *mockWatcher) Watch(config *GalileoConfig) {
	m.config = config
}

func local(to *GalileoConfig) error {
	to.Config = *DefaultConfig("")
	return nil
}

func TestUpdater(t *testing.T) {
	testServer, rsp := newTestServer()
	defer testServer.Close()
	resource := model.NewResource(
		"Galileo-Dial",
		"galileo",
		"SDK",
		"DemoService",
		model.Production,           // 物理环境，只能在 model.Production 和 model.Development 枚举，正式环境必须是 model.Production
		"formal",                   // 用户环境，一般是 formal 和 test (或形如 3c170118 等自定义), 正式环境必须是 formal
		"set.sz.1",                 // set 名，可以为空
		"sz",                       // 城市，可以为空
		"127.0.0.1",                // 实例 IP，可以为空
		"test.galileo.SDK.sz10010", // 容器名，可以为空
	)
	RegisterResource(
		resource, WithLocalDecoder(DecodeFunc(local)), WithOcpAddr(""), WithDuration(time.Microsecond),
		WithVerbose("info"),
	)
	updater := GetUpdater(resource.Target)
	assert.NotNil(t, updater)
	assert.Equal(t, time.Microsecond, updater.duration)
	assert.Equal(t, "info", updater.config.Verbose)

	// 注册一个 mock watcher
	watcher := &mockWatcher{}
	AddWatcher(resource.Target, watcher)
	assert.Equal(t, 1, len(updater.watchers))

	update := updater.Update()
	assert.Equal(t, false, update)
	assert.NotNil(t, updater.config)

	updater.config.OcpAddr = testServer.URL
	update = updater.Update()
	assert.Equal(t, true, update)
	assert.Equal(t, "galileo", updater.config.Config.TenantId)

	// rsp 使用的 json.Unmarshal 赋值，对于数组会重用，参见 !514 的 bug
	b := &rsp.MetricsConfig.Processor.HistogramBuckets[0].Buckets
	*b = append(*b, 10)
	update = updater.Update()
	assert.Equal(t, true, update)

	update = updater.Update()
	assert.Equal(t, false, update)

	// 检查 mock watcher 是否收到了新的配置
	assert.Equal(t, &updater.config, watcher.config)
	assert.Equal(t, &updater.config, updater.GetConfig())

	rsp.Version = -2
	rsp.Msg = "hello"
	updater.Update()
	// version 版本低，应该使用 local 配置
	assert.NotEqual(t, rsp.Msg, updater.GetConfig().Config.Msg)

	rsp.Version = 1
	rsp.Msg = "success"
	updater.Update()
	assert.Equal(t, rsp.Msg, updater.GetConfig().Config.Msg)

	err := RegisterResource(
		resource,
		WithOcpAddr(testServer.URL),
		WithLocalDecoder(DecodeFunc(local)),
		WithDuration(time.Microsecond),
	)
	assert.Equal(t, errs.ErrResourceAlreadyRegistered, err)
	go updater.updateByTick()
	time.Sleep(time.Millisecond)

	assert.Equal(t, "galileo", updater.GetConfig().Resource.TenantId)
	resource.TenantId = "galileo"
	assert.Equal(t, *resource, updater.GetConfig().Resource)
	assert.Equal(t, testServer.URL, GetUpdater(resource.Target).GetConfig().OcpAddr)
	err = UnregisterResource(resource.Target)
	assert.NoError(t, err)
}

// TestUnregisterResource 测试 UnregisterResource 函数
func TestUnregisterResource(t *testing.T) {
	resource := &model.Resource{
		Target:   "a",
		TenantId: "galileo",
	}
	// 注册资源
	err := RegisterResource(resource)
	assert.NoError(t, err, "Expected no error when registering resource")

	// 测试注销成功
	err = UnregisterResource(resource.Target)
	assert.NoError(t, err, "Expected no error when unregistering resource")

	// 测试注销不存在的资源
	err = UnregisterResource(resource.Target)
	assert.Error(t, err, "Expected error when unregistering a non-registered resource")
	assert.Equal(t, errs.ErrTargetNotExist, err)
}
