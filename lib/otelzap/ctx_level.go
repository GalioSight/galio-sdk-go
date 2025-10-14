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

// Package otelzap ...
package otelzap

import (
	"context"

	"galiosight.ai/galio-sdk-go/exporters/otlp/traces/tracestate"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const galileoVendor = "g" // 伽利略 SDK traceState 的 vendor Key

// WithContextDyeingLevel 支持命中染色突破日志级别
func WithContextDyeingLevel(s coreStrategy) zap.Option {
	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		if s != DefaultStrategy {
			core = NewMatchCore(core, false, s)
		}
		return newCtxCore(core, func(c zapcore.Core, ctx context.Context) zapcore.Core {
			sc := trace.SpanContextFromContext(ctx)
			ts, _ := tracestate.Parse(sc.TraceState().Get(galileoVendor)) // TODO toraxie 可能比较低效，先实现功能，再优化性能
			if m, ok := c.(*MatchCore); ok {
				c = m.Core
			}
			return NewMatchCore(c, ts.Sample.RootStrategy == tracestate.StrategyDyeing, s)
		})
	})
}

// WithContextSampleLevel 支持命中采样突破日志级别，系统内置，可以通过配置启用
func WithContextSampleLevel(s coreStrategy) zap.Option {
	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		if s != DefaultStrategy {
			core = NewMatchCore(core, false, s)
		}
		return newCtxCore(core, func(c zapcore.Core, ctx context.Context) zapcore.Core {
			sc := trace.SpanContextFromContext(ctx)
			if m, ok := c.(*MatchCore); ok {
				c = m.Core
			}
			return NewMatchCore(c, sc.IsSampled(), s)
		})
	})
}
