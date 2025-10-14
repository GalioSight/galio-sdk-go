// Copyright 2025 Tencent Galileo Authors
//
// Copyright 2025 Tencent OpenTelemetry Oteam
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

// Package traces ...
package traces

import (
	"context"

	sdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"galiosight.ai/galio-sdk-go/exporters/otlp/traces/internal"
)

type spanInjector struct {
	gen sdk.IDGenerator
}

// NewSpanIDInjector 创建一个 SpanIDInjector
func NewSpanIDInjector(gen sdk.IDGenerator) sdk.IDGenerator {
	if gen == nil {
		gen = internal.NewRandom()
	}
	return spanInjector{gen: gen}
}

type spanIDKey struct{}

func (gen spanInjector) NewSpanID(ctx context.Context, traceID trace.TraceID) trace.SpanID {
	if id, ok := ctx.Value(spanIDKey{}).(trace.SpanID); ok {
		return id
	}
	return gen.gen.NewSpanID(ctx, traceID)
}

func (gen spanInjector) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	id, ok := ctx.Value(spanIDKey{}).(trace.SpanID)
	tid, sid := gen.gen.NewIDs(ctx)
	if ok {
		sid = id
	}
	return tid, sid
}

// InjectSpanID 注入 SpanID
func InjectSpanID(ctx context.Context, id trace.SpanID) context.Context {
	return context.WithValue(ctx, spanIDKey{}, id)
}
