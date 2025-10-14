// Package traces ...
//
// Copyright 2024 Tencent Galileo Authors
package traces

import (
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"galiosight.ai/galio-sdk-go/exporters/otlp/traces/tracestate"
)

// workflowDefer WorkflowPathSampler 分成 2 个部分，这里是后置采样部分
type workflowDefer struct {
	impl DeferredSampler
}

var _ DeferredSampler = (*workflowDefer)(nil)

// NewWorkflowDefer 创建一个 WorkflowPathSampler
func NewWorkflowDefer(impl DeferredSampler) *workflowDefer {
	return &workflowDefer{impl: impl}
}

func (w *workflowDefer) DeferSample(sp Span) tracestate.Strategy {
	return w.impl.DeferSample(sp)
}

func (w *workflowDefer) ShouldSample(sp sdktrace.ReadOnlySpan) tracestate.Strategy {
	ret := w.impl.ShouldSample(sp)
	if ret <= tracestate.StrategyNotMatch {
		ts, _ := tracestate.Parse(sp.SpanContext().TraceState().Get(galileoVendor))
		if ts.Workflow.Sampled() {
			// 所有 workflow 后置采样都命中
			return tracestate.StrategyMatch
		}
	}
	return ret
}

func (w *workflowDefer) UpdateConfig(cfg *DeferredSampleConfig) {
	w.impl.UpdateConfig(cfg)
}
