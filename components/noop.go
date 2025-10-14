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

package components

import (
	"context"

	"galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/configs/ocp"
	"galiosight.ai/galio-sdk-go/model"
	"go.opentelemetry.io/otel/trace"
)

// NoopMetricsProcessor 空对象模式，可以提升性能，避免每次判断空指针。
// 因为空指针只在未初始化的时候会产生，正常初始化完成之后是不会有空指针的。
// 所以刚开始创建一个空对象 NoopMetricsProcessor。等初始化完成，用真实对象代替。
type NoopMetricsProcessor struct {
}

// Watch 配置
func (n NoopMetricsProcessor) Watch(readOnlyConfig *ocp.GalileoConfig) {
}

// GetStats 获取自监控数据。
func (n NoopMetricsProcessor) GetStats() *model.SelfMonitorStats {
	return nil
}

// ProcessClientMetrics 处理主调监控。
func (n NoopMetricsProcessor) ProcessClientMetrics(clientMetrics *model.ClientMetrics) {
}

// ProcessServerMetrics 处理被调监控。
func (n NoopMetricsProcessor) ProcessServerMetrics(serverMetrics *model.ServerMetrics) {
}

// ProcessNormalMetric 处理属性监控。
func (n NoopMetricsProcessor) ProcessNormalMetric(normalMetric *model.NormalMetric) {
}

// ProcessCustomMetrics 处理用户自定义监控。
func (n NoopMetricsProcessor) ProcessCustomMetrics(customMetrics *model.CustomMetrics) {
}

// UpdateConfig 更新配置。
func (n NoopMetricsProcessor) UpdateConfig(cfg *configs.Metrics) {
}

// NewNoopTracesExporter 创建空的 traces 导出器。
func NewNoopTracesExporter() NoopTracesExporter {
	return NoopTracesExporter{
		Tracer: trace.NewNoopTracerProvider().Tracer(""),
	}
}

// NoopTracesExporter 空的 traces 导出器。
type NoopTracesExporter struct {
	trace.Tracer
}

// Watch 空对象 Watch
func (n NoopTracesExporter) Watch(readOnlyConfig *ocp.GalileoConfig) {
}

// Start 空对象 Start。
func (n NoopTracesExporter) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (
	context.Context, trace.Span,
) {
	return n.Tracer.Start(ctx, spanName, opts...)
}

// UpdateConfig 空对象 UpdateConfig。
func (n NoopTracesExporter) UpdateConfig(cfg *configs.Traces) {
}

// NoopProfilersProcessor 空的性能数据处理器
type NoopProfilersProcessor struct {
}

// Watch 更新配置
func (n NoopProfilersProcessor) Watch(readOnlyConfig *ocp.GalileoConfig) {
}

// Start 开起采集性能数据处理器。
func (n NoopProfilersProcessor) Start() {
}

// Shutdown 停止采集性能数据并关闭性能数据处理器
func (n NoopProfilersProcessor) Shutdown() {
}

// UpdateConfig 更新配置。
func (n NoopProfilersProcessor) UpdateConfig(cfg *configs.Profiles) {
}
