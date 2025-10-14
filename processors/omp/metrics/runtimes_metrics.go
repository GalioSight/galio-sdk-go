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
	"time"

	"galiosight.ai/galio-sdk-go/lib/times"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/processors/omp/metrics/runtimes"
)

// reportRuntimes 定时上报运行时监控。
func (p *processor) reportRuntimes() {
	if !p.cfg.Processor.EnableProcessMetrics {
		return
	}
	windowSeconds := int64(p.cfg.Processor.ProcessMetricsSeconds)
	times.WaitAlign(windowSeconds)
	ticker := time.NewTicker(time.Second * time.Duration(windowSeconds))
	defer ticker.Stop()
	for t := range ticker.C {
		metrics := &model.Metrics{}
		metrics.NormalLabels = p.normalLabels
		metrics.TimestampMs = t.Unix() / windowSeconds * windowSeconds * 1000 // s -> 对齐 -> ms.
		runtimes.Write(metrics)
		p.exporter.Export(metrics)
	}
}

// reportGalileoRuntimes 定时上报 galileo runtime 指标
func (p *processor) reportGalileoRuntimes() {
	if !p.cfg.Processor.EnableProcessMetrics {
		return
	}
	ticker := time.NewTicker(time.Second) // 默认 1s 窗口
	defer ticker.Stop()
	for t := range ticker.C {
		metrics := &model.Metrics{}
		metrics.NormalLabels = p.normalLabels
		metrics.TimestampMs = t.UnixMilli()
		runtimes.WriteGalileoMetric(metrics)
		p.exporter.Export(metrics)
	}
}
