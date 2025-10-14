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

package delta

import (
	"bytes"
	"io"
	"runtime"
	"runtime/pprof"
	"strings"
	"testing"

	"github.com/google/pprof/profile"
	"github.com/stretchr/testify/require"
)

func BenchmarkFastProfiler(b *testing.B) {
	c := newFastDeltaCaculator(
		[]DeltaValueType{
			{Type: "alloc_objects", Unit: "count"},
			{Type: "alloc_space", Unit: "bytes"},
		},
	)
	round1, err := readPprofFile("testdata/big-heap.pprof")
	if err != nil {
		b.Fatal(err)
	}
	round2, err := readPprofFile("testdata/big-heap.pprof")
	if err != nil {
		b.Fatal(err)
	}

	b.Run(
		"warm-up", func(b *testing.B) {
			b.SetBytes(int64(len(round1)))
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				if err := c.CalculateDeltas(round1, io.Discard); err != nil {
					b.Fatal(err)
				}
				if err := c.CalculateDeltas(round2, io.Discard); err != nil {
					b.Fatal(err)
				}
			}
		},
	)

	b.Run(
		"stable", func(b *testing.B) {
			b.SetBytes(int64(len(round1)))
			b.ReportAllocs()

			if err := c.CalculateDeltas(round1, io.Discard); err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := c.CalculateDeltas(round2, io.Discard); err != nil {
					b.Fatal(err)
				}
			}
			b.StopTimer()
			reportHeapInuseSpace(b)
		},
	)

}

func BenchmarkSimpleProfiler(b *testing.B) {
	p := NewSimpleProfiler(
		[]DeltaValueType{
			{Type: "alloc_objects", Unit: "count"},
			{Type: "alloc_space", Unit: "bytes"},
		},
	)
	round1, err := readPprofFile("testdata/big-heap.pprof")
	if err != nil {
		b.Fatal(err)
	}
	round2, err := readPprofFile("testdata/big-heap.pprof")
	if err != nil {
		b.Fatal(err)
	}

	b.Run(
		"warm-up", func(b *testing.B) {
			b.SetBytes(int64(len(round1)))
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				if _, err := p.Delta(round1); err != nil {
					b.Fatal(err)
				}
				if _, err := p.Delta(round2); err != nil {
					b.Fatal(err)
				}
			}
		},
	)

	b.Run(
		"stable", func(b *testing.B) {
			b.SetBytes(int64(len(round1)))
			b.ReportAllocs()

			if _, err := p.Delta(round1); err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := p.Delta(round2); err != nil {
					b.Fatal(err)
				}
			}
			b.StopTimer()
			reportHeapInuseSpace(b)
		},
	)

}

func reportHeapInuseSpace(b *testing.B) {
	// {"inuse_space", "bytes"}
	runtime.GC()

	var buf bytes.Buffer
	pprof.Lookup("heap").WriteTo(&buf, 0)
	profile, err := profile.Parse(&buf)
	require.NoError(b, err)

	var sum float64
	for _, s := range profile.Sample {
		sum = sum + sumSampleValue(s)
	}

	b.ReportMetric(sum, "heap-inuse-B/op")
}

func sumSampleValue(s *profile.Sample) float64 {
	if s.Value[3] == 0 {
		return 0
	}
	var sum float64
	for _, loc := range s.Location {
		for _, line := range loc.Line {
			if containsFunction(line.Function.Name) {
				sum += float64(s.Value[3])
				return sum
			}
		}
	}
	return 0
}

var filter = []string{
	"go/sdk/base/processors/omp/profiles/delta.(*FastProfiler)",
	"go/sdk/base/processors/omp/profiles/delta.(*SimpleProfiler)",
}

func containsFunction(s string) bool {
	for i := range filter {
		if strings.Contains(s, filter[i]) {
			return true
		}
	}
	return false
}
