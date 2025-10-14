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

// Package traces ...
package traces

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	semconv "galiosight.ai/galio-sdk-go/semconv/v1.0.0"
)

func Test_adaptiveSampler_minCount(t *testing.T) {
	type arg struct {
		key  string
		want bool
	}
	tests := []struct {
		name           string
		minSampleCount int32
		args           []arg
	}{
		{
			"最小采集次数=0", 0,
			[]arg{
				{"key1", false},
			},
		},
		{
			"最小采集次数=1", 1,
			[]arg{
				{"key1", true},
				{"key1", false},
			},
		},
		{
			"最小采集次数=2", 2,
			[]arg{
				{"key1", true}, {"key1", true},
				{"key1", false}, {"key1", false},
			},
		},
		{
			"多维度组合验证", 1,
			[]arg{
				{"key1", true}, {"key1", false},
				{"key2", true}, {"key2", false},
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				s := NewAdaptiveSampler(
					WithMinSampleCount(tt.minSampleCount),
				)
				for _, arg := range tt.args {
					if got := s.minCount(arg.key); got != arg.want {
						t.Errorf("adaptiveSampler.minCount() = %v, want %v", got, arg.want)
					}
				}
				s.close()
			},
		)
	}
}

func Test_shiftWindow(t *testing.T) {
	var n int32 = 5
	s := NewAdaptiveSampler(
		WithMinSampleCount(n),
		WithWindowInterval(time.Minute),
	)

	key := "key1111"
	for i := 0; i < 100; i++ {
		for j := 0; j < int(n); j++ {
			r := s.minCount(key)
			assert.Equal(t, j <= int(n), r)
		}
		s.shiftWindow()
	}
}

func TestMinCountKey(t *testing.T) {
	tests := []struct {
		callerService string
		callerMethod  string
		calleeService string
		calleeMethod  string
		expect        string
	}{
		{"A", "a", "B", "b", "b_B"},
		{"C", "c", "B", "b", "b_B"},
		{"D", "d", "C", "c", "c_C"},
	}

	a := assert.New(t)
	for _, test := range tests {
		p := sdktrace.SamplingParameters{}
		p.Attributes = append(
			p.Attributes,
			semconv.TrpcCallerServiceKey.String(test.callerService),
			semconv.TrpcCallerMethodKey.String(test.callerMethod),
			semconv.TrpcCalleeServiceKey.String(test.calleeService),
			semconv.TrpcCalleeMethodKey.String(test.calleeMethod),
		)
		t.Run(
			"", func(t *testing.T) {
				a.Equal(test.expect, minCountKey(&p))
			},
		)
	}
}
