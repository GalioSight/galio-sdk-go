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

// Package trace ...
package traces

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"galiosight.ai/galio-sdk-go/model"
	"github.com/stretchr/testify/assert"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// go test -v -bench=MinCount -benchtime=5s .
func BenchmarkMinCount(b *testing.B) {
	var minSampleCount int32 = 1
	s := NewAdaptiveSampler(
		WithMinSampleCount(minSampleCount),
		WithWindowInterval(time.Minute),
	)

	var i int32
	b.RunParallel(
		func(pb *testing.PB) {
			for pb.Next() {
				msg := fmt.Sprint(atomic.AddInt32(&i, 1))
				s.minCount(msg)
			}
		},
	)
}

func Benchmark_randomSampler(b *testing.B) {
	t, err := trace.TraceIDFromHex("bbfa18aad8c45cc9e77bcc69b3408f94")
	assert.Nil(b, err)
	s := NewAdaptiveSampler(
		WithServer(model.RpcSamplingConfig{
			Rpc: []model.RpcConfig{
				{
					Name:     "name1",
					Fraction: 0.3,
				},
			},
		}),
	)
	p := &sdktrace.SamplingParameters{
		TraceID: t,
		Name:    "name1",
		Kind:    trace.SpanKindServer,
	}
	b.RunParallel(
		func(pb *testing.PB) {
			for pb.Next() {
				s.opts.randomSampleConf.randomSampler(p)
			}
		},
	)
}
