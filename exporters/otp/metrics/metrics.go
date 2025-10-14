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

// Package metrics : otp 协议 metrics 数据导出模块
//
// 1. 接收到的数据，先进行分页，避免数据太大。
// 2. 分页放入 chan 队列之中，用于异步发送。
// 3. 启动若干异步线程，持续从 chan 中消费数据，用于发送。
// 4. 发送数据使用均匀限流器限制发送速率，避免后端服务瞬间收到大量的包，导致尖刺。
// 5. 限流器的速率根据当前队列长度进行动态调整，预期能在时间窗口内发送完当前队列中的所有的数据。
// 6. 数据是 pb 格式，先序列化，然后使用 snappy 压缩。
// 7. 数据序列化及压缩过程中，需要频繁用到 []byte，为减少 gc，进行对象重用。每个线程使用自己的对象。
// 7. 通过 HTTP post 方式发送数据到 otp 后端服务。
package metrics

import (
	"galiosight.ai/galio-sdk-go/components"
	"galiosight.ai/galio-sdk-go/configs"
	otphttp "galiosight.ai/galio-sdk-go/exporters/otp/http"
	"galiosight.ai/galio-sdk-go/lib/file"
	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
)

// metricsExporter 监控导出器实现。
type metricsExporter struct {
	httpExporter otphttp.HTTPExporter
	cfg          *configs.Metrics
	out          chan pageMetrics
	log          *logs.Wrapper
	stats        *model.SelfMonitorStats
	fileExporter *file.Exporter
}

// UpdateConfig 更新配置。
func (m *metricsExporter) UpdateConfig(cfg *configs.Metrics) {
	m.httpExporter.UpdateConfig(
		cfg.Exporter.MaxRetryCount,
		cfg.Exporter.Collector,
	)
}

// GetStats 返回自监控统计数据。
// 注意返回的是指针，调用方只能读取结果，不能修改它。
func (m *metricsExporter) GetStats() *model.SelfMonitorStats {
	return m.stats
}

// NewExporter 根据配置创建导出器。
func NewExporter(cfg *configs.Metrics) (components.MetricsExporter, error) {
	cfg = mergeDefaultCfg(cfg)
	exporter := &metricsExporter{
		cfg: cfg,
		httpExporter: otphttp.NewHTTPGeneralExporter(
			int(cfg.Exporter.TimeoutMs), cfg.Exporter.Collector.Addr, cfg.Log,
			otphttp.WithDirectIPPorts(cfg.Exporter.Collector.DirectIpPort),
			otphttp.WithHeaders(
				map[string]string{
					model.TenantHeaderKey:    cfg.Resource.TenantId,
					model.TargetHeaderKey:    cfg.Resource.Target,
					model.SchemaURLHeaderKey: cfg.SchemaURL,
					model.APIKeyHeaderKey:    cfg.APIKey,
				},
			),
			otphttp.WithMaxRetryCount(cfg.Exporter.MaxRetryCount),
		),
		out:          make(chan pageMetrics, cfg.Exporter.BufferSize),
		log:          cfg.Log,
		stats:        cfg.Stats,
		fileExporter: file.NewExporter(cfg.Exporter.ExportToFile, "galileo/metrics", cfg.Log),
	}
	exporter.run()
	return exporter, nil
}

// mergeDefaultCfg 如果 cfg 参数未配置，使用默认配置填充。
func mergeDefaultCfg(cfg *configs.Metrics) *configs.Metrics {
	if cfg.Stats == nil {
		cfg.Stats = &model.SelfMonitorStats{}
	}
	if cfg.Exporter.ThreadCount <= 0 {
		cfg.Exporter.ThreadCount = 10
	}
	if cfg.Exporter.BufferSize <= 0 {
		cfg.Exporter.BufferSize = 1000 * 1000
	}
	if cfg.Exporter.PageSize <= 0 {
		cfg.Exporter.PageSize = 1000
	}
	if cfg.Exporter.WindowSeconds < 1 {
		cfg.Exporter.WindowSeconds = 1
	}
	if cfg.Exporter.TimeoutMs <= 0 {
		cfg.Exporter.TimeoutMs = 1000
	}
	return cfg
}

// run 会启动若干个线程，执行数据上报任务。
// 因为数据上报量会非常大，每个数据包创建一个线程的话，会导致大量线程，gc 负担很大，可能会影响性能。
// 所以此处线程数量是固定的。
// 一般是 1 个分页协程，10 个发送协程。
// 启动一个线程，每个时间窗口更新所需的 tps，同时调整自己的上报速率。
func (m *metricsExporter) run() {
	for i := 0; i < int(m.cfg.Exporter.ThreadCount); i++ {
		go m.worker(i)
	}
}

// worker 将 chan 中的数据，按照一定的速率上报给后端。
//
// 通常 SDK 数据是会进行聚合，到一个时间窗口再进行上报。
// 如果数据量非常大，短时间大量上报，可能会导致后端压力太大。
// 所以需要进行速率控制。
// 一般情况下，上报数据是不能随便丢弃的，所以速率并不是一个硬性指标。
// 只是为了让后端的负载相对平滑，不要有太多的尖刺。
// 否则后端负载尖刺时，可能会出现内存不足，连接失败等各种问题。
// 所以限流器的速率需要根据业务数据量，进行自适应调整。
// 过高过低都是不好的。
// 自适应调整速率的方式是，定时统计上报的平均 QPS，然后调整 limiter 的参数。
// 上报时，会更新三个自监控指标。
func (m *metricsExporter) worker(i int) {
	r := otphttp.NewReuseObject()
	for page := range m.out {
		err := m.httpExporter.Export(page.metrics, r)
		if m.cfg.Exporter.ExportToFile {
			m.fileExporter.Export(page.metrics)
		}
		// model.PutMetrics(page.metrics) // TODO(jaimeyang) 内存优化，待重构。
		m.stats.MetricsStats.ReportHandledTotal.Inc()
		m.stats.MetricsStats.ReportHandledRowsTotal.Add(int64(page.size))
		if err != nil {
			m.log.Errorf("[galileo]metricsExporter.worker|err=%v\n", err)
			m.stats.MetricsStats.ReportErrorTotal.Inc()
			m.stats.MetricsStats.ReportErrorRowsTotal.Add(int64(page.size))
		}
	}
}

// Export 将数据放到 chan 中，然后通过多个 worker 并发进行上报。
// 该函数是并发安全的。
// 先进行分页，再放到队列中。目的是控制单包大小，方便数据平滑，避免单包过大导致的发送超时。
// 如果数据量过多，chan 满的话，会阻塞住。
func (m *metricsExporter) Export(metrics *model.Metrics) {
	total := len(metrics.ClientMetrics) + len(metrics.ServerMetrics) + len(metrics.NormalMetrics) +
		len(metrics.CustomMetrics)
	pageSize := int(m.cfg.Exporter.PageSize)
	for processed := 0; processed < total; {
		page, pageCount := nextPage(metrics, pageSize)
		processed += pageCount
		m.out <- pageMetrics{metrics: page, size: pageCount}
	}
}

func nextPage(metrics *model.Metrics, pageSize int) (*model.Metrics, int) {
	page := &model.Metrics{
		TimestampMs:  metrics.TimestampMs,
		NormalLabels: metrics.NormalLabels,
	}
	pageCount := 0
	if c := min(len(metrics.ClientMetrics), pageSize-pageCount); c != 0 {
		page.ClientMetrics = make([]*model.ClientMetricsOTP, 0, c)
		page.ClientMetrics = append(page.ClientMetrics, metrics.ClientMetrics[0:c]...)
		metrics.ClientMetrics = metrics.ClientMetrics[c:len(metrics.ClientMetrics)]
		pageCount += c
	}
	if c := min(len(metrics.ServerMetrics), pageSize-pageCount); c != 0 {
		page.ServerMetrics = make([]*model.ServerMetricsOTP, 0, c)
		page.ServerMetrics = append(page.ServerMetrics, metrics.ServerMetrics[0:c]...)
		metrics.ServerMetrics = metrics.ServerMetrics[c:len(metrics.ServerMetrics)]
		pageCount += c
	}
	if c := min(len(metrics.NormalMetrics), pageSize-pageCount); c != 0 {
		page.NormalMetrics = make([]*model.NormalMetricOTP, 0, c)
		page.NormalMetrics = append(page.NormalMetrics, metrics.NormalMetrics[0:c]...)
		metrics.NormalMetrics = metrics.NormalMetrics[c:len(metrics.NormalMetrics)]
		pageCount += c
	}
	if c := min(len(metrics.CustomMetrics), pageSize-pageCount); c != 0 {
		page.CustomMetrics = make([]*model.CustomMetricsOTP, 0, c)
		page.CustomMetrics = append(page.CustomMetrics, metrics.CustomMetrics[0:c]...)
		metrics.CustomMetrics = metrics.CustomMetrics[c:len(metrics.CustomMetrics)]
		pageCount += c
	}
	return page, pageCount
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
