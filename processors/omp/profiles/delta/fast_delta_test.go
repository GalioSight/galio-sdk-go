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
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/pprof/profile"
	"github.com/stretchr/testify/require"
)

func TestFastDeltaCaclulator(t *testing.T) {
	tests := []struct {
		name             string
		round1           string
		round2           string
		durationNanos    int64
		deltaSampleTypes []DeltaValueType
	}{
		{
			name:          "heap",
			round1:        "testdata/heap.round1.pprof",
			round2:        "testdata/heap.round2.pprof",
			durationNanos: 5960465000,
			deltaSampleTypes: []DeltaValueType{
				{Type: "alloc_objects", Unit: "count"},
				{Type: "alloc_space", Unit: "bytes"},
			},
		},
		{
			name:          "block",
			round1:        "testdata/block.round1.pprof",
			round2:        "testdata/block.round2.pprof",
			durationNanos: 60137720785,
			deltaSampleTypes: []DeltaValueType{
				{Type: "contentions", Unit: "count"},
				{Type: "delay", Unit: "nanoseconds"},
			},
		},
		{
			name:          "mutex",
			round1:        "testdata/mutex.round1.pprof",
			round2:        "testdata/mutex.round2.pprof",
			durationNanos: 10106602752,
			deltaSampleTypes: []DeltaValueType{
				{Type: "contentions", Unit: "count"},
				{Type: "delay", Unit: "nanoseconds"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				round1, err := readPprofFile(tt.round1)
				if err != nil {
					t.Fatal(err)
				}
				round2, err := readPprofFile(tt.round2)
				if err != nil {
					t.Fatal(err)
				}

				fp := newFastDeltaCaculator(tt.deltaSampleTypes)

				if err := fp.CalculateDeltas(round1, io.Discard); err != nil {
					t.Fatal(err)
				}

				out1 := new(bytes.Buffer)
				if err := fp.CalculateDeltas(round2, out1); err != nil {
					t.Fatal(err)
				}

				delta1, err := profile.ParseData(out1.Bytes())
				if err != nil {
					t.Fatal(err)
				}

				// 使用 simpleProfiler 去验证
				round1Prof, err := profile.ParseData(round1)
				if err != nil {
					t.Fatal(err)
				}
				sp := &SimpleProfiler{
					prevProf:    round1Prof,
					sampleTypes: tt.deltaSampleTypes,
				}
				out2, err := sp.Delta(round2)
				if err != nil {
					t.Fatal(err)
				}
				delta2, err := profile.ParseData(out2)
				if err != nil {
					t.Fatal(err)
				}

				delta2.Scale(-1)
				diff, err := profile.Merge([]*profile.Profile{delta1, delta2})
				if err != nil {
					t.Fatal(err)
				}
				if len(diff.Sample) != 0 {
					t.Errorf("the diff of fast amd simple delta is not empty.\nfast:\n%v\nsimple:\n%v", delta1, delta2)
				}
				require.Equal(t, tt.durationNanos, delta1.DurationNanos)
			},
		)
	}
}

func TestNegtiveValue(t *testing.T) {
	//data := map[string]*profile.Profile{}
	tests := []struct {
		name             string
		deltaSampleTypes []DeltaValueType
	}{
		{
			name: "mutex",
			deltaSampleTypes: []DeltaValueType{
				{Type: "contentions", Unit: "count"},
				{Type: "delay", Unit: "nanoseconds"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				fp := newFastDeltaCaculator(tt.deltaSampleTypes)
				path := "testdata/mutex_full"
				dirs, err := os.ReadDir(path)
				if err != nil {
					t.Fatal(err)
				}
				for _, dir := range dirs {
					if !dir.IsDir() {
						continue
					}
					file := filepath.Join(path, dir.Name(), "mutex.pprof")
					profBytes, err := readPprofFile(file)
					if err != nil {
						t.Fatal(err)
					}
					out := new(bytes.Buffer)
					if err := fp.CalculateDeltas(profBytes, out); err != nil {
						t.Fatal(err)
					}
					delta, err := profile.ParseData(out.Bytes())
					if err != nil {
						t.Fatal(err)
					}
					for _, sample := range delta.Sample {
						for _, v := range sample.Value {
							if v < 0 {
								t.Errorf("sample has negative value: %v", sample.Value)
							}
						}
					}
				}
			},
		)
	}
}

func readPprofFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if isGzipData(data) {
		r := bytes.NewReader(data)
		gzr, err := gzip.NewReader(r)
		if err != nil {
			return nil, err
		}
		data, err = io.ReadAll(gzr)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}
