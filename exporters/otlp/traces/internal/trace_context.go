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

// Package internal ...
package internal

import (
	"context"
	"encoding/hex"
	"strings"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	traceparentHeader = "traceparent"
	tracestateHeader  = "tracestate"
	parentDelimiter   = "-"
	maxVersion        = 254
	supportedVersion  = 0
)

// TraceContext a really quick version
type TraceContext struct{}

var _ propagation.TextMapPropagator = TraceContext{}

// Inject set tracecontext from the Context into the carrier.
func (tc TraceContext) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return
	}
	state := sc.TraceState()
	if ts := Convert(&state).String(); ts != "" {
		carrier.Set(tracestateHeader, ts)
	}

	// Clear all flags other than the trace-context supported sampling bit.
	flags := sc.TraceFlags() & trace.FlagsSampled

	var sb strings.Builder
	sb.Grow(2 + 32 + 16 + 2 + 3)
	sb.WriteString("00-")
	writeTraceID(&sb, sc)
	sb.WriteByte('-')
	writeSpanID(&sb, sc)
	sb.WriteByte('-')
	writeFlags(&sb, byte(flags))
	carrier.Set(traceparentHeader, sb.String())
}

func writeFlags(sb *strings.Builder, flags byte) {
	var flag [2]byte
	src := [1]byte{flags}
	hex.Encode(flag[:], src[:])
	sb.Write(flag[:])
}

func writeSpanID(sb *strings.Builder, sc trace.SpanContext) {
	var spanID [16]byte
	srcSpanID := sc.SpanID()
	hex.Encode(spanID[:], srcSpanID[:])
	sb.Write(spanID[:])
}

func writeTraceID(sb *strings.Builder, sc trace.SpanContext) {
	var traceID [32]byte
	srcTraceID := sc.TraceID()
	hex.Encode(traceID[:], srcTraceID[:])
	sb.Write(traceID[:])
}

// Extract reads tracecontext from the carrier into a returned Context.
//
// The returned Context will be a copy of ctx and contain the extracted
// tracecontext as the remote SpanContext. If the extracted tracecontext is
// invalid, the passed ctx will be returned directly instead.
func (tc TraceContext) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	sc := tc.extract(carrier)
	if !sc.IsValid() {
		return ctx
	}
	return trace.ContextWithRemoteSpanContext(ctx, sc)
}

// hex.Decode 支持大写字符，这里需要额外排除
func upperHex(v string) bool {
	for _, c := range v {
		if c >= 'A' && c <= 'F' {
			return true
		}
	}
	return false
}

func (tc TraceContext) extract(carrier propagation.TextMapCarrier) trace.SpanContext {
	h := carrier.Get(traceparentHeader)
	bad := trace.SpanContext{}
	if h == "" {
		return bad
	}

	var ver [1]byte
	if !extractPart(ver[:], &h, 2) {
		return bad
	}
	version := int(ver[0])
	if version > maxVersion {
		return bad
	}

	var scc trace.SpanContextConfig
	if !extractPart(scc.TraceID[:], &h, 32) {
		return bad
	}
	if !extractPart(scc.SpanID[:], &h, 16) {
		return bad
	}

	var opts [1]byte
	if !extractPart(opts[:], &h, 2) {
		return bad
	}
	if version == 0 && (h != "" || opts[0] > 2) {
		// version 0 not allow extra
		// version 0 not allow other flag
		return bad
	}

	// Clear all flags other than the trace-context supported sampling bit.
	scc.TraceFlags = trace.TraceFlags(opts[0]) & trace.FlagsSampled
	ts, _ := ParseTraceState(carrier.Get(tracestateHeader))
	scc.TraceState = *ts.Convert()
	scc.Remote = true

	sc := trace.NewSpanContext(scc)
	if !sc.IsValid() {
		return bad
	}

	return sc
}

func extractPart(dst []byte, h *string, n int) bool {
	part, left, _ := strings.Cut(*h, parentDelimiter)
	*h = left
	if len(part) != n || upperHex(part) {
		return false
	}
	if p, err := hex.Decode(dst, []byte(part)); err != nil || p != n/2 {
		return false
	}
	return true
}

// Fields returns the keys who's values are set with Inject.
func (tc TraceContext) Fields() []string {
	return []string{traceparentHeader, tracestateHeader}
}
