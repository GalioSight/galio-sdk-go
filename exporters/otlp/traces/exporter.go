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

package traces

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"

	"galiosight.ai/galio-sdk-go/components"
	"galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/configs/ocp"
	"galiosight.ai/galio-sdk-go/configs/traces"
	attrutil "galiosight.ai/galio-sdk-go/exporters/otlp/traces/attribute"
	"galiosight.ai/galio-sdk-go/exporters/otlp/traces/internal"
	"galiosight.ai/galio-sdk-go/model"
)

type exporter struct {
	trace.Tracer
	sampler       *adaptiveSampler
	deferred      DeferredSampler
	enableProfile bool // 是否开启 span 与 profile 关联
}

type GalileoExporter = exporter // 导出

// Watch 观察配置
func (e *exporter) Watch(readOnlyConfig *ocp.GalileoConfig) {
	e.UpdateConfig(
		traces.NewConfig(
			&readOnlyConfig.Resource,
			traces.WithProcessor(&readOnlyConfig.Config.TracesConfig.Processor),
			traces.WithExporter(&readOnlyConfig.Config.TracesConfig.Exporter),
		),
	)
}

// Tracer 伽利略的 Tracer
type Tracer interface {
	trace.Tracer
	DeferredSampler() DeferredSampler
}

var _ components.TracesExporter = (*exporter)(nil)

// UpdateConfig 配置热更新
func (e *exporter) UpdateConfig(cfg *configs.Traces) {
	e.enableProfile = cfg.Processor.EnableProfile
	e.sampler.UpdateConfig(updateSamplerOption(&cfg.Processor)...)
	e.deferred.UpdateConfig(updateDeferredConfig(&cfg.Processor))
}

// Start 创建一个 span 和包含这个 span 的 context
func (e *exporter) Start(
	ctx context.Context,
	spanName string, opts ...trace.SpanStartOption,
) (context.Context, trace.Span) {
	ctx, span := e.Tracer.Start(ctx, spanName, opts...)
	ds := NewSpan(span)
	if span.SpanContext().IsSampled() && e.enableProfile {
		s := &profileSpan{
			Span:      ds,
			originCtx: ctx,
		}
		var labels []string
		if span.SpanContext().HasSpanID() {
			labels = append(labels, "span_id", span.SpanContext().SpanID().String())
		}
		if spanName != "" {
			labels = append(labels, "span_name", spanName)
		}
		if ros, ok := span.(sdktrace.ReadOnlySpan); ok {
			attributes := ros.Attributes()
			for _, attr := range attributes {
				if attr.Value.AsString() == "" {
					continue
				}
				labels = append(labels, string(attr.Key), attr.Value.AsString())
			}
		}
		ctx = addPProfLabels(ctx, labels)
		return trace.ContextWithSpan(ctx, s), s
	}
	// 使用我们的 span 覆盖 otel 原生的 span，方便使用
	return trace.ContextWithSpan(ctx, ds), ds
}

func (e *exporter) DeferredSampler() DeferredSampler {
	return e.deferred
}

func updateSamplerOption(tracesProcessor *model.TracesProcessor) []AdaptiveSamplerOption {
	return []AdaptiveSamplerOption{
		WithFraction(tracesProcessor.Sampler.Fraction),
		WithDeferredSample(tracesProcessor.EnableDeferredSample),
		WithDyeing(tracesProcessor.Sampler.EnableDyeing, tracesProcessor.Sampler.Dyeing),
		WithEnableMinSample(tracesProcessor.Sampler.EnableMinSample),
		WithWorkflow(&tracesProcessor.WorkflowSampler), // 为了不影响自适应采样器已有单测逻辑，又希望伽利略插件默认启用 Workflow 采样器，故此处显式启用
		WithLimiter(tracesProcessor.Sampler.RateLimit),
		WithBloomDyeing(tracesProcessor.Sampler.EnableBloomDyeing, tracesProcessor.Sampler.BloomDyeing),
		WithServer(tracesProcessor.Sampler.Server),
		WithClient(tracesProcessor.Sampler.Client),
	}
}

func updateDeferredConfig(tracesProcessor *model.TracesProcessor) *DeferredSampleConfig {
	return &DeferredSampleConfig{
		Enabled:            tracesProcessor.EnableDeferredSample,
		SampleError:        tracesProcessor.DeferredSampleError,
		SampleSlowDuration: time.Duration(tracesProcessor.DeferredSampleSlowDurationMs) * time.Millisecond,
		ErrorFraction:      tracesProcessor.Sampler.ErrorFraction,
	}
}

// NewExporter 构建 galileo trace exporter
func NewExporter(cfg *configs.Traces) (components.TracesExporter, error) {
	sampler := NewAdaptiveSampler(updateSamplerOption(&cfg.Processor)...)
	deferredSampler := NewWorkflowDefer(NewDeferredSampler(updateDeferredConfig(&cfg.Processor)))
	tp, err := NewTracerProvider(
		cfg.Exporter.Collector.Addr,
		WithSampler(sampler),
		WithDeferredSampler(deferredSampler),
		WithGRPCDialOption(grpc.WithChainUnaryInterceptor(recovery())),
		WithModel(cfg.SchemaURL, &cfg.Resource),
		WithBatchSpanProcessorOption(
			// TODO toraxie 目前不能 reload，后续有需求再补
			WithMaxQueueSize(int(cfg.Exporter.BufferSize)),
			WithBatchTimeout(time.Duration(cfg.Exporter.WindowSeconds)*time.Second),
			WithMaxExportBatchSize(int(cfg.Exporter.PageSize)),
			WithMaxPacketSize(int(cfg.Exporter.PacketSize)),
			WithExportToFile(cfg.Exporter.ExportToFile),
			WithLog(cfg.Log),
		),
		WithResourceSpanProcessorOption(
			// 这里的属性是上报到 span 的 Tags
			WithNameSpace(cfg.Resource.Namespace),
			WithEnvName(cfg.Resource.EnvName),
		),
		WithAPIKey(cfg.APIKey),
	)
	if err != nil {
		cfg.Stats.TracesStats.InitErrorTotal.Inc()
		return nil, err
	}
	ep := &exporter{
		Tracer:   tp.Tracer(""),
		sampler:  sampler,
		deferred: deferredSampler,
	}
	ep.UpdateConfig(cfg)
	tpw := &tracerProviderWrapper{
		TracerProvider: tp,
		ep:             ep,
	}
	otel.SetTracerProvider(tpw)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			internal.TraceContext{},
			propagation.Baggage{},
		),
	)
	attrutil.Affinity.SetTarget(cfg.Resource.Target)
	ocp.AddWatcher(cfg.Resource.Target, ep)
	return ep, nil
}

func (e *exporter) UserSampler() UserSampler {
	return e.sampler.user
}

// SetUserSampler 设置用户自定义采样器
func (e *exporter) SetUserSampler(u UserSampler) {
	e.sampler.customUser = true // 现在的实现无法检查是否设置回默认值，所以凡是调用过该函数都认为是自定义过了
	e.sampler.user = u
}

func recovery() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
	) (err error) {
		defer func() {
			if rec := recover(); rec != nil {
			}
		}()
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
