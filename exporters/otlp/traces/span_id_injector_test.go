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
	"go.opentelemetry.io/otel/trace"
)

func (s *suited) TestSpanIDInjector() {
	gen := NewSpanIDInjector(nil)
	sid := gen.NewSpanID(s.ctx, trace.TraceID{})
	s.NotEqual("", sid.String())
	_, sid = gen.NewIDs(s.ctx)
	s.NotEqual("", sid.String())
	ctx := InjectSpanID(s.ctx, trace.SpanID([8]byte{0x1}))
	sid = gen.NewSpanID(ctx, trace.TraceID{})
	s.Equal("0100000000000000", sid.String())
	_, sid = gen.NewIDs(ctx)
	s.Equal("0100000000000000", sid.String())
}
