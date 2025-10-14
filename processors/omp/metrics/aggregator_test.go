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
	"math/rand"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"galiosight.ai/galio-sdk-go/components"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/processors/omp/metrics/point"
)

func newTestProcessor(sampleMonitors []model.SampleMonitor) *processor {
	exporter := newExporter()
	cfg := newProcessorCfg()
	cfg.Processor.SampleMonitors = sampleMonitors
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
	return p
}

// Benchmark_aggregator_flushBuffer 导出时未开启采样测试，导出 10w 维度。
// goos: darwin
// goarch: amd64
// pkg: galiosight.ai/galio-sdk-go/processors/omp/metrics
// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
// Benchmark_aggregator_flushBuffer-12                        10000              8646 ns/op            2197 B/op         51 allocs/op
// Benchmark_aggregator_flushBuffer-12                        10000              8527 ns/op            2197 B/op         51 allocs/op
// Benchmark_aggregator_flushBuffer-12                        10000              8538 ns/op            2190 B/op         51 allocs/op
// Benchmark_aggregator_flushBuffer-12                        10000              8897 ns/op            2189 B/op         51 allocs/op
// Benchmark_aggregator_flushBuffer-12                        10000              8730 ns/op            2194 B/op         51 allocs/op
func Benchmark_aggregator_flushBuffer(b *testing.B) {
	p := newTestProcessor([]model.SampleMonitor{})
	for i := 0; i < 100000; i++ {
		ompTestIDMetricsReport(p)
	}
	reader := p.aggregator.bufferChangeAndGetReader() // 读写 buffer 切换。
	now := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.aggregator.flushBuffer(now, reader)
	}
}

// Benchmark_aggregator_flushBuffer_sample_10 导出时开启采样测试，导出 10w 维度，采样率 0.1。
// goos: darwin
// goarch: amd64
// pkg: galiosight.ai/galio-sdk-go/processors/omp/metrics
// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
// Benchmark_aggregator_flushBuffer_sample_10-12              10000              4755 ns/op             497 B/op          5 allocs/op
// Benchmark_aggregator_flushBuffer_sample_10-12              10000              4665 ns/op             500 B/op          5 allocs/op
// Benchmark_aggregator_flushBuffer_sample_10-12              10000              4728 ns/op             495 B/op          5 allocs/op
// Benchmark_aggregator_flushBuffer_sample_10-12              10000              4776 ns/op             497 B/op          5 allocs/op
// Benchmark_aggregator_flushBuffer_sample_10-12              10000              4704 ns/op             497 B/op          5 allocs/op
func Benchmark_aggregator_flushBuffer_sample_10(b *testing.B) {
	p := newTestProcessor(
		[]model.SampleMonitor{
			{MonitorName: "default", SampleType: model.MetricsSampleType_METRICS_SAMPLE_TYPE_ROWS, Fraction: 0.1},
		},
	)
	for i := 0; i < 100000; i++ {
		ompTestIDMetricsReport(p)
	}
	reader := p.aggregator.bufferChangeAndGetReader() // 读写 buffer 切换。
	now := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.aggregator.flushBuffer(now, reader)
	}
}

// Benchmark_aggregator_flushBuffer_sample_50 导出时开启采样测试，导出 10w 维度，采样率 0.5。
// goos: darwin
// goarch: amd64
// pkg: galiosight.ai/galio-sdk-go/processors/omp/metrics
// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
// Benchmark_aggregator_flushBuffer_sample_50-12              10000              6594 ns/op            1255 B/op         26 allocs/op
// Benchmark_aggregator_flushBuffer_sample_50-12              10000              6346 ns/op            1254 B/op         26 allocs/op
// Benchmark_aggregator_flushBuffer_sample_50-12              10000              6389 ns/op            1258 B/op         26 allocs/op
// Benchmark_aggregator_flushBuffer_sample_50-12              10000              6500 ns/op            1248 B/op         26 allocs/op
// Benchmark_aggregator_flushBuffer_sample_50-12              10000              7017 ns/op            1255 B/op         26 allocs/op
func Benchmark_aggregator_flushBuffer_sample_50(b *testing.B) {
	p := newTestProcessor(
		[]model.SampleMonitor{
			{MonitorName: "default", SampleType: model.MetricsSampleType_METRICS_SAMPLE_TYPE_ROWS, Fraction: 0.5},
		},
	)
	for i := 0; i < 100000; i++ {
		ompTestIDMetricsReport(p)
	}
	reader := p.aggregator.bufferChangeAndGetReader() // 读写 buffer 切换。
	now := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.aggregator.flushBuffer(now, reader)
	}
}

func ompTestIDMetricsReport(ompProcessor components.MetricsProcessor) {
	c := model.GetCustomMetrics(3, 1)
	defer model.PutCustomMetrics(c)
	c.Metrics[0].Name = "test_id"
	c.Metrics[0].Aggregation = model.Aggregation_AGGREGATION_COUNTER
	c.Metrics[0].Value = 1
	c.CustomLabels[0].Name = "test_v1"
	c.CustomLabels[0].Value = "test.v1." + strconv.Itoa(rand.Intn(100))
	c.CustomLabels[1].Name = "test_v2"
	c.CustomLabels[1].Value = "test.v2." + strconv.Itoa(rand.Intn(100))
	c.CustomLabels[2].Name = "test_v3"
	c.CustomLabels[2].Value = "test.v3." + strconv.Itoa(rand.Intn(10))
	ompProcessor.ProcessCustomMetrics(c)
}
