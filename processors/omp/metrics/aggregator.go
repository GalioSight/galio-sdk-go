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
	"sync"
	"time"

	"galiosight.ai/galio-sdk-go/components"
	"galiosight.ai/galio-sdk-go/lib/timedmax"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/processors/omp/metrics/point"
	uberatomic "go.uber.org/atomic"
)

// TODO(jaimeyang) 重构后，最佳性能配置待重新测试。
const (
	shardCount = 11 // 数据分 11 片。
)

// shard 数据分片。
type shard struct {
	multis map[PK]*multi
	mu     sync.RWMutex // 保护 multis 并发读写。
}

func newShard() *shard {
	return &shard{
		multis: make(map[PK]*multi),
	}
}

type buffer struct {
	shards     [model.MaxGroup][shardCount]*shard
	multiCount uberatomic.Int64 // 当前 buffer 的多值点个数。
	pointCount uberatomic.Int64 // 当前 buffer 的单值点个数。
}

func newBuffer() *buffer {
	b := &buffer{}
	for groupID := range b.shards {
		for shardID := range b.shards[groupID] {
			b.shards[groupID][shardID] = newShard()
		}
	}
	return b
}

// aggregator 双 buffer 聚合器。
type aggregator struct {
	normalLabels           *model.NormalLabels                // 属性标签，启动后唯一确定。
	windowFunc             func() time.Duration               // 窗口大小配置，函数化方便热更新。
	bucketFunc             func(name string) point.BucketFunc // 分桶配置。
	overloadProtectionFunc func() bool                        // 是否过载。
	exporter               components.MetricsExporter         // 导出器。
	stats                  *model.SelfMonitorStats            // 自监控统计。
	writer                 *buffer
	reader                 *buffer
	mu                     sync.Mutex // 保护 writer 的并发安全。
	sampler                *sampler
}

func newAggregator(
	normalLabels *model.NormalLabels,
	windowFunc func() time.Duration,
	bucketFunc func(name string) point.BucketFunc,
	overloadProtectionFunc func() bool,
	stats *model.SelfMonitorStats,
	exporter components.MetricsExporter,
	sampler *sampler,
) *aggregator {
	a := &aggregator{
		normalLabels:           normalLabels,
		windowFunc:             windowFunc,
		bucketFunc:             bucketFunc,
		overloadProtectionFunc: overloadProtectionFunc,
		stats:                  stats,
		exporter:               exporter,
		sampler:                sampler,
	}
	a.setWriter(newBuffer())
	a.setReader(newBuffer())
	go a.swapBuffer() // 开启读写 buffer 切换。
	return a
}

// swapBuffer 读写 buffer 切换。
func (a *aggregator) swapBuffer() {
	window := a.windowFunc()
	ticker := time.NewTicker(window)
	defer ticker.Stop()
	for t := range ticker.C {
		reader := a.bufferChangeAndGetReader() // 读写 buffer 切换。
		a.flushBuffer(t, reader)               // 老数据刷新落库，必须同步，不允许异步。
		window = a.resetTicker(window, ticker) // 窗口大小热更新处理。
	}
}

// bufferChangeAndGetReader 读写 buffer 切换，且返回准备要读的 *buffer。
func (a *aggregator) bufferChangeAndGetReader() *buffer {
	a.mu.Lock() // 拿到这把锁，一定没有人在写。
	defer a.mu.Unlock()
	writer := a.getWriter()
	reader := a.getReader()
	a.setWriter(reader)
	a.setReader(writer)
	return writer // 之前用于写的，现在用于读，这个 writer 已经确定没有人在写。
}

func (a *aggregator) getWriter() *buffer {
	return a.writer
}

func (a *aggregator) setWriter(b *buffer) {
	a.writer = b
}

func (a *aggregator) getReader() *buffer {
	return a.reader
}

func (a *aggregator) setReader(b *buffer) {
	a.reader = b
}

// maxPointCount 按 1 分钟滑动窗口，统计最大的点数。
var maxPointCount = timedmax.NewMaxValueTracker(time.Minute, 0)

// flushBuffer 落库 1 个 buffer。
func (a *aggregator) flushBuffer(begin time.Time, buffer *buffer) {
	if time.Since(begin) >= time.Second {
		a.stats.DoubleBufferChangeSlow.Inc()
	}
	metrics := &model.Metrics{
		TimestampMs:  time.Now().UnixMilli(),
		NormalLabels: a.normalLabels,
	}
	exportCount := 0
	for groupID := range buffer.shards {
		for shardID := range buffer.shards[groupID] {
			s := buffer.shards[groupID][shardID]
			exportCount += a.toOTP(s, metrics)
		}
	}
	multiCount := buffer.multiCount.Swap(0)
	pointCount := buffer.pointCount.Swap(0)
	a.stats.MultiCount.Add(-multiCount)
	a.stats.MaxPointCount.Store(int64(maxPointCount.Update(float64(a.stats.PointCount.Load()))))
	a.stats.PointCount.Add(-pointCount)
	a.stats.ExportCount.Add(int64(exportCount))
	a.exporter.Export(metrics)
}

// toOTP 对指标数据进行采样和放大复原，并导出到 OTP 结构中。
// 返回值：最终导出数（注：5 个分桶算导出 5 个点）。
func (a *aggregator) toOTP(s *shard, metrics *model.Metrics) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	exportCount := 0
	for pk, m := range s.multis {
		if factor, ok := a.sampler.sample(m.monitorName, pk.labelsHashCode); ok {
			// 修改数据和导出到 otp 结构需要在同一个锁周期完成。
			m.mu.Lock()
			if factor != 1 { // 采样后数据变少，需要对采样数据根据放大系数进行修改，使采样后的数据贴合原始数据。
				m.change(factor) // 需要和 toOTPFunc 结合使用，change 之后应立即导出，否则数据不正确。
			}
			exportCount += m.toOTPFunc(m, metrics)
			m.mu.Unlock()
		}
		delete(s.multis, pk) // 删除。
		putMulti(m)
	}
	return exportCount
}

// aggregate 聚合入口。
func (a *aggregator) aggregate(
	pk *PK,
	rpcLabels *model.RPCLabels,
	customLabels []model.Label,
	monitorName string,
	extractor model.OMPMetric,
) {
	pk.set(extractor)

	a.mu.Lock() // 拿到这把锁，一定没有人在读。
	defer a.mu.Unlock()
	b := a.getWriter()
	s := b.shards[pk.group][pk.labelsHashCode%shardCount]
	if m := a.getOrNewMulti(b, s, pk, extractor, rpcLabels, customLabels, monitorName); m != nil { // shard 线程安全
		m.update(extractor) // multi 线程安全
	}
}

// getOrNewMulti 获取 or 创建一个 *multi。
func (a *aggregator) getOrNewMulti(
	b *buffer,
	s *shard,
	pk *PK,
	extractor model.OMPMetric,
	rpcLabels *model.RPCLabels,
	customLabels []model.Label,
	monitorName string,
) *multi {
	s.mu.RLock()
	m, ok := s.multis[*pk]
	s.mu.RUnlock()
	if ok { // 存在，返回。
		return m
	}

	if a.overloadProtectionFunc() { // 过载保护，过载后不允许新线创建。
		a.stats.DiscardMultiCount.Inc()
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok = s.multis[*pk]
	if ok { // 存在，返回，有并发创建。
		return m
	}
	m = getMulti(pk)
	m.setPoints(extractor, a.bucketFunc)
	m.setRPCLabels(rpcLabels)
	m.setCustomLabels(customLabels)
	m.toOTPFunc = getMultiToOTPFunc(extractor)
	m.monitorName = monitorName
	pk.copyTo(&m.pk)
	s.multis[m.pk] = m
	b.multiCount.Inc()
	b.pointCount.Add(int64(m.pk.pointCount))
	a.stats.MultiCount.Inc()
	a.stats.PointCount.Add(int64(m.pk.pointCount))
	return m
}
