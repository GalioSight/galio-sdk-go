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

package runtimes

import (
	"runtime"

	"galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/processors/omp/metrics/point"
)

// writeGoMetrics 上报逻辑迁移自：https://github.com/VictoriaMetrics/metrics/blob/master/go_metrics.go
func writeGoMetrics(metrics *model.Metrics) {
	// 内存相关上报。
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	writeGoMemStats(metrics, &ms)
	writeGoGc(metrics, &ms)
	// 进程、线程、cgo、核数相关上报。
	metrics.AddNormalMetric("go_cgo_calls_count", model.Aggregation_AGGREGATION_SET, float64(runtime.NumCgoCall()))
	metrics.AddNormalMetric("go_cpu_count", model.Aggregation_AGGREGATION_SET, float64(runtime.NumCPU()))
	metrics.AddNormalMetric("go_gomaxprocs", model.Aggregation_AGGREGATION_SET, float64(runtime.GOMAXPROCS(0)))
	metrics.AddNormalMetric("go_goroutines", model.Aggregation_AGGREGATION_SET, float64(runtime.NumGoroutine()))
	numThread, _ := runtime.ThreadCreateProfile(nil)
	metrics.AddNormalMetric("go_threads", model.Aggregation_AGGREGATION_SET, float64(numThread))
	// 兼容伽利略页面补充上报。
	writeGcHistogram(metrics, &ms)
	metrics.AddNormalMetric("process_cpu_cores", model.Aggregation_AGGREGATION_SET, float64(runtime.NumCPU()))
	metrics.AddNormalMetric("go_max_process_num", model.Aggregation_AGGREGATION_SET, float64(runtime.GOMAXPROCS(0)))
}

// writeGoMemStats go 内存统计上报。
func writeGoMemStats(metrics *model.Metrics, ms *runtime.MemStats) {
	metrics.AddNormalMetric("go_memstats_alloc_bytes", model.Aggregation_AGGREGATION_SET, float64(ms.Alloc))
	metrics.AddNormalMetric("go_memstats_alloc_bytes_total", model.Aggregation_AGGREGATION_SET, float64(ms.TotalAlloc))
	metrics.AddNormalMetric(
		"go_memstats_buck_hash_sys_bytes", model.Aggregation_AGGREGATION_SET, float64(ms.BuckHashSys),
	)
	metrics.AddNormalMetric("go_memstats_frees_total", model.Aggregation_AGGREGATION_SET, float64(ms.Frees))
	metrics.AddNormalMetric("go_memstats_heap_alloc_bytes", model.Aggregation_AGGREGATION_SET, float64(ms.HeapAlloc))
	metrics.AddNormalMetric("go_memstats_heap_idle_bytes", model.Aggregation_AGGREGATION_SET, float64(ms.HeapIdle))
	metrics.AddNormalMetric("go_memstats_heap_inuse_bytes", model.Aggregation_AGGREGATION_SET, float64(ms.HeapInuse))
	metrics.AddNormalMetric("go_memstats_heap_objects", model.Aggregation_AGGREGATION_SET, float64(ms.HeapObjects))
	metrics.AddNormalMetric(
		"go_memstats_heap_released_bytes", model.Aggregation_AGGREGATION_SET, float64(ms.HeapReleased),
	)
	metrics.AddNormalMetric("go_memstats_heap_sys_bytes", model.Aggregation_AGGREGATION_SET, float64(ms.HeapSys))
	metrics.AddNormalMetric("go_memstats_lookups_total", model.Aggregation_AGGREGATION_SET, float64(ms.Lookups))
	metrics.AddNormalMetric("go_memstats_mallocs_total", model.Aggregation_AGGREGATION_SET, float64(ms.Mallocs))
	metrics.AddNormalMetric(
		"go_memstats_mcache_inuse_bytes", model.Aggregation_AGGREGATION_SET, float64(ms.MCacheInuse),
	)
	metrics.AddNormalMetric("go_memstats_mcache_sys_bytes", model.Aggregation_AGGREGATION_SET, float64(ms.MCacheSys))
	metrics.AddNormalMetric("go_memstats_mspan_inuse_bytes", model.Aggregation_AGGREGATION_SET, float64(ms.MSpanInuse))
	metrics.AddNormalMetric("go_memstats_mspan_sys_bytes", model.Aggregation_AGGREGATION_SET, float64(ms.MSpanSys))
	metrics.AddNormalMetric("go_memstats_other_sys_bytes", model.Aggregation_AGGREGATION_SET, float64(ms.OtherSys))
	metrics.AddNormalMetric("go_memstats_stack_inuse_bytes", model.Aggregation_AGGREGATION_SET, float64(ms.StackInuse))
	metrics.AddNormalMetric("go_memstats_stack_sys_bytes", model.Aggregation_AGGREGATION_SET, float64(ms.StackSys))
	metrics.AddNormalMetric("go_memstats_sys_bytes", model.Aggregation_AGGREGATION_SET, float64(ms.Sys))
}

// writeGoGc 上报 go gc 统计。
func writeGoGc(metrics *model.Metrics, ms *runtime.MemStats) {
	metrics.AddNormalMetric("go_memstats_gc_cpu_fraction", model.Aggregation_AGGREGATION_SET, ms.GCCPUFraction)
	metrics.AddNormalMetric("go_memstats_gc_sys_bytes", model.Aggregation_AGGREGATION_SET, float64(ms.GCSys))
	metrics.AddNormalMetric(
		"go_memstats_last_gc_time_seconds", model.Aggregation_AGGREGATION_SET, float64(ms.LastGC)/1e9,
	)
	metrics.AddNormalMetric("go_memstats_next_gc_bytes", model.Aggregation_AGGREGATION_SET, float64(ms.NextGC))

	metrics.AddNormalMetric(
		"go_gc_duration_seconds_sum", model.Aggregation_AGGREGATION_SET, float64(ms.PauseTotalNs)/1e9,
	)
	metrics.AddNormalMetric("go_gc_duration_seconds_count", model.Aggregation_AGGREGATION_SET, float64(ms.NumGC))
	metrics.AddNormalMetric("go_gc_forced_count", model.Aggregation_AGGREGATION_SET, float64(ms.NumForcedGC))
}

const (
	// Galileo Runtime 监控项名字
	galileoRuntimeMonitorName = "GalileoRuntime"
	// 服务被调处理请求量 qps 的指标名
	rpcServerHandledQPSMetricName = "rpc_server_handled_qps"
)

// writeServerQPS 上报 galileo 服务被调接收请求量 qps，每秒上报一次
func writeServerQPS(metrics *model.Metrics) {
	// set 聚合类型，代表单机 QPS
	metricName := model.CustomName(
		galileoRuntimeMonitorName, rpcServerHandledQPSMetricName,
		model.Aggregation_AGGREGATION_SET,
	)
	metrics.AddCustomMetric(
		galileoRuntimeMonitorName, metricName, model.Aggregation_AGGREGATION_SET,
		float64(loadAndResetHandledTotal()),
	)
}

var (
	// lastGcNum 上次 gc num
	lastGcNum = -1
	// gcBucket 默认 gc 耗时分桶。
	gcBucket = configs.NewBucket(
		[]float64{
			0,
			0.0001, // 100 us
			0.0003, // 300 us
			0.0005, // 500 us
			0.0007, // 700 us
			0.001,  // 1 ms
			0.003,  // 3 ms
			0.005,  // 5 ms
			0.007,  // 7 ms
			0.01,   // 10 ms
			0.03,   // 30 ms
			0.05,   // 50 ms
			0.07,   // 70 ms
			0.1,    // 100 ms
			0.3,    // 300 ms
			0.5,    // 500 ms
			0.7,    // 700 ms
			1,      // 1 s
		},
	)
	// gcHistogram gc 耗时 point。
	// 分桶参考：
	// 注：某些服务对 gc 比较敏感，所以细化了些，一个容器多几根时间线不会引起数据膨胀。
	gcHistogram = newGCHistogram()
)

func newGCHistogram() *point.Point {
	p := point.Get(
		model.Aggregation_AGGREGATION_HISTOGRAM, "go_gc_pause_seconds",
	)
	p.SetBucket(func() *configs.Bucket { return gcBucket })
	return p
}

// writeGcHistogram 上报 gc 耗时 histogram。
// 区别于 writeGoGc，精准上报每一次 gc 的情况。
// 注：如果一个服务的 go gc 特别频繁，存在漏报数据的情况。如：writeGcHistogram 20 秒执行一次，期间 gc 超过 256 次，会发生漏报。
func writeGcHistogram(metrics *model.Metrics, ms *runtime.MemStats) {
	for i := lastGcNum + 1; i <= int(ms.NumGC); i++ {
		pauseNs := ms.PauseNs[(i+255)%256]
		pauseS := float64(pauseNs) / 1e9 // ns -> s
		gcHistogram.Update(pauseS)
	}
	lastGcNum = int(ms.NumGC)

	otp := model.NewNormalMetricsOTP()
	// 当且仅当 p.toOTPFunc 为 nil 时会返回 err，此处不可能有 err。
	cnt, _ := gcHistogram.ToOTP(otp, 0)
	// 如果此时间段没有 gc，就没有数据，会命中这个分支。
	// 此时调用方需要判断，避免上报空数据。
	// 否则会导致后台收到一个指标名为空的数据，带来干扰。
	// 请参阅 histogramToOTP 的实现。
	if cnt == 0 {
		return
	}
	metrics.NormalMetrics = append(metrics.NormalMetrics, otp)
}
