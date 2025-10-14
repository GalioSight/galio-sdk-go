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

// Package traces 追踪功能实现。
package traces

import (
	"context"
	"math/rand"
	"time"

	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/atomic"

	"galiosight.ai/galio-sdk-go/exporters/otlp/traces/tracestate"
)

var _ sdktrace.SpanProcessor = (*DeferredSampleProcessor)(nil)

// DeferredSampler 延迟采样，处理 span.End 之后的过滤条件 返回 true 则保留，返回 false 则 drop
type DeferredSampler interface {
	// 需要手动提前调用，可以在此时机更新 attributes，可以重复调用，结果一致
	DeferSample(sp Span) tracestate.Strategy
	// 由 SDK 在 Span.End 时调用，此时 sp 是 snapshot，无法更新 attributes，不能重复调用
	ShouldSample(sp sdktrace.ReadOnlySpan) tracestate.Strategy

	UpdateConfig(cfg *DeferredSampleConfig)
}

type deferredSampler struct {
	cfg *DeferredSampleConfig

	sampledCounter   atomic.Int64
	errorCounter     atomic.Int64
	slowCounter      atomic.Int64
	unsampledCounter atomic.Int64
}

// DeferredSampleConfig 延迟采样配置
type DeferredSampleConfig struct {
	Enabled            bool          // 是否启用，不启用则不会过滤
	SampleError        bool          // 采样出错的
	SampleSlowDuration time.Duration // 采样慢操作的
	ErrorFraction      float64       // 错误采样的比例
}

// NewDeferredSampler 根据选项创建一个
func NewDeferredSampler(cfg *DeferredSampleConfig) *deferredSampler {
	return &deferredSampler{cfg: cfg}
}

func (s *deferredSampler) shouldSample(sp sdktrace.ReadOnlySpan) tracestate.Strategy {
	// 出错的
	if s.cfg.SampleError && sp.Status().Code == codes.Error {
		// internal span 的都是 codes.Unset，也需要忽略
		if randSample(s.cfg.ErrorFraction) {
			s.errorCounter.Inc()
			return tracestate.StrategyError
		}
	}
	// 高耗时的
	if s.cfg.SampleSlowDuration != 0 {
		end := sp.EndTime()
		if end.IsZero() { // 手动调用后置采样
			end = time.Now()
		}
		if end.Sub(sp.StartTime()) >= s.cfg.SampleSlowDuration {
			s.slowCounter.Inc()
			return tracestate.StrategySlow
		}
	}
	s.unsampledCounter.Inc()
	return tracestate.StrategyNotMatch
}

// ShouldSample 不允许重入
func (s *deferredSampler) ShouldSample(sp sdktrace.ReadOnlySpan) tracestate.Strategy {
	// 已经采样的
	if sp.SpanContext().IsSampled() {
		return tracestate.StrategyMatch
	}
	if !s.cfg.Enabled {
		return tracestate.StrategyNotMatch
	}
	// 之前已经手动执行过后置采样了，直接读取结果并返回
	for _, a := range sp.Attributes() {
		if a.Key == deferredSampleKey {
			return tracestate.ParseStrategy(a.Value.AsString())
		}
	}
	// 兜底降级，如果多次调用可能造成结果不一致
	return s.shouldSample(sp)
}

// DeferSample 手动执行后置采样计算，可以重复调用
func (s *deferredSampler) DeferSample(sp Span) tracestate.Strategy {
	// 已经采样的
	if sp.SpanContext().IsSampled() {
		return tracestate.StrategyMatch
	}
	if !s.cfg.Enabled {
		return tracestate.StrategyNotMatch
	}
	if sp.DeferStrategy() == tracestate.StrategyNotExist {
		sp.SetDeferStrategy(s.shouldSample(sp))
	}
	return sp.DeferStrategy()
}

func (s *deferredSampler) UpdateConfig(cfg *DeferredSampleConfig) {
	s.cfg = cfg
}

func randSample(fraction float64) bool {
	// 大量用户会配置 1，判断一下热点分支，可以减少一次 rand 调用。
	if fraction == 1 {
		return true
	}
	return rand.Float64() < fraction
}

// DeferredSampleProcessor 延迟采样 processor, 处理 span.End 之后的过滤条件
type DeferredSampleProcessor struct {
	next            sdktrace.SpanProcessor
	deferredSampler DeferredSampler
}

// NewDeferredSampleProcessor 创建一个延迟采样 processor
func NewDeferredSampleProcessor(
	next sdktrace.SpanProcessor,
	sampleFunc DeferredSampler,
) *DeferredSampleProcessor {
	return &DeferredSampleProcessor{
		next:            next,
		deferredSampler: sampleFunc,
	}
}

// OnStart is called when a span is started. It is called synchronously
// and should not block.
func (p *DeferredSampleProcessor) OnStart(parent context.Context, s sdktrace.ReadWriteSpan) {
	p.next.OnStart(parent, s)
}

// OnEnd is called when span is finished. It is called synchronously and
// hence not block.
func (p *DeferredSampleProcessor) OnEnd(s sdktrace.ReadOnlySpan) {
	if p.deferredSampler == nil {
		// keep
		p.next.OnEnd(s)
		return
	}
	strategy := p.deferredSampler.ShouldSample(s)
	if strategy >= tracestate.StrategyMatch {
		// keep
		p.next.OnEnd(s)
		return
	}
	// drop
}

// Shutdown is called when the SDK shuts down. Any cleanup or release of
// resources held by the processor should be done in this call.
//
// Calls to OnStart, OnEnd, or ForceFlush after this has been called
// should be ignored.
//
// All timeouts and cancellations contained in ctx must be honored, this
// should not block indefinitely.
func (p *DeferredSampleProcessor) Shutdown(ctx context.Context) error {
	return p.next.Shutdown(ctx)
}

// ForceFlush exports all ended spans to the configured Exporter that have not yet
// been exported.  It should only be called when absolutely necessary, such as when
// using a FaaS provider that may suspend the process after an invocation, but before
// the Processor can export the completed spans.
func (p *DeferredSampleProcessor) ForceFlush(ctx context.Context) error {
	return p.next.ForceFlush(ctx)
}
