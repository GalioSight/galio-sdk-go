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

// Package internal ...
package internal

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"sync"

	sdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Copy From go.opentelemetry.io/otel/sdk@v1.28.0/trace/id_generator.go make it exportable
type randomIDGenerator struct {
	sync.Mutex
	randSource *rand.Rand
}

var _ sdk.IDGenerator = (*randomIDGenerator)(nil)

// NewSpanID returns a non-zero span ID from a randomly-chosen sequence.
func (gen *randomIDGenerator) NewSpanID(ctx context.Context, traceID trace.TraceID) trace.SpanID {
	gen.Lock()
	defer gen.Unlock()
	sid := trace.SpanID{}
	for {
		_, _ = gen.randSource.Read(sid[:])
		if sid.IsValid() {
			break
		}
	}
	return sid
}

// NewIDs returns a non-zero trace ID and a non-zero span ID from a
// randomly-chosen sequence.
func (gen *randomIDGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	gen.Lock()
	defer gen.Unlock()
	tid := trace.TraceID{}
	sid := trace.SpanID{}
	for {
		_, _ = gen.randSource.Read(tid[:])
		if tid.IsValid() {
			break
		}
	}
	for {
		_, _ = gen.randSource.Read(sid[:])
		if sid.IsValid() {
			break
		}
	}
	return tid, sid
}

// NewRandom 生成 Random ID
func NewRandom() sdk.IDGenerator {
	gen := &randomIDGenerator{}
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	gen.randSource = rand.New(rand.NewSource(rngSeed))
	return gen
}
