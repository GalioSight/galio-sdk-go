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

// Package traces trace 导出器
package traces

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	apitrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"

	expres "galiosight.ai/galio-sdk-go/internal/resource"
	"galiosight.ai/galio-sdk-go/model"
	semconv "galiosight.ai/galio-sdk-go/semconv/v1.0.0"

	// 启用 gzip 压缩，需要 import gzip 包
	_ "google.golang.org/grpc/encoding/gzip"
)

var (
	// MaxSendMessageSize gRPC 最大消息处理，默认 4M
	MaxSendMessageSize = 4 * 1024 * 1024
)

// NewTracerProvider 创建一个 apitrace.TracerProvider 对象
func NewTracerProvider(endpoint string, options ...SetupOption) (apitrace.TracerProvider, error) {
	o := &setupOptions{
		httpEnabled: true, // 由于外网不支持 grpc，所以默认使用 http
	}
	for _, opt := range options {
		opt(o)
	}
	exporter, err := newSpanExporter(endpoint, o)
	if err != nil {
		return nil, err
	}
	var providerOpts []sdktrace.TracerProviderOption
	// 设置采样器
	providerOpts = append(providerOpts, sdktrace.WithSampler(o.sampler))

	var processor sdktrace.SpanProcessor
	processor = NewDeferredSampleProcessor(
		NewBatchSpanProcessor(exporter, o.batchSpanOption...),
		o.deferredSampler,
	)

	if o.resource.SchemaURL() == semconv.SchemaURL {
		// OMP v1 使用旧兼容逻辑，NewResource 会塞 trpc.namespace 到 attr 中，
		// 这种非透明逻辑放到 baseSDK 中非常不合理。
		processor = NewResourceProcessor(processor, o.resSpanProcessorOption...)
	}

	// 设置 processor
	providerOpts = append(providerOpts, sdktrace.WithSpanProcessor(processor), sdktrace.WithResource(o.resource))
	providerOpts = append(providerOpts, TracerProviderOptions()...)
	traceProvider := sdktrace.NewTracerProvider(providerOpts...)
	return traceProvider, nil
}

// tracerProviderWrapper 包装 apitrace.TracerProvider 并替换 Tracer 方法，引入 exporter 做为 global Tracer
type tracerProviderWrapper struct {
	apitrace.TracerProvider
	ep *exporter
}

// Tracer 包装 apitrace.TracerProvider 的 Tracer 方法，返回 exporter 做为具体实现
func (tpw *tracerProviderWrapper) Tracer(name string, opts ...apitrace.TracerOption) apitrace.Tracer {
	ep := tpw.ep
	ep.Tracer = tpw.TracerProvider.Tracer(name, opts...)
	return ep
}

// setupOptions 分为 1) resource 参数; 2）启用配置、sampler
type setupOptions struct {
	model    model.Resource
	resource *resource.Resource

	logEnabled             bool
	metricEnabled          bool
	httpEnabled            bool
	grpcDialOptions        []grpc.DialOption
	sampler                sdktrace.Sampler
	deferredSampler        DeferredSampler
	batchSpanOption        []BatchSpanProcessorOption
	resSpanProcessorOption []ResourceSpanProcessorOption
	apiKey                 string
}

// SetupOption OpenTelemetry 配置选项
type SetupOption func(*setupOptions)

func newSpanExporter(endpoint string, o *setupOptions) (sdktrace.SpanExporter, error) {
	if o.httpEnabled {
		return newHTTPSpanExporter(endpoint, o)
	}
	return newGRPCSpanExporter(endpoint, o)
}

func newHTTPSpanExporter(endpoint string, o *setupOptions) (sdktrace.SpanExporter, error) {
	otlpTraceOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
		otlptracehttp.WithHeaders(
			map[string]string{
				model.TenantHeaderKey: o.model.GetTenantId(),
				model.TargetHeaderKey: o.model.GetTarget(),
				model.APIKeyHeaderKey: o.apiKey,
			},
		),
	}
	// 如果是内网域名，则使用 http，否则使用 https。
	// 由于配置原因，otlp 内网域名不支持 https。
	exporter, err := otlptracehttp.New(context.Background(), otlpTraceOpts...)
	if err != nil {
		return nil, err
	}
	return exporter, nil
}

func newGRPCSpanExporter(endpoint string, o *setupOptions) (sdktrace.SpanExporter, error) {
	otlpTraceOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithCompressor("gzip"),
		otlptracegrpc.WithHeaders(
			map[string]string{
				model.TenantHeaderKey: o.model.GetTenantId(),
				model.TargetHeaderKey: o.model.GetTarget(),
				model.APIKeyHeaderKey: o.apiKey,
			},
		),
		otlptracegrpc.WithDialOption(grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(MaxSendMessageSize))),
	}
	if len(o.grpcDialOptions) > 0 {
		otlpTraceOpts = append(otlpTraceOpts, otlptracegrpc.WithDialOption(o.grpcDialOptions...))
	}
	exporter, err := otlptracegrpc.New(context.Background(), otlpTraceOpts...)
	if err != nil {
		return nil, err
	}
	return exporter, nil
}

// WithLogEnabled 是否开启 log
func WithLogEnabled(enabled bool) SetupOption {
	return func(options *setupOptions) {
		options.logEnabled = enabled
	}
}

// WithGRPCDialOption 指定 gRPC dial option
func WithGRPCDialOption(opts ...grpc.DialOption) SetupOption {
	return func(cfg *setupOptions) {
		cfg.grpcDialOptions = opts
	}
}

// WithModel 指定 resource
func WithModel(schemaURL string, res *model.Resource) SetupOption {
	return func(options *setupOptions) {
		options.model = *res
		options.resource = expres.GenResource(schemaURL, res, expres.SchemaTypeTrace)
	}
}

// WithSampler 指定 sampler
func WithSampler(sampler sdktrace.Sampler) SetupOption {
	return func(options *setupOptions) {
		options.sampler = sampler
	}
}

// WithDeferredSampler 传入延迟采样过滤函数
func WithDeferredSampler(deferredSampler DeferredSampler) SetupOption {
	return func(cfg *setupOptions) {
		cfg.deferredSampler = deferredSampler
	}
}

// WithBatchSpanProcessorOption 设置异步批量上报参数
func WithBatchSpanProcessorOption(opts ...BatchSpanProcessorOption) SetupOption {
	return func(cfg *setupOptions) {
		cfg.batchSpanOption = opts
	}
}

// WithResourceSpanProcessorOption 设置 Resource 环境量上报参数
func WithResourceSpanProcessorOption(opts ...ResourceSpanProcessorOption) SetupOption {
	return func(cfg *setupOptions) {
		cfg.resSpanProcessorOption = opts
	}
}

// WithMetricEnabled 启用 metric
func WithMetricEnabled(enabled bool) SetupOption {
	return func(cfg *setupOptions) {
		cfg.metricEnabled = enabled
	}
}

// WithHTTPEnabled 是否启用 HTTP 传输协议，默认为 gRPC
func WithHTTPEnabled(enabled bool) SetupOption {
	return func(cfg *setupOptions) {
		cfg.httpEnabled = enabled
	}
}

// Shutdown 进程结束前上传所有未上传数据
func Shutdown(ctx context.Context) error {
	if tp, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider); ok {
		if err := tp.Shutdown(ctx); err != nil {
			return err
		}
	}
	return nil
}

// WithAPIKey 传入伽利略 API Key
func WithAPIKey(apiKey string) SetupOption {
	return func(opts *setupOptions) {
		opts.apiKey = apiKey
	}
}
