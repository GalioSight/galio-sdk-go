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

package strings

import (
	"math"
	"testing"
)

func TestLeFloatToString(t *testing.T) {
	type args struct {
		f float64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "1",
			args: args{f: 1.0},
			want: "1",
		},
		{
			name: "0.1",
			args: args{f: 0.1},
			want: "0.1",
		},
		{
			name: "0.01",
			args: args{f: 0.01},
			want: "0.01",
		},
		{
			name: "0.001",
			args: args{f: 0.001},
			want: "0.001",
		},
		{
			name: "0.0001",
			args: args{f: 0.0001},
			want: "0.0001",
		},
		{
			name: "+Inf",
			args: args{f: math.Inf(1)},
			want: "+Inf",
		},
		{
			name: "-Inf",
			args: args{f: math.Inf(-1)},
			want: "-Inf",
		},
		{
			name: "NaN",
			args: args{f: math.NaN()},
			want: "NaN",
		},
		{
			name: "-1",
			args: args{f: -1},
			want: "-1",
		},
		{
			name: "10",
			args: args{f: 10},
			want: "10",
		},
		{
			name: "100",
			args: args{f: 100},
			want: "100",
		},
		{
			name: "1000",
			args: args{f: 1000},
			want: "1000",
		},
		{
			name: "100000000",
			args: args{f: 100000000},
			want: "1e+08",
		},
		{
			name: "100000000.000000001",
			args: args{f: 100000000.000000001},
			want: "1e+08",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := LeFloatToString(tt.args.f); got != tt.want {
					t.Errorf("LeFloatToString() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestLeFloatToStringEqual(t *testing.T) {
	type args struct {
		f float64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "1",
			args: args{f: 1.0},
		},
		{
			name: "0.1",
			args: args{f: 0.1},
		},
		{
			name: "0.01",
			args: args{f: 0.01},
		},
		{
			name: "0.001",
			args: args{f: 0.001},
		},
		{
			name: "0.0001",
			args: args{f: 0.0001},
		},
		{
			name: "Inf",
			args: args{f: math.Inf(1)},
		},
		{
			name: "-Inf",
			args: args{f: math.Inf(-1)},
		},
		{
			name: "NaN",
			args: args{f: math.NaN()},
		},
		{
			name: "-1",
			args: args{f: -1},
		},
		{
			name: "10",
			args: args{f: 10},
		},
		{
			name: "100",
			args: args{f: 100},
		},
		{
			name: "1000",
			args: args{f: 1000},
		},
		{
			name: "100000000",
			args: args{f: 100000000},
		},
		{
			name: "100000000.000000001",
			args: args{f: 100000000.000000001},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := LeFloatToString(tt.args.f); got != FloatToString(tt.args.f) {
					t.Errorf("LeFloatToString() = %v, want %v", got, FloatToString(tt.args.f))
				}
			},
		)
	}
}

func BenchmarkFloatToString(b *testing.B) {
	for n := 0; n < b.N; n++ {
		FloatToString(1)
	}
}

func BenchmarkLeFloatToString(b *testing.B) {
	for n := 0; n < b.N; n++ {
		LeFloatToString(1)
	}
}
