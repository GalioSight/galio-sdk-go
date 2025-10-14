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

// Package galio 这是基础 SDK，支持 ocp+omp+otlp 等协议。
// 用户通常应该只调用此包内的方法。其他包的内部函数不保证的稳定性。
// 此包内的 public 函数，要保证兼容性，不能随意变更签名或者删除。
package galio

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"galiosight.ai/galio-sdk-go/components"
	"galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/exporters/otlp/traces"
	"galiosight.ai/galio-sdk-go/helper"
	"galiosight.ai/galio-sdk-go/lib/otelzap"
	"galiosight.ai/galio-sdk-go/model"
)

var (
	// defaultMetricsProcessor 默认的指标数据处理器。
	defaultMetricsProcessor components.MetricsProcessor = components.NoopMetricsProcessor{}

	// defaultTracesExporter 默认的追踪数据导出器
	defaultTracesExporter components.TracesExporter = components.NewNoopTracesExporter()

	defaultProfilesProcessor components.ProfilesProcessor = components.NoopProfilersProcessor{}

	// defaultLogger 默认的日志对象
	defaultLogger *zap.Logger = zap.NewNop()

	// eventLogger 默认的事件日志对象
	eventLogger *zap.Logger = zap.NewNop()

	// WithTracerProviderOptions 支持覆盖 provider options (可能会导致 galileo 功能失效，请谨慎使用)
	WithTracerProviderOptions = traces.WithTracerProviderOptions
	// NewSpanIDInjector 支持注入 SpanID
	NewSpanIDInjector = traces.NewSpanIDInjector
	// InjectSpanID 注入 span id
	InjectSpanID = traces.InjectSpanID
)

// SetDefaultMetricsProcessor 设置默认的指标处理器。
// 此函数通常在插件初始化时调用。
// 此函数不是线程安全的。
func SetDefaultMetricsProcessor(metricsProcessor components.MetricsProcessor) {
	defaultMetricsProcessor = metricsProcessor
}

// GetDefaultMetricsProcessor 获取默认的指标处理器
// 此函数通常在调用 SetDefaultMetricsProcessor 之后执行。
// 此函数未加锁。
// 不要同时并发调用 SetDefaultMetricsProcessor。
func GetDefaultMetricsProcessor() components.MetricsProcessor {
	return defaultMetricsProcessor
}

// NewMetricsProcessor 创建一个 MetricsProcessor。
// 通常情况下，此方法只需要调用一次，创建出对象后要进行重用。
// 注意，此方法开销是非常大的，内部会创建多个协程，所以不能频繁创建。
// 此方法是线程安全的。
// 多次调用 NewMetricsProcessor 时，会创建出多个 MetricsProcessor 对象。
// 如果配置有错误，会返回 error。
// 所以业务只能忽略错误，或者 panic 掉，由业务自己决定。
func NewMetricsProcessor(metricsCfg *configs.Metrics) (components.MetricsProcessor, error) {
	return helper.GetMetricsProcessor(metricsCfg)
}

// ClientMetrics 主调指标数据上报。
// 此方法是线程安全的。
func ClientMetrics(clientMetrics *model.ClientMetrics) {
	defaultMetricsProcessor.ProcessClientMetrics(clientMetrics)
}

// ServerMetrics 被调指标数据上报。
// 此方法是线程安全的。
func ServerMetrics(serverMetrics *model.ServerMetrics) {
	defaultMetricsProcessor.ProcessServerMetrics(serverMetrics)
}

// NormalMetric 属性指标上报。
// 此方法是线程安全的。
func NormalMetric(normalMetric *model.NormalMetric) {
	defaultMetricsProcessor.ProcessNormalMetric(normalMetric)
}

// CustomMetrics 自定义指标上报。
// 此方法是线程安全的。
func CustomMetrics(customMetrics *model.CustomMetrics) {
	defaultMetricsProcessor.ProcessCustomMetrics(customMetrics)
}

// NewTracesExporter 创建一个 TracesExporter。
// 通常情况下，此方法只需要调用一次，创建出对象后可以进行重用。
// 此方法是线程安全的。
func NewTracesExporter(cfg *configs.Traces) (components.TracesExporter, error) {
	return helper.GetTracesExporter(cfg)
}

// SetDefaultTracesExporter 设置默认的指标追踪器。
// 此函数通常在插件初始化时调用。
// 此函数不是线程安全的。
func SetDefaultTracesExporter(exporter components.TracesExporter) {
	defaultTracesExporter = exporter
}

// GetDefaultTracesExporter 获取默认的
// 此函数通常在调用 SetDefaultTracesExporter 之后执行。
// 此函数未加锁。
// 不要同时并发调用 SetDefaultTracesExporter。
func GetDefaultTracesExporter() components.TracesExporter {
	return defaultTracesExporter
}

// Start 启动追踪。
func Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return defaultTracesExporter.Start(ctx, spanName, opts...)
}

// WithSpan 为方法生产一个指定名字的 span。
func WithSpan(
	ctx context.Context, spanName string, fn func(ctx context.Context) error,
	opts ...trace.SpanStartOption,
) error {
	ctx, sp := defaultTracesExporter.Start(ctx, spanName, opts...)
	defer func() {
		if tracer, ok := defaultTracesExporter.(traces.Tracer); ok {
			if span, ok := sp.(traces.Span); ok {
				// 要么都不执行后置采样，要么提前执行，保证 gate.state 一致，便于理解。
				tracer.DeferredSampler().DeferSample(span)
			}
		}
		sp.End()
	}()
	return fn(ctx)
}

// NewLogsExporter 创建一个 LogsExporter。
// 通常情况下，此方法只需要调用一次，创建出对象后可以进行重用。
// 此方法是线程安全的。
func NewLogsExporter(cfg *configs.Logs) (components.LogsExporter, error) {
	return helper.GetLogsExporter(cfg)
}

// NewProfilesProcessor 创建一个 ProfilesProcessor
// 通常情况下，此方法只需要调用一次，创建出对象后可以进行重用。
// 注意，此方法开销是非常大的，内部会创建多个协程，所以不能频繁创建。
// 此方法是线程安全的。
func NewProfilesProcessor(cfg *configs.Profiles) (components.ProfilesProcessor, error) {
	return helper.GetProfilesProcessor(cfg)
}

// SetDefaultProfilesProcessor 设置默认的性能数据处理器。
// 此函数通常在插件初始化时调用。
// 此函数不是线程安全的。
func SetDefaultProfilesProcessor(processor components.ProfilesProcessor) {
	defaultProfilesProcessor = processor
}

// GetDefaultProfilesProcessor 获取默认的 ProfilesProcessor
// 此函数通常在调用 SetDefaultProfilesProcessor 之后执行。
// 此函数未加锁。
// 不要同时并发调用 SetDefaultProfilesProcessor
func GetDefaultProfilesProcessor() components.ProfilesProcessor {
	return defaultProfilesProcessor
}

// NewLogger 创建一个 LogsExporter。
// 通常情况下，此方法只需要调用一次，创建出对象后可以进行重用。
// 此方法是线程安全的。
func NewLogger(cfg *configs.Logs, options ...zap.Option) (*zap.Logger, error) {
	return otelzap.NewLogger(cfg, options...)
}

// GetLogger 获取日志对象
func GetLogger() *zap.Logger {
	return defaultLogger
}

// SetLogger 设置日志对象
func SetLogger(log *zap.Logger) {
	defaultLogger = log
}

// SetEventLogger 设置事件日志对象
func SetEventLogger(log *zap.Logger) {
	eventLogger = log
}

// ReportEvent 上报一种特殊的日志作为事件。
// 非 trpc 框架里面，调用此函数前，需要先调用 SetEventLogger 进行初始化。
// 具体的事件上报包含以下几个方面：
// msg：上报的具体信息，如 panic 栈信息。
// source：事件来源，例如 PCG-123、STKE、galileo 等，表示谁上报的这个事件，可以是平台、进程名、包名等，不能为空。
// domain：事件领域，如 browser（浏览器应用事件）、device（移动应用事件）或者 k8s（Kubernetes 的事件），不能为空。
// name：事件名称，如 panic、exception 等，不能为空。
// extFields：扩展字段，用于提供更多上下文信息，可以是任意数量的 zap.Field，用于记录和查询。
// 注意：source、domain 和 name 会以"event.source"、"event.domain"、"event.name"的形式上报。
func ReportEvent(msg, source, domain, name string, extFields ...zap.Field) {
	// 将必须字段和扩展字段添加到 eventFields 中，然后上报。
	eventFields := []zap.Field{
		zap.String("event.source", source),
		zap.String("event.domain", domain),
		zap.String("event.name", name),
	}
	eventFields = append(eventFields, extFields...)

	eventLogger.Error(msg, eventFields...)
}

var noopSpan = traces.NewSpan(
	func() trace.Span {
		_, span := trace.NewNoopTracerProvider().Tracer("").Start(context.Background(), "noop")
		return span
	}(),
)

// SpanFromContext 获取伽利略 Span, 可以使用更多用户友好的函数
func SpanFromContext(ctx context.Context) traces.Span {
	span, _ := trace.SpanFromContext(ctx).(traces.Span)
	if span == nil {
		return noopSpan
	}
	return span
}
