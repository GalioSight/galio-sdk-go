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

package metrics

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/configs/ocp"
	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/lib/strings"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/protocols"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
)

// exporter 模拟的导出器，仅做简单收集、统计。
type exporter struct {
	normalLabels *model.NormalLabels
	clients      []*model.ClientMetricsOTP
	servers      []*model.ServerMetricsOTP
	normals      []*model.NormalMetricOTP
	customs      []*model.CustomMetricsOTP
	count        atomic.Int64
}

func newExporter() *exporter {
	return &exporter{}
}

// Export 模拟导出，仅做简单收集、统计。
func (e *exporter) Export(metrics *model.Metrics) {
	e.normalLabels = metrics.NormalLabels
	e.clients = append(e.clients, metrics.ClientMetrics...)
	e.servers = append(e.servers, metrics.ServerMetrics...)
	e.normals = append(e.normals, metrics.NormalMetrics...)
	e.customs = append(e.customs, metrics.CustomMetrics...)
	c := len(metrics.ClientMetrics) + len(metrics.ServerMetrics) +
		len(metrics.NormalMetrics) + len(metrics.CustomMetrics)
	e.count.Add(int64(c))
}

// UpdateConfig 更新配置。
func (e *exporter) UpdateConfig(_ *configs.Metrics) {
}

func (e *exporter) GetStats() *model.SelfMonitorStats {
	return nil
}

func newProcessorCfg() *configs.Metrics {
	return &configs.Metrics{
		Log: logs.NopWrapper(),
		Resource: model.Resource{
			Target: "PCG-123.xxx.yyy",
		},
		Processor: model.MetricsProcessor{
			Protocol:       protocols.OMP,
			WindowSeconds:  1,
			ClearSeconds:   10,
			ExpiresSeconds: 100,
			PointLimit:     math.MaxInt64,
		},
		ConvertName: true,
	}
}

// TestProcessClientMetrics 测试主调监控处理。
func TestProcessClientMetrics(t *testing.T) {
	// 构造处理器。
	exporter := newExporter()
	cfg := newProcessorCfg()
	processor, err := NewProcessor(cfg, exporter)
	assert.Nil(t, err)
	// 处理数据。
	clientMetrics := model.GetClientMetrics(2)
	defer model.PutClientMetrics(clientMetrics)
	clientMetrics.RpcLabels.Fields[0].Name = model.RPCLabels_callee_ip
	clientMetrics.RpcLabels.Fields[0].Value = "e.f.g.hh"
	clientMetrics.RpcLabels.Fields[1].Name = model.RPCLabels_caller_ip
	clientMetrics.RpcLabels.Fields[1].Value = "e.f.gg.h"
	clientMetrics.Metrics[model.ClientMetricStartedTotalPoint].Value = 1
	clientMetrics.Metrics[model.ClientMetricHandledTotalPoint].Value = 1
	clientMetrics.Metrics[model.ClientMetricHandledSecondsPoint].Value = 0.01
	processor.ProcessClientMetrics(clientMetrics)
	clientMetrics.Metrics[model.ClientMetricStartedTotalPoint].Value = 1
	clientMetrics.Metrics[model.ClientMetricHandledTotalPoint].Value = 1
	clientMetrics.Metrics[model.ClientMetricHandledSecondsPoint].Value = 0.1
	processor.ProcessClientMetrics(clientMetrics)
	// 断言数量统计。
	assert.Equal(t, int64(1), processor.GetStats().MultiCount.Load())
	assert.Equal(t, int64(3), processor.GetStats().PointCount.Load())
	// 等待数据导出。
	time.Sleep(time.Duration(cfg.Processor.WindowSeconds*2) * time.Second)
	// 断言导出。
	assert.Equal(t, model.NormalLabels_target, exporter.normalLabels.Fields[model.NormalLabels_target].Name)
	assert.Equal(t, "PCG-123.xxx.yyy", exporter.normalLabels.Fields[model.NormalLabels_target].Value)
	assert.Equal(t, true, exporter.count.Load() >= int64(1))
	assert.Equal(t, true, len(exporter.clients) >= 1)
	sum := 0.01 + 0.1
	count := int64(2)
	ranges := []string{
		"1.000e-02...2.500e-02",
		"1.000e-01...2.500e-01",
	}
	counts := []int64{1, 1}
	otp := model.ClientMetricsOTP{
		RpcClientStartedTotal:   2,
		RpcClientHandledTotal:   2,
		RpcClientHandledSeconds: model.NewHistogram(count, sum, counts, ranges),
		RpcLabels:               &clientMetrics.RpcLabels,
	}
	assert.Equal(t, otp.String(), exporter.clients[0].String())
}

// TestProcessServerMetrics 测试被调监控处理。
func TestProcessServerMetrics(t *testing.T) {
	// 构造处理器。
	exporter := newExporter()
	cfg := newProcessorCfg()
	processor, err := NewProcessor(cfg, exporter)
	assert.Nil(t, err)
	// 处理数据。
	serverMetrics := model.GetServerMetrics(2)
	defer model.PutServerMetrics(serverMetrics)
	serverMetrics.RpcLabels.Fields[0].Name = model.RPCLabels_callee_ip
	serverMetrics.RpcLabels.Fields[0].Value = "aa.b.c.d"
	serverMetrics.RpcLabels.Fields[1].Name = model.RPCLabels_caller_ip
	serverMetrics.RpcLabels.Fields[1].Value = "ee.f.g.h"
	serverMetrics.Metrics[model.ServerMetricStartedTotalPoint].Value = 1
	serverMetrics.Metrics[model.ServerMetricHandledTotalPoint].Value = 1
	serverMetrics.Metrics[model.ServerMetricHandledSecondsPoint].Value = 0.01
	processor.ProcessServerMetrics(serverMetrics)
	serverMetrics.Metrics[model.ServerMetricStartedTotalPoint].Value = 1
	serverMetrics.Metrics[model.ServerMetricHandledTotalPoint].Value = 1
	serverMetrics.Metrics[model.ServerMetricHandledSecondsPoint].Value = 0.1
	processor.ProcessServerMetrics(serverMetrics)
	// 等待数据导出。
	time.Sleep(time.Duration(cfg.Processor.WindowSeconds*2) * time.Second)
	// 断言。
	assert.Equal(t, true, exporter.count.Load() >= int64(1))
	assert.Equal(t, true, len(exporter.servers) >= 1)
	sum := 0.01 + 0.1
	count := int64(2)
	ranges := []string{
		"1.000e-02...2.500e-02",
		"1.000e-01...2.500e-01",
	}
	counts := []int64{1, 1}
	otp := model.ServerMetricsOTP{
		RpcServerStartedTotal:   2,
		RpcServerHandledTotal:   2,
		RpcServerHandledSeconds: model.NewHistogram(count, sum, counts, ranges),
		RpcLabels:               &serverMetrics.RpcLabels,
	}
	assert.Equal(t, otp.String(), exporter.servers[0].String())
}

// TestProcessNormalMetric 测试属性监控处理。
func TestProcessNormalMetric(t *testing.T) {
	// 构造处理器。
	exporter := newExporter()
	cfg := newProcessorCfg()
	processor, err := NewProcessor(cfg, exporter)
	assert.Nil(t, err)
	// 处理数据。
	normalMetrics := model.GetNormalMetric()
	defer model.PutNormalMetric(normalMetrics)
	normalMetrics.Metric.Name = "test_ProcessNormalMetric"
	normalMetrics.Metric.Aggregation = model.Aggregation_AGGREGATION_COUNTER
	normalMetrics.Metric.Value = 1
	processor.ProcessNormalMetric(normalMetrics)
	normalMetrics.Metric.Value = 10
	processor.ProcessNormalMetric(normalMetrics)
	// 等待数据导出。
	time.Sleep(time.Duration(cfg.Processor.WindowSeconds*2) * time.Second)
	// 断言。
	assert.Equal(t, true, exporter.count.Load() >= int64(1))
	assert.Equal(t, true, len(exporter.normals) >= 1)
	otp := model.NormalMetricOTP{
		Metric: &model.MetricOTP{},
	}
	otp.SetName(0, normalMetrics.Metric.Name)
	otp.SetAggregation(0, normalMetrics.Metric.Aggregation)
	otp.SetCount(0, 11)
	assert.Equal(t, otp.String(), exporter.normals[0].String())
}

// TestProcessCustomMetric 测试自定义监控处理。
func TestProcessCustomMetric(t *testing.T) {
	// 构造处理器。
	exporter := newExporter()
	cfg := newProcessorCfg()
	processor, err := NewProcessor(cfg, exporter)
	assert.Nil(t, err)
	// 处理数据。
	customLabels := []model.Label{
		{
			Name:  "k1",
			Value: "v1",
		},
	}
	processor.ProcessCustomMetrics(
		&model.CustomMetrics{
			Metrics: []model.Metric{
				{
					Name:        "test_custom_metric",
					Aggregation: model.Aggregation_AGGREGATION_AVG,
					Value:       1,
				},
			},
			CustomLabels: customLabels,
		},
	)
	processor.ProcessCustomMetrics(
		&model.CustomMetrics{
			Metrics: []model.Metric{
				{
					Name:        "test_custom_metric",
					Aggregation: model.Aggregation_AGGREGATION_AVG,
					Value:       3,
				},
			},
			CustomLabels: customLabels,
		},
	)
	// 等待数据导出。
	time.Sleep(time.Duration(cfg.Processor.WindowSeconds*2) * time.Second)
	// 断言。
	assert.Equal(t, true, exporter.count.Load() >= int64(1))
	assert.Equal(t, true, len(exporter.customs) >= 1)
	otp := model.CustomMetricsOTP{
		Metrics: []*model.MetricOTP{
			{
				Name: "custom_counter_default_test_custom_metric",
			},
		},
		CustomLabels: []*model.Label{
			{
				Name:  "k1",
				Value: "v1",
			},
		},
		MonitorName: "default",
	}
	otp.SetAggregation(0, model.Aggregation_AGGREGATION_AVG)
	otp.SetAvg(0, 4, 2)
	assert.Equal(t, otp.String(), exporter.customs[0].String())
}

// TestHashCollision 简单测试 hash 碰撞（完备测试需要大量资源）。
func TestHashCollision(t *testing.T) {
	// 构造处理器。
	exporter := newExporter()
	cfg := newProcessorCfg()
	cfg.Processor.WindowSeconds = 10
	cfg.Processor.ClearSeconds = 10
	cfg.Processor.ExpiresSeconds = 10
	cfg.ConvertName = false
	processor, err := NewProcessor(cfg, exporter)
	assert.Nil(t, err)
	// 处理数据。
	c := model.GetCustomMetrics(2, 1)
	defer model.PutCustomMetrics(c)
	const maxSize = 100
	testIDCount := maxSize
	testLayerCount := maxSize
	for i := 0; i < testIDCount; i++ {
		testID := "test_id_" + strconv.Itoa(i)
		for j := 0; j < testLayerCount; j++ {
			testLayer := "test_layer_" + strconv.Itoa(j)
			c.Metrics[0].Name = "test_hash_collision_custom"
			c.Metrics[0].Aggregation = model.Aggregation_AGGREGATION_COUNTER
			c.Metrics[0].Value = 1
			c.CustomLabels[0].Name = "testID"
			c.CustomLabels[0].Value = testID
			c.CustomLabels[1].Name = "testLayer"
			c.CustomLabels[1].Value = testLayer
			processor.ProcessCustomMetrics(c)
		}
	}
	// 断言
	assert.Equal(t, int64(testIDCount*testLayerCount), processor.GetStats().MultiCount.Load())
	assert.Equal(t, int64(testIDCount*testLayerCount), processor.GetStats().PointCount.Load())

	// 处理数据。
	clientMetrics := model.GetClientMetrics(2)
	defer model.PutClientMetrics(clientMetrics)
	callerContainerCount := maxSize
	calleeContainerCount := maxSize
	for i := 0; i < callerContainerCount; i++ {
		callerContainer := "caller_container_sh_" + strconv.Itoa(i)
		for j := 0; j < calleeContainerCount; j++ {
			calleeContainer := "callee_container_sh_" + strconv.Itoa(j)
			clientMetrics.Metrics[0].Name = model.ClientMetrics_rpc_client_started_total
			clientMetrics.Metrics[0].Aggregation = model.Aggregation_AGGREGATION_COUNTER
			clientMetrics.Metrics[0].Value = 1
			clientMetrics.RpcLabels.Fields[0].Name = model.RPCLabels_caller_container
			clientMetrics.RpcLabels.Fields[0].Value = callerContainer
			clientMetrics.RpcLabels.Fields[1].Name = model.RPCLabels_callee_container
			clientMetrics.RpcLabels.Fields[1].Value = calleeContainer
			processor.ProcessClientMetrics(clientMetrics)
		}
	}
	// 断言
	assert.Equal(t, int64(testIDCount*testLayerCount)*2, processor.GetStats().MultiCount.Load())
}

// TestOverloadProtection 测试过载保护。
func TestOverloadProtection(t *testing.T) {
	// 构造处理器。
	exporter := newExporter()
	cfg := newProcessorCfg()
	cfg.Processor.WindowSeconds = 10
	cfg.Processor.ClearSeconds = 10
	cfg.Processor.ExpiresSeconds = 10
	cfg.Processor.PointLimit = 5 // 限制上报 5 个。
	processor, err := NewProcessor(cfg, exporter)
	assert.Nil(t, err)
	// 处理数据。
	c := model.GetCustomMetrics(1, 1)
	defer model.PutCustomMetrics(c)
	testIDCount := 10
	for i := 0; i < testIDCount; i++ {
		testID := "test_id_2" + strconv.Itoa(i)
		c.Metrics[0].Name = "test_overload_protection"
		c.Metrics[0].Aggregation = model.Aggregation_AGGREGATION_COUNTER
		c.Metrics[0].Value = 1
		c.CustomLabels[0].Name = "test_id_2"
		c.CustomLabels[0].Value = testID
		processor.ProcessCustomMetrics(c) // 第 5 个开始丢弃。
	}
	// 断言
	assert.Equal(t, int64(5), processor.GetStats().MultiCount.Load())
	assert.Equal(t, int64(5), processor.GetStats().PointCount.Load())
	assert.Equal(t, int64(5), processor.GetStats().DiscardMultiCount.Load())
}

// TestGoMetrics 测试 go runtime metrics。
func TestGoMetrics(t *testing.T) {
	// 构造处理器。
	exporter := newExporter()
	cfg := newProcessorCfg()
	cfg.Processor.WindowSeconds = 1
	cfg.Processor.ClearSeconds = 10
	cfg.Processor.ExpiresSeconds = 10
	cfg.Processor.EnableProcessMetrics = true
	cfg.Processor.ProcessMetricsSeconds = 1
	_, err := NewProcessor(cfg, exporter)
	assert.Nil(t, err)
	// 等 go runtime 上报。
	time.Sleep(time.Duration(cfg.Processor.ProcessMetricsSeconds*2) * time.Second)
	assert.True(t, exporter.count.Load() != 0)
}

// TestHotUpdate 测试配置热更新。
func TestHotUpdate(t *testing.T) {
	// 构造处理器。
	exporter := newExporter()
	cfg := newProcessorCfg()
	cfg.Processor.WindowSeconds = 1
	cfg.Processor.ClearSeconds = 10
	cfg.Processor.ExpiresSeconds = 10
	cfg.Processor.HistogramBuckets = nil
	cfg.Processor.HistogramBuckets = append(
		cfg.Processor.HistogramBuckets,
		model.HistogramBucket{Name: "name11", Buckets: []float64{0, 1, 2, 3}},
	)
	cfg.Processor.HistogramBuckets = append(
		cfg.Processor.HistogramBuckets,
		model.HistogramBucket{Name: "name22", Buckets: []float64{0, 1, 3}},
	)
	cfg.Log = logs.DefaultWrapper()
	p, err := NewProcessor(cfg, exporter)
	assert.Nil(t, err)
	time.Sleep(time.Duration(cfg.Processor.WindowSeconds) * 2 * time.Second)
	cfg = newProcessorCfg()
	cfg.Processor.HistogramBuckets = nil
	cfg.Processor.HistogramBuckets = append(
		cfg.Processor.HistogramBuckets,
		model.HistogramBucket{Name: "name11", Buckets: []float64{0, 1, 3}},
	)
	cfg.Processor.HistogramBuckets = append(
		cfg.Processor.HistogramBuckets,
		model.HistogramBucket{Name: "name22", Buckets: []float64{0, 1, 3}},
	)
	cfg.Processor.HistogramBuckets = append(
		cfg.Processor.HistogramBuckets,
		model.HistogramBucket{Name: "name33", Buckets: []float64{0, 1, 3}},
	)
	cfg.Processor.WindowSeconds = 2
	p.UpdateConfig(cfg)
	time.Sleep(time.Duration(cfg.Processor.WindowSeconds) * 2 * time.Second)
}

// TestHotUpdateBucket 测试分桶配置热更新。
func TestHotUpdateBucket(t *testing.T) {
	name := "TestHotUpdateBucket"
	// 构造处理器。
	exporter := newExporter()
	cfg := newProcessorCfg()
	cfg.Processor.WindowSeconds = 1
	cfg.Processor.ClearSeconds = 1
	cfg.Processor.ExpiresSeconds = 1
	cfg.Processor.EnableProcessMetrics = false
	cfg.Processor.ProcessMetricsSeconds = 1
	cfg.Processor.HistogramBuckets = nil
	cfg.Processor.HistogramBuckets = append(
		cfg.Processor.HistogramBuckets,
		model.HistogramBucket{Name: name, Buckets: []float64{2, 5}},
	)
	cfg.Log = logs.DefaultWrapper()
	p, err := NewProcessor(cfg, exporter)
	assert.Nil(t, err)
	// 处理数据。
	normalMetrics := model.GetNormalMetric()
	defer model.PutNormalMetric(normalMetrics)
	normalMetrics.Metric.Name = name
	normalMetrics.Metric.Aggregation = model.Aggregation_AGGREGATION_HISTOGRAM
	normalMetrics.Metric.Value = 3
	p.ProcessNormalMetric(normalMetrics)
	// 等导出。
	time.Sleep(time.Duration(cfg.Processor.WindowSeconds) * 2 * time.Second)
	// 第一次断言。
	fmt.Printf("%+v\n", exporter.normals)
	vmRangeOld := strings.VMRangeFloatToString(2) + strings.VMRangeSeparator + strings.VMRangeFloatToString(5)
	assert.True(t, vmRangeExist(exporter.normals, vmRangeOld))
	// 分桶配置热更新。
	cfg.Processor.HistogramBuckets = nil
	cfg.Processor.HistogramBuckets = append(
		cfg.Processor.HistogramBuckets,
		model.HistogramBucket{Name: name, Buckets: []float64{0, 1, 6}},
	)
	p.Watch(
		&ocp.GalileoConfig{
			Config: model.GetConfigResponse{
				MetricsConfig: model.MetricsConfig{
					Enable:    true,
					Processor: cfg.Processor,
					Exporter:  cfg.Exporter,
				},
			},
			Resource: cfg.Resource,
		},
	)
	p.UpdateConfig(cfg)
	// 等导出后，热更新生效。
	time.Sleep(time.Duration(cfg.Processor.WindowSeconds) * 2 * time.Second)
	// 导出器清空。
	exporter.normals = nil
	// 处理数据。
	normalMetrics.Metric.Value = 2
	p.ProcessNormalMetric(normalMetrics)
	normalMetrics.Metric.Value = 3
	p.ProcessNormalMetric(normalMetrics)
	// 等导出。
	time.Sleep(time.Duration(cfg.Processor.WindowSeconds) * 2 * time.Second)
	// 断言。
	assert.False(t, vmRangeExist(exporter.normals, vmRangeOld))
	vmRangeNew := strings.VMRangeFloatToString(1) + strings.VMRangeSeparator + strings.VMRangeFloatToString(6)
	assert.True(t, vmRangeExist(exporter.normals, vmRangeNew))
}

func vmRangeExist(normals []*model.NormalMetricOTP, vmrange string) bool {
	for _, normal := range normals {
		h := normal.Metric.GetHistogram()
		if h == nil {
			continue
		}
		for _, bucket := range h.Buckets {
			if bucket.Range == vmrange {
				return true
			}
		}
	}
	return false
}

// TestProcessServerMetrics_SecondLevel 测试被调监控处理（秒级监控）。
func TestProcessServerMetrics_SecondLevel(t *testing.T) {
	// 构造处理器。
	exporter := newExporter()
	cfg := newProcessorCfg()
	cfg.Processor.WindowSeconds = 20 // 正常情况 20 秒才有数据。
	now := time.Now().Unix()
	cfg.Processor.SecondGranularitys = []model.SecondGranularity{
		// 当前时间前后 60 秒，开启被调秒级监控。
		{MonitorName: model.RPCServer, BeginSecond: now - 60, EndSecond: now + 60},
	}
	processor, err := NewProcessor(cfg, exporter)
	assert.Nil(t, err)
	// 处理数据。
	serverMetrics := model.GetServerMetrics(2)
	defer model.PutServerMetrics(serverMetrics)
	serverMetrics.RpcLabels.Fields[0].Name = model.RPCLabels_callee_ip
	serverMetrics.RpcLabels.Fields[0].Value = "aa.b.c.d"
	serverMetrics.RpcLabels.Fields[1].Name = model.RPCLabels_caller_ip
	serverMetrics.RpcLabels.Fields[1].Value = "ee.f.g.h"
	serverMetrics.Metrics[model.ServerMetricStartedTotalPoint].Value = 1
	serverMetrics.Metrics[model.ServerMetricHandledTotalPoint].Value = 1
	serverMetrics.Metrics[model.ServerMetricHandledSecondsPoint].Value = 0.01
	processor.ProcessServerMetrics(serverMetrics)
	serverMetrics.Metrics[model.ServerMetricStartedTotalPoint].Value = 1
	serverMetrics.Metrics[model.ServerMetricHandledTotalPoint].Value = 1
	serverMetrics.Metrics[model.ServerMetricHandledSecondsPoint].Value = 0.1
	processor.ProcessServerMetrics(serverMetrics)
	// 等待数据导出。
	time.Sleep(time.Second + 20*time.Millisecond) // 1 秒多一些，就有数据了。
	// 断言。
	assert.Equal(t, true, exporter.count.Load() >= int64(1))
	assert.Equal(t, true, len(exporter.servers) >= 1)
	sum := float64(0.01 + 0.1)
	count := int64(2)
	ranges := []string{
		"1.000e-02...2.500e-02",
		"1.000e-01...2.500e-01",
	}
	counts := []int64{1, 1}
	otp := model.ServerMetricsOTP{
		RpcServerStartedTotal:   2,
		RpcServerHandledTotal:   2,
		RpcServerHandledSeconds: model.NewHistogram(count, sum, counts, ranges),
		RpcLabels:               &serverMetrics.RpcLabels,
	}
	assert.Equal(t, otp.String(), exporter.servers[0].String())
}

// TestProcessCustomMetric_Sample 测试自定义监控处理。（指标采样）
func TestProcessCustomMetric_Sample(t *testing.T) {
	// 构造处理器。
	exporter := newExporter()
	cfg := newProcessorCfg()
	cfg.Processor.SampleMonitors = []model.SampleMonitor{
		{MonitorName: "test_monitor", SampleType: model.MetricsSampleType_METRICS_SAMPLE_TYPE_ROWS, Fraction: 0.2},
	}
	cfg.Processor.HistogramBuckets = []model.HistogramBucket{
		{Name: "custom_histogram_test_monitor_test_histogram_metric", Buckets: []float64{0, 20, 40, 60, 80}},
	}
	processor, err := NewProcessor(cfg, exporter)
	assert.Nil(t, err)
	// 测试 10 * 10 * 100 个维度，每个维度上报一个点。测试允许误差为 0.1。(误差大小和维度以及上报数量相关，此处测试用例误差通过)
	const (
		v1count          = 10                                   // v1 维度数量
		v2count          = 10                                   // v2 维度数量
		v3count          = 100                                  // v3 维度数量
		vcount           = 10                                   // 每个维度上报点数
		total            = v1count * v2count * v3count * vcount // 上报总数
		bucketCount      = total / 5                            // 每个 bucket 的统计数量
		randMaxValue     = 100                                  // 上报随机值的最大值
		avgExpectedValue = randMaxValue / 2                     // 期望 avg 值
	)
	next := 0
	for i := 0; i < v1count; i++ {
		for j := 0; j < v2count; j++ {
			for k := 0; k < v3count; k++ {
				for l := 0; l < vcount; l++ {
					// 处理数据。
					customLabels := []model.Label{
						{Name: "k1", Value: "label1" + strconv.Itoa(i)},
						{Name: "k2", Value: "label2" + strconv.Itoa(j)},
						{Name: "k3", Value: "label3" + strconv.Itoa(k)},
					}
					v := next
					next = (next + 1) % randMaxValue
					processor.ProcessCustomMetrics(
						&model.CustomMetrics{
							Metrics: []model.Metric{
								{Name: "test_set_metric", Aggregation: model.Aggregation_AGGREGATION_SET, Value: 1},
								{Name: "test_sum_metric", Aggregation: model.Aggregation_AGGREGATION_SUM, Value: 1},
								{
									Name:        "test_avg_metric",
									Aggregation: model.Aggregation_AGGREGATION_AVG,
									Value:       float64(v),
								},
								{
									Name:        "test_max_metric",
									Aggregation: model.Aggregation_AGGREGATION_MAX,
									Value:       float64(v),
								},
								{
									Name:        "test_min_metric",
									Aggregation: model.Aggregation_AGGREGATION_MIN,
									Value:       float64(v),
								},
								{
									Name:        "test_histogram_metric",
									Aggregation: model.Aggregation_AGGREGATION_HISTOGRAM,
									Value:       float64(v),
								},
								{
									Name:        "test_counter_metric",
									Aggregation: model.Aggregation_AGGREGATION_COUNTER,
									Value:       1,
								},
							},
							CustomLabels: customLabels,
							MonitorName:  "test_monitor",
						},
					)
				}
			}
		}
	}
	// 等待数据导出。
	time.Sleep(time.Duration(cfg.Processor.WindowSeconds*2) * time.Second)
	// 聚合计算导出结果，仅测试总体值，不关心维度。
	var setValue, sumValue, avgSum, maxValue, minValue, histogramSum, counterValue float64
	var avgCount, histogramCount int64
	histogramBucketCount := make(map[string]int64, 5)
	for i := range exporter.customs {
		setValue = exporter.customs[i].Metrics[0].GetValue()
		sumValue += exporter.customs[i].Metrics[1].GetValue()
		avgSum += exporter.customs[i].Metrics[2].GetAvg().GetSum()
		avgCount += exporter.customs[i].Metrics[2].GetAvg().GetCount()
		maxValue = math.Max(maxValue, exporter.customs[i].Metrics[3].GetValue())
		minValue = math.Min(minValue, exporter.customs[i].Metrics[4].GetValue())
		histogramSum += exporter.customs[i].Metrics[5].GetHistogram().GetSum()
		histogramCount += exporter.customs[i].Metrics[5].GetHistogram().GetCount()
		buckets := exporter.customs[i].Metrics[5].GetHistogram().GetBuckets()
		for j := range buckets {
			histogramBucketCount[buckets[j].GetRange()] += buckets[j].GetCount()
		}
		counterValue += exporter.customs[i].Metrics[6].GetValue()
	}
	// 断言。
	assert.Equal(t, float64(1), setValue)
	assert.True(t, math.Abs(sumValue-total)/total < 0.1)
	assert.True(t, math.Abs(avgSum/float64(avgCount)-avgExpectedValue)/avgExpectedValue < 0.1)
	assert.True(t, math.Abs(maxValue-randMaxValue-1)/randMaxValue < 0.1)
	assert.True(t, math.Abs(minValue-0)/randMaxValue < 0.1)
	assert.True(t, math.Abs(histogramSum/float64(histogramCount)-avgExpectedValue)/avgExpectedValue < 0.1)
	for _, c := range histogramBucketCount {
		assert.True(t, math.Abs(float64(c-bucketCount))/bucketCount < 0.1)
	}
	assert.True(t, math.Abs(counterValue-total)/total < 0.1)
}

// TestProcessServerMetric_Sample 测试被调监控处理。（指标采样）
func TestProcessServerMetric_Sample(t *testing.T) {
	// 构造处理器。
	exporter := newExporter()
	cfg := newProcessorCfg()
	cfg.Processor.SampleMonitors = []model.SampleMonitor{
		{MonitorName: "rpc_server", SampleType: model.MetricsSampleType_METRICS_SAMPLE_TYPE_ROWS, Fraction: 0.2},
	}
	cfg.Processor.HistogramBuckets = []model.HistogramBucket{
		{Name: "rpc_server_handled_seconds", Buckets: []float64{0, 20, 40, 60, 80}},
	}
	processor, err := NewProcessor(cfg, exporter)
	assert.Nil(t, err)
	// 测试 10 * 10 * 100 个维度，每个维度上报一个点。测试允许误差为 0.1。(误差大小和维度以及上报数量相关，此处测试用例误差通过)
	const (
		v1count          = 10                                   // v1 维度数量
		v2count          = 10                                   // v2 维度数量
		v3count          = 100                                  // v3 维度数量
		vcount           = 100                                  // 每个维度上报点数
		total            = v1count * v2count * v3count * vcount // 上报总数
		bucketCount      = total / 5                            // 每个 bucket 的统计数量
		randMaxValue     = 100                                  // 上报随机值的最大值
		avgExpectedValue = randMaxValue / 2                     // 期望 avg 值
	)
	for i := 0; i < v1count; i++ {
		for j := 0; j < v2count; j++ {
			for k := 0; k < v3count; k++ {
				for l := 0; l < vcount; l++ {
					// 处理数据。
					v := rand.Intn(randMaxValue)
					serverMetrics := model.GetServerMetrics(3)
					serverMetrics.RpcLabels.Fields[0].Name = model.RPCLabels_caller_server
					serverMetrics.RpcLabels.Fields[0].Value = "caller_server_" + strconv.Itoa(i)
					serverMetrics.RpcLabels.Fields[1].Name = model.RPCLabels_caller_method
					serverMetrics.RpcLabels.Fields[1].Value = "caller_method_" + strconv.Itoa(j)
					serverMetrics.RpcLabels.Fields[2].Name = model.RPCLabels_caller_ip
					serverMetrics.RpcLabels.Fields[2].Value = "caller_ip_" + strconv.Itoa(k)
					serverMetrics.Metrics[model.ServerMetricStartedTotalPoint].Value = 1
					serverMetrics.Metrics[model.ServerMetricHandledTotalPoint].Value = 1
					serverMetrics.Metrics[model.ServerMetricHandledSecondsPoint].Value = float64(v)
					processor.ProcessServerMetrics(serverMetrics)
				}
			}
		}
	}
	// 等待数据导出。
	time.Sleep(time.Duration(cfg.Processor.WindowSeconds*2) * time.Second)
	// 聚合计算导出结果，仅测试总体值，不关心维度。
	var startedValue, handledValue int64
	var histogramSum float64
	var histogramCount int64
	histogramBucketCount := make(map[string]int64, 5)
	for i := range exporter.servers {
		startedValue += exporter.servers[i].GetRpcServerStartedTotal()
		handledValue += exporter.servers[i].GetRpcServerHandledTotal()
		histogramSum += exporter.servers[i].GetRpcServerHandledSeconds().GetSum()
		histogramCount += exporter.servers[i].GetRpcServerHandledSeconds().GetCount()
		buckets := exporter.servers[i].GetRpcServerHandledSeconds().GetBuckets()
		for j := range buckets {
			histogramBucketCount[buckets[j].GetRange()] += buckets[j].GetCount()
		}
	}
	// 断言。
	fmt.Println(startedValue)
	fmt.Println(handledValue)
	fmt.Println(histogramSum / float64(histogramCount))
	assert.True(t, math.Abs(float64(startedValue)-total)/total < 0.1)
	assert.True(t, math.Abs(float64(handledValue)-total)/total < 0.1)
	assert.True(t, math.Abs(histogramSum/float64(histogramCount)-avgExpectedValue)/avgExpectedValue < 0.1)
	for _, c := range histogramBucketCount {
		assert.True(t, math.Abs(float64(c-bucketCount))/bucketCount < 0.1)
	}
}

func Test_searchAggregator(t *testing.T) {
	tests := []struct {
		giveWindow time.Duration
		giveWraps  []*aggregatorWrap
		wantWindow time.Duration
	}{
		{
			giveWindow: time.Second,
			giveWraps: []*aggregatorWrap{
				{window: time.Second, aggregator: &aggregator{}},
			},
			wantWindow: time.Second,
		},
		{
			giveWindow: time.Second,
			giveWraps: []*aggregatorWrap{
				{window: time.Second, aggregator: &aggregator{}},
				{window: time.Second * 5, aggregator: &aggregator{}},
			},
			wantWindow: time.Second,
		},
		{
			giveWindow: time.Second * 5,
			giveWraps: []*aggregatorWrap{
				{window: time.Second, aggregator: &aggregator{}},
				{window: time.Second * 5, aggregator: &aggregator{}},
			},
			wantWindow: time.Second * 5,
		},
		{
			giveWindow: 0,
			giveWraps: []*aggregatorWrap{
				{window: time.Second, aggregator: &aggregator{}},
				{window: time.Second * 5, aggregator: &aggregator{}},
			},
			wantWindow: time.Second,
		},
		{
			giveWindow: time.Second * 10,
			giveWraps: []*aggregatorWrap{
				{window: time.Second, aggregator: &aggregator{}},
				{window: time.Second * 5, aggregator: &aggregator{}},
			},
			wantWindow: time.Second * 5,
		},
		{
			giveWindow: time.Second * 3,
			giveWraps: []*aggregatorWrap{
				{window: time.Second, aggregator: &aggregator{}},
				{window: time.Second * 5, aggregator: &aggregator{}},
			},
			wantWindow: time.Second,
		},
		{
			giveWindow: time.Second,
			giveWraps: []*aggregatorWrap{
				{window: time.Second, aggregator: &aggregator{}},
				{window: time.Second * 5, aggregator: &aggregator{}},
				{window: time.Second * 10, aggregator: &aggregator{}},
			},
			wantWindow: time.Second,
		},
		{
			giveWindow: time.Second * 5,
			giveWraps: []*aggregatorWrap{
				{window: time.Second, aggregator: &aggregator{}},
				{window: time.Second * 5, aggregator: &aggregator{}},
				{window: time.Second * 10, aggregator: &aggregator{}},
			},
			wantWindow: time.Second * 5,
		},
		{
			giveWindow: time.Second * 10,
			giveWraps: []*aggregatorWrap{
				{window: time.Second, aggregator: &aggregator{}},
				{window: time.Second * 5, aggregator: &aggregator{}},
				{window: time.Second * 10, aggregator: &aggregator{}},
			},
			wantWindow: time.Second * 10,
		},
		{
			giveWindow: time.Second * 9,
			giveWraps: []*aggregatorWrap{
				{window: time.Second, aggregator: &aggregator{}},
				{window: time.Second * 5, aggregator: &aggregator{}},
				{window: time.Second * 10, aggregator: &aggregator{}},
			},
			wantWindow: time.Second * 5,
		},
	}
	for i, tt := range tests {
		got := searchAggregator(tt.giveWindow, tt.giveWraps)
		require.Equalf(t, tt.wantWindow, got.window, "case=%d", i)
	}
}
