// Copyright 2024 Tencent Galileo Authors

// Package traces ...
package traces

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	traceconf "galiosight.ai/galio-sdk-go/configs/traces"
	"galiosight.ai/galio-sdk-go/exporters/otlp/traces/tracestate"
)

type suited struct {
	suite.Suite
	ctx     context.Context
	tracer  Tracer
	disable Tracer    // 未开启后置采样
	sampled *exporter // 前置采样命中
	profile *exporter // profile 采样
}

func (s *suited) SetupSuite() {
	s.ctx = context.Background()
	res := testResource()
	c := traceconf.NewConfig(res)
	c.Processor.EnableDeferredSample = true
	c.Processor.DeferredSampleError = true
	c.Processor.DeferredSampleSlowDurationMs = 1
	c.Processor.Sampler.ErrorFraction = 1
	c.Processor.Sampler.EnableMinSample = false
	c.Processor.Sampler.Fraction = 0
	{
		tracer, err := NewExporter(c)
		s.NoError(err)
		s.tracer = tracer.(Tracer)
	}
	{
		nc := c
		nc.Processor.EnableDeferredSample = false
		tracer, err := NewExporter(nc)
		s.NoError(err)
		s.disable = tracer.(Tracer)
	}
	{
		nc := c
		nc.Processor.Sampler.Fraction = 1
		tracer, err := NewExporter(nc)
		s.NoError(err)
		s.sampled = tracer.(*exporter)
	}
	{
		nc := c
		nc.Processor.EnableProfile = true
		tracer, err := NewExporter(nc)
		s.NoError(err)
		s.profile = tracer.(*exporter)
	}
}

func (s *suited) TearDownSuite() {
}

func (s *suited) SetupTest() {
}

func (s *suited) TearDownTest() {
}

func (s *suited) TestDeferSample() {
	_, sp := s.tracer.Start(s.ctx, "test-span")
	span := sp.(Span)
	sp.SetStatus(codes.Error, "")
	strategy := s.tracer.DeferredSampler().DeferSample(span)
	s.Equal(tracestate.StrategyError, strategy)
	s.Equal(true, span.IsSampled())
	errcnt := errCount(s.tracer)
	s.Equal(strategy, s.tracer.DeferredSampler().DeferSample(span)) // 重复调用结果一致
	s.Equal(errcnt, errCount(s.tracer))                             // 并且没有执行实际的判断
	sp.SetStatus(codes.Ok, "")

	span.SetDeferStrategy(tracestate.StrategyNotExist)
	strategy = s.tracer.DeferredSampler().DeferSample(span)
	s.Equal(tracestate.StrategyNotMatch, strategy)
	s.Equal(false, span.IsSampled())

	span.SetDeferStrategy(tracestate.StrategyNotExist)
	time.Sleep(time.Millisecond * 2)
	strategy = s.tracer.DeferredSampler().DeferSample(span)
	s.Equal(tracestate.StrategySlow, strategy)
	s.Equal(true, span.IsSampled())

	span.SetDeferStrategy(tracestate.StrategyNotExist)
	s.Equal(tracestate.StrategyNotMatch, s.disable.DeferredSampler().DeferSample(span))
	s.Equal(false, span.IsSampled())

	_, sp = s.sampled.Start(s.ctx, "test-span")
	span = sp.(Span)
	s.Equal(tracestate.StrategyMatch, s.sampled.DeferredSampler().DeferSample(span))
	s.Equal(true, span.IsSampled())
}

func errCount(tracer Tracer) int64 {
	return tracer.DeferredSampler().(*workflowDefer).impl.(*deferredSampler).errorCounter.Load()
}

func (s *suited) TestShouldSample() {
	_, sp := s.tracer.Start(s.ctx, "test-span")
	span := sp.(Span)
	sp.SetStatus(codes.Error, "")
	s.tracer.DeferredSampler().DeferSample(span)

	errcnt := errCount(s.tracer)
	s.Equal(tracestate.StrategyError, s.tracer.DeferredSampler().ShouldSample(span))
	s.Less(errcnt, errCount(s.tracer)) // 因为没有标记，会重新执行判断逻辑

	span.End() // 打入标记
	errcnt = errCount(s.tracer)
	s.Equal(tracestate.StrategyError, s.tracer.DeferredSampler().ShouldSample(span))
	s.Equal(errcnt, errCount(s.tracer)) // 有标记，则没有重复执行判断

	s.Equal(tracestate.StrategyNotMatch, s.disable.DeferredSampler().ShouldSample(span))

	_, sp = s.sampled.Start(s.ctx, "test-span")
	span = sp.(Span)
	s.Equal(tracestate.StrategyMatch, s.sampled.DeferredSampler().ShouldSample(span))
}

func (s *suited) TestProfileSpan() {
	nctx, sp := s.profile.Start(s.ctx, "profile-span")
	s.True(trace.SpanFromContext(nctx) == sp)
}

func TestDefer(t *testing.T) {
	suite.Run(t, new(suited))
}
