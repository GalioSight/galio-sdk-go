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
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/pprof/profile"
	pprofile "github.com/google/pprof/profile"
)

func TestSimpleDeltaProfilerDelta(t *testing.T) {
	tests := []struct {
		name       string
		deltaTypes []DeltaValueType
		old        string
		new        string
		want       string
	}{
		{
			name: "Heap",
			deltaTypes: []DeltaValueType{
				{Type: "alloc_objects", Unit: "count"},
				{Type: "alloc_space", Unit: "bytes"},
			},
			old: `
alloc_objects/count alloc_space/bytes inuse_objects/count inuse_space/bytes
main 3 6 12 24
main;bar 2 4 8 16
main;foo 5 10 20 40`,
			new: `
alloc_objects/count alloc_space/bytes inuse_objects/count inuse_space/bytes
main 4 8 16 32
main;bar 2 4 8 16
main;foo 8 16 32 64
main;foobar 7 14 28 56`,
			want: `
alloc_objects/count alloc_space/bytes inuse_objects/count inuse_space/bytes
main 1 2 16 32
main;bar 0 0 8 16
main;foo 3 6 32 64
main;foobar 7 14 28 56`,
		},
		{
			name: "Mutex",
			deltaTypes: []DeltaValueType{
				{Type: "contentions", Unit: "count"},
				{Type: "delay", Unit: "nanoseconds"},
			},
			old: `
contentions/count delay/nanoseconds
main 3 1
main;bar 2 1
main;foo 5 1`,
			new: `
contentions/count delay/nanoseconds
main 4 1
main;bar 3 1
main;foo 8 1
main;foobar 7 1`,
			want: `
contentions/count delay/nanoseconds
main 1 0
main;bar 1 0
main;foo 3 0
main;foobar 7 1`,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				profiler, err := getSimplerProfiler(tt.deltaTypes)
				if err != nil {
					t.Fatal(err)
				}
				old, err := textToProfile(tt.old)
				if err != nil {
					t.Fatal(err)
				}
				profiler.prevProf = old
				new, err := textToProfile(tt.new)
				if err != nil {
					t.Fatal(err)
				}
				var buf bytes.Buffer
				if err = new.Write(&buf); err != nil {
					t.Fatal(err)
				}
				delta, err := profiler.Delta(buf.Bytes())
				if err != nil {
					t.Fatal(err)
				}
				got, err := pprofile.ParseData(delta)
				if err != nil {
					t.Fatal(err)
				}
				want, err := textToProfile(tt.want)
				if err != nil {
					t.Fatal(err)
				}
				if !sameProfilesString(got, want) {
					t.Errorf("TestSimpleDeltaProfilerDelta \ngot %+v, \nwant %+v", got, want)
				}
			},
		)
	}
}

func sameProfilesString(got, want *pprofile.Profile) bool {
	gotString := fmt.Sprintf("%+v", got)
	wantString := fmt.Sprintf("%+v", want)
	return strings.Compare(gotString, wantString) == 0
}

func getSimplerProfiler(st []DeltaValueType) (*SimpleProfiler, error) {
	return &SimpleProfiler{
		sampleTypes: st,
	}, nil
}

func textToProfile(text string) (*pprofile.Profile, error) {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("text has to be more than two lines")
	}
	prof := &pprofile.Profile{
		SampleType: []*profile.ValueType{
			{
				Type: "samples",
				Unit: "count",
			},
		},
		PeriodType: &profile.ValueType{},
	}
	var err error
	if prof, err = setSampleType(lines[0], prof); err != nil {
		return nil, err
	}
	return setStackTrace(lines[1:], prof)
}

func setSampleType(header string, prof *pprofile.Profile) (*pprofile.Profile, error) {
	header = strings.TrimSpace(header)
	prof.SampleType = nil
	for _, sampleType := range strings.Split(header, " ") {
		// heap profile 需要设置正确的 PeriodType
		if sampleType == "alloc_space" {
			prof.PeriodType = &pprofile.ValueType{Type: "space", Unit: "bytes"}
		}
		parts := strings.Split(sampleType, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("bad sample type header: %s", header)
		}
		prof.SampleType = append(
			prof.SampleType, &profile.ValueType{
				Type: parts[0],
				Unit: parts[1],
			},
		)
	}
	return prof, nil
}

func setStackTrace(lines []string, prof *pprofile.Profile) (*pprofile.Profile, error) {
	var (
		functionID = uint64(1)
		locationID = uint64(1)
	)
	m := &profile.Mapping{ID: 1, HasFunctions: true}
	prof.Mapping = []*profile.Mapping{m}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		parts := strings.Split(line, " ")
		// parts 长度为 stack + values，比如，
		// (header) alloc_objects/count alloc_space/bytes inuse_objects/count inuse_space/bytes
		// (line) main;bar 2 4 8 16
		if len(parts) != len(prof.SampleType)+1 {
			return nil, fmt.Errorf("bad length of line: %s", line)
		}

		sample := &pprofile.Sample{}
		for _, sampleValue := range parts[1:] {
			val, err := strconv.ParseInt(sampleValue, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("bad sample value: %s, %+v", line, err)
			}
			sample.Value = append(sample.Value, val)
		}

		stack := strings.Split(parts[0], ";")
		for i := range stack {
			frame := stack[len(stack)-i-1]
			function := &pprofile.Function{
				ID:   functionID,
				Name: frame,
			}
			prof.Function = append(prof.Function, function)
			functionID++
			location := &pprofile.Location{
				ID:      locationID,
				Address: locationID,
				Mapping: m,
				Line:    []profile.Line{{Function: function}},
			}
			prof.Location = append(prof.Location, location)
			locationID++

			sample.Location = append(sample.Location, location)
		}

		prof.Sample = append(prof.Sample, sample)
	}
	return prof, prof.CheckValid()
}
