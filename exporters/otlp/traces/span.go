// Copyright 2024 Tencent Galileo Authors

package traces

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	apitrace "go.opentelemetry.io/otel/trace"

	"galiosight.ai/galio-sdk-go/exporters/otlp/traces/internal"
	"galiosight.ai/galio-sdk-go/exporters/otlp/traces/tracestate"
	"galiosight.ai/galio-sdk-go/semconv"
)

// Span 是一个自定义的 Span，用于兼容非 ReadWriteSpan 的场景
type Span interface {
	sdktrace.ReadWriteSpan // 最强力的 Span，简化逻辑，避免绕来绕去
	DeferStrategy() tracestate.Strategy
	SetDeferStrategy(tracestate.Strategy)
	// 用户友好 API
	TraceState() tracestate.State
	// 额外判断后置采样命中
	IsSampled() bool
}

// NewSpan 将非 ReadWriteSpan 转换为 Span
func NewSpan(s trace.Span) Span {
	ret := &deferSpan{deferred: tracestate.StrategyNotExist}
	if w, ok := s.(sdktrace.ReadWriteSpan); ok {
		ret.ReadWriteSpan = w
	} else {
		// 非 ReadWriteSpan 只可能是 sdktrace 返回的是 nonRecordingSpan
		ret.ReadWriteSpan = noReadSpan{Span: s}
	}
	ret.state, _ = tracestate.Parse(s.SpanContext().TraceState().Get(galileoVendor))
	return ret
}

type deferSpan struct {
	sdktrace.ReadWriteSpan
	deferred tracestate.Strategy
	state    tracestate.State // 一个只读的 tracestate，对外 API
}

func (ds *deferSpan) DeferStrategy() tracestate.Strategy {
	return ds.deferred
}

func (ds *deferSpan) SetDeferStrategy(s tracestate.Strategy) {
	ds.deferred = s
}

func (ds *deferSpan) IsSampled() bool {
	return ds.SpanContext().IsSampled() || ds.deferred >= tracestate.StrategyMatch
}

func (ds *deferSpan) TraceState() tracestate.State {
	return ds.state
}

// 后置采样命中结果
var deferredSampleKey = attribute.Key("galileo.deferred")

func (ds *deferSpan) End(options ...trace.SpanEndOption) {
	if ds.deferred > tracestate.StrategyMatch {
		// 用于 UI 显示，所以确保有具体值才设置
		// 覆盖后置采样结果
		state := ds.SpanContext().TraceState()
		ds.state.Sample.SampledStrategy = ds.deferred
		ts := internal.Convert(&state).Insert(galileoVendor, ds.state.String())
		ds.SetAttributes(semconv.GalileoStateKey.String(ts.String()))
	}
	if ds.deferred != tracestate.StrategyNotExist {
		// 用于标记，透传到 deferredSampleProcessor 中，无论后置采样是何种结果，都需要记录。
		ds.SetAttributes(deferredSampleKey.String(ds.deferred.String()))
	}
	ds.ReadWriteSpan.End(options...)
}

var _ Span = (*deferSpan)(nil)

// ClearParentSampling 正确的清理上游采样标记
// 原来逻辑中没有更新 trace state，会导致由 workflow 采样的 span 被错误的识别成继承上游采样，被用户看到。
func ClearParentSampling(sc apitrace.SpanContext) apitrace.SpanContext {
	ts := sc.TraceState()
	parsed, _ := tracestate.Parse(ts.Get(galileoVendor))
	parsed.Sample.RootStrategy = tracestate.StrategyNotExist
	parsed.Sample.SampledStrategy = tracestate.StrategyNotExist
	state := internal.Convert(&ts).Insert(galileoVendor, parsed.String())
	return apitrace.NewSpanContext(apitrace.SpanContextConfig{
		TraceID:    sc.TraceID(),
		SpanID:     sc.SpanID(),
		TraceFlags: sc.TraceFlags() &^ apitrace.FlagsSampled,
		TraceState: *state.Convert(),
		Remote:     true,
	})
}

type noReadSpan struct {
	trace.Span
	sdktrace.ReadOnlySpan
}

var _ (sdktrace.ReadWriteSpan) = noReadSpan{}

// === 提供常用的函数实现，注，如果调用其他 ReadOnlySpan 的函数，可能会导致 panic ===
func (n noReadSpan) SpanContext() trace.SpanContext {
	return n.Span.SpanContext()
}

func (n noReadSpan) Name() string {
	return ""
}

func (n noReadSpan) Parent() trace.SpanContext {
	return trace.SpanContext{}
}

func (n noReadSpan) SpanKind() trace.SpanKind {
	return trace.SpanKindUnspecified
}

func (n noReadSpan) StartTime() time.Time {
	return time.Time{}
}

func (n noReadSpan) EndTime() time.Time {
	return time.Time{}
}

func (n noReadSpan) Attributes() []attribute.KeyValue {
	return nil
}

func (n noReadSpan) Links() []sdktrace.Link {
	return nil
}

func (n noReadSpan) Status() sdktrace.Status {
	return sdktrace.Status{}
}

func (n noReadSpan) DroppedAttributes() int {
	return 0
}

func (n noReadSpan) DroppedLinks() int {
	return 0
}

func (n noReadSpan) DroppedEvents() int {
	return 0
}

func (n noReadSpan) ChildSpanCount() int {
	return 0
}
