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
	"runtime/pprof"

	"go.opentelemetry.io/otel/trace"
)

type profileSpan struct {
	Span
	// span 的 context，用于恢复 goroutine 的 pprof label
	originCtx context.Context
}

// End 结束 span 并将当前 goroutine 的 pprof label 还原回开启该 span 之前
func (s *profileSpan) End(options ...trace.SpanEndOption) {
	s.Span.End(options...)
	// 将 goroutine 的 pprof label 设置回未加 span id 的状态
	pprof.SetGoroutineLabels(s.originCtx)
}

func addPProfLabels(ctx context.Context, labels []string) context.Context {
	pprofCtx := pprof.WithLabels(ctx, pprof.Labels(labels...))
	pprof.SetGoroutineLabels(pprofCtx)
	return ctx
}
