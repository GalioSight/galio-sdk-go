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
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func sampleContext(m bool) context.Context {
	var sample trace.TraceFlags
	if m {
		sample = trace.FlagsSampled
	}
	sc := trace.NewSpanContext(trace.SpanContextConfig{TraceFlags: sample})
	return trace.ContextWithSpanContext(context.Background(), sc)
}

func dyeingContext(m bool) context.Context {
	var dyeing string
	if m {
		dyeing = "g=r:2"
	}
	ts, _ := trace.ParseTraceState(dyeing)
	sc := trace.NewSpanContext(trace.SpanContextConfig{TraceState: ts})
	return trace.ContextWithSpanContext(context.Background(), sc)
}

func TestWithLevel(t *testing.T) {
	tests := []struct {
		opt  zap.Option
		ctx  context.Context
		want bool
	}{
		{WithContextSampleLevel(MustLogTraced), sampleContext(true), true},
		{WithContextSampleLevel(MustLogTraced), sampleContext(false), false},
		{WithContextDyeingLevel(MustLogTraced), dyeingContext(true), true},
		{WithContextDyeingLevel(MustLogTraced), dyeingContext(false), false},
	}

	a := assert.New(t)
	for _, test := range tests {
		t.Run(
			"", func(t *testing.T) {
				logger := zap.New(&ctxCore{}, test.opt)
				m, ok := logger.With(Context(test.ctx)).Core().(*ctxCore).Core.(*MatchCore)
				a.True(ok)
				a.Equal(test.want, m.matched)
			},
		)
	}
}

func TestContextWith(t *testing.T) {
	var n int
	opt := ContextWith(func(core zapcore.Core, ctx context.Context) zapcore.Core {
		n++
		return core
	})
	logger := zap.New(&ctxCore{}, opt)
	logger.With(Context(context.Background()))
	a := assert.New(t)
	// for coverity
	a.Equal(1, n)
}
