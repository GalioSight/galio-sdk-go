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

// Package metrics omp 协议监控处理器实现。
// omp 协议定义：https://galiosight.ai/eco/tree/master/proto
// omp 协议设计：
// 整体流程：监控数据处理（ProcessXXX 函数）→ shard 存储 → 定时调用 exporter 导出（Export 函数）。
package metrics

import (
	"sort"
	"sync/atomic"
	"time"

	"galiosight.ai/galio-sdk-go/components"
	"galiosight.ai/galio-sdk-go/configs"
	config "galiosight.ai/galio-sdk-go/configs/metrics"
	"galiosight.ai/galio-sdk-go/configs/ocp"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/processors/omp/metrics/point"
	"galiosight.ai/galio-sdk-go/processors/omp/metrics/runtimes"
)

// aggregatorWrap 聚合器包装。
type aggregatorWrap struct {
	window     time.Duration // 聚合窗口。
	aggregator *aggregator   // 聚合器。
}

// processor 监控处理器实现。
type processor struct {
	cfg          *configs.Metrics           // 监控处理器配置。
	exporter     components.MetricsExporter // 导出器。
	stats        *model.SelfMonitorStats    // 自监控统计。
	normalLabels *model.NormalLabels        // 容器唯一属性标签，生命周期内不变。
	// 正常 10s or 20s 聚合，伽利略允许某个监控项单独配置秒级监控，因此这里有两个聚合器。
	aggregator   *aggregator       // 双 bufer 聚合器（常规）。
	aggregator1s []*aggregatorWrap // 双 bufer 聚合器（秒级，聚合窗口列表：1s，5s，10s）。
	sampler      *sampler          // 采样器。
}

// Watch 更新配置。注意 Resource 是初始化时就确定了的，后面不再发生变化。
func (p *processor) Watch(readOnlyConfig *ocp.GalileoConfig) {
	p.UpdateConfig(
		config.NewConfig(
			&p.cfg.Resource,
			config.WithProcessor(&readOnlyConfig.Config.MetricsConfig.Processor),
			config.WithExporter(&readOnlyConfig.Config.MetricsConfig.Exporter),
		),
	)
}

func newProcessor(cfg *configs.Metrics, exporter components.MetricsExporter) *processor {
	return &processor{
		cfg:          cfg,
		exporter:     exporter,
		stats:        cfg.Stats,
		normalLabels: model.ResourceToLabels(&cfg.Resource),
		sampler:      newSampler(cfg.Processor.SampleMonitors),
	}
}

// NewProcessor 构造 omp 协议监控处理器。
func NewProcessor(cfg *configs.Metrics, exporter components.MetricsExporter) (components.MetricsProcessor, error) {
	fixConfig(cfg)
	p := newProcessor(cfg, exporter)
	bucketFunc := func(name string) point.BucketFunc {
		return p.getBucket(name)
	}
	overloadProtectionFunc := func() bool {
		return p.stats.PointCount.Load() >= p.cfg.Processor.PointLimit
	}
	windowFunc := func() time.Duration {
		return time.Duration(atomic.LoadInt32(&p.cfg.Processor.WindowSeconds)) * time.Second
	}
	p.aggregator = newAggregator(
		p.normalLabels, windowFunc, bucketFunc, overloadProtectionFunc, p.stats, exporter,
		p.sampler,
	)
	p.aggregator1s = newAggregatorWraps(
		[]time.Duration{time.Second, time.Second * 5, time.Second * 10}, // 预留窗口 1s 5s 10s
		p.normalLabels, bucketFunc, overloadProtectionFunc, p.stats, exporter, p.sampler,
	)
	go p.reportRuntimes()        // 上报运行时监控。
	go p.reportGalileoRuntimes() // 上报 galileo runtime 监控
	ocp.AddWatcher(cfg.Resource.Target, p)
	return p, nil
}

func newAggregatorWraps(
	windows []time.Duration,
	normalLabels *model.NormalLabels,
	bucketFunc func(name string) point.BucketFunc,
	overloadProtectionFunc func() bool,
	stats *model.SelfMonitorStats,
	exporter components.MetricsExporter,
	sampler *sampler,
) []*aggregatorWrap {
	sort.Slice(
		windows, func(i, j int) bool {
			return windows[i] < windows[j]
		},
	)
	aggregatorWraps := make([]*aggregatorWrap, len(windows))
	for i := range windows {
		wrap := &aggregatorWrap{window: windows[i]}
		wrap.aggregator = newAggregator(
			normalLabels,
			func() time.Duration { return wrap.window },
			bucketFunc,
			overloadProtectionFunc,
			stats,
			exporter,
			sampler,
		)
		aggregatorWraps[i] = wrap
	}
	return aggregatorWraps
}

// ProcessClientMetrics 处理主调监控。
func (p *processor) ProcessClientMetrics(c *model.ClientMetrics) {
	p.cfg.GetIgnoreLabels().IgnoreClientLabels(c)
	pk := getPK()
	defer putPK(pk)
	p.getAggregator(model.ClientGroup, model.RPCClient).aggregate(pk, &c.RpcLabels, nil, model.RPCClient, c)
}

// ProcessServerMetrics 处理被调监控。
func (p *processor) ProcessServerMetrics(s *model.ServerMetrics) {
	runtimes.AddRpcServerHandledTotal()
	p.cfg.GetIgnoreLabels().IgnoreServerLabels(s)
	pk := getPK()
	defer putPK(pk)
	p.getAggregator(model.ServerGroup, model.RPCServer).aggregate(pk, &s.RpcLabels, nil, model.RPCServer, s)
}

// ProcessNormalMetric 处理属性监控。
func (p *processor) ProcessNormalMetric(n *model.NormalMetric) {
	pk := getPK()
	defer putPK(pk)
	p.getAggregator(model.NormalGroup, model.NormalProperty).aggregate(pk, nil, nil, model.NormalProperty, n)
}

// ProcessCustomMetrics 处理自定义监控。
// 会进行维度过载保护，根据配置忽略掉部分 label。
// 会对中文指标和维度进行编码，转成英文，因为后端不支持中文指标。
// 会将指标值按聚合策略存储到对应的 hash 桶中，定时上报。
// 会收集指标的元数据，上报到 OCP 服务，用于数据管理。
func (p *processor) ProcessCustomMetrics(c *model.CustomMetrics) {
	if c.MonitorName == "" {
		c.MonitorName = "default"
	}
	p.cfg.GetIgnoreLabels().IgnoreCustomLabels(c)
	if p.cfg.ConvertName {
		convertName(c)
	}
	p.processCustomMetrics(c)
}

func (p *processor) processCustomMetrics(c *model.CustomMetrics) {
	pk := getPK()
	defer putPK(pk)
	p.getAggregator(model.CustomGroup, c.MonitorName).aggregate(pk, nil, c.CustomLabels, c.MonitorName, c)
}

// GetStats 获取统计数据。
func (p *processor) GetStats() *model.SelfMonitorStats {
	return p.stats
}

func (p *processor) getAggregator(group model.MetricGroup, monitor string) *aggregator {
	ok, window := p.cfg.SecondGranularitys.Enabled(group, monitor)
	if ok { // 有秒级聚合配置。
		wrap := searchAggregator(window, p.aggregator1s) // 根据配置，搜索合适的聚合器。
		return wrap.aggregator
	}
	return p.aggregator
}

// searchAggregator 根据窗口大小，搜索合适的聚合器。
func searchAggregator(window time.Duration, wraps []*aggregatorWrap) *aggregatorWrap {
	lb := lowerBound(wraps, window)
	count := len(wraps)
	if lb >= count { // 比最大的还大：1s-5s-10s，输入 11s，得到 10s。
		return wraps[len(wraps)-1]
	}
	if wraps[lb].window == window { // 刚刚好命中：1s-5s-10s，输入 5s，得到 5s。
		return wraps[lb]
	}
	lb--
	if lb < 0 { // 比最小的还小：1s-5s-10s，输入 0s，得到 1s。
		return wraps[0]
	}
	return wraps[lb] // 中间值：1s-5s-10s，输入 3s，得到 1s。
}

func lowerBound(array []*aggregatorWrap, target time.Duration) int {
	low, high, mid := 0, len(array)-1, 0
	for low <= high {
		mid = (low + high) / 2
		if array[mid].window >= target {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	return low
}
