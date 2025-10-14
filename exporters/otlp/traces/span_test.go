// Copyright 2024 Tencent Galileo Authors

// Package traces ...
package traces

import (
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	semconv "galiosight.ai/galio-sdk-go/semconv/v1.0.0"
)

func (s *suited) TestNoReadSpan() {
	// for coverage
	_, sp := s.disable.Start(s.ctx, "test-span")
	span := sp.(Span)
	s.Equal(false, span.SpanContext().IsSampled())
	s.Equal("", span.Name())
	s.Equal(false, span.Parent().IsSampled())
	s.Equal(trace.SpanKindUnspecified, span.SpanKind())
	s.Equal(true, span.StartTime().IsZero())
	s.Equal(true, span.EndTime().IsZero())
	s.Nil(span.Attributes())
	s.Nil(span.Links())
	s.Equal(codes.Unset, span.Status().Code)
	s.Equal(0, span.DroppedAttributes())
	s.Equal(0, span.DroppedLinks())
	s.Equal(0, span.DroppedEvents())
	s.Equal(0, span.ChildSpanCount())
}

func (s *suited) TestClearParentSampling() {
	_, sp := s.sampled.Start(s.ctx, "")
	sc := sp.SpanContext()
	s.True(sc.IsSampled())
	s.Equal("g=w:1;s:4;r:4", sc.TraceState().String())
	sc = ClearParentSampling(sp.SpanContext())
	s.False(sc.IsSampled())
	s.Equal("g=w:1", sc.TraceState().String())
}

func (s *suited) TestSetAttribute() {
	// attribute 是会去重的，详见 doc.go
	_, sp := s.sampled.Start(s.ctx, "")
	span := sp.(Span)
	sp.SetAttributes(semconv.TrpcCalleeMethodKey.String("method_a"))
	n := len(span.Attributes())
	sp.SetAttributes(semconv.TrpcCalleeMethodKey.String("method_b"))
	s.Equal(n, len(span.Attributes()))
}

func (s *suited) TestTraceState() {
	nctx, _ := s.sampled.Start(s.ctx, "")
	// test SpanFronContext 获取到我们派生的 span
	span, _ := trace.SpanFromContext(nctx).(Span)
	s.NotNil(span)
	ts := span.TraceState()
	s.Equal("w:1;s:4;r:4", ts.String())
}
