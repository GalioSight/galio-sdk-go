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

package bloom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	ss := []string{
		"aaaaaa",
		"bbbbbb",
		"cccccc",
		"dddddd",
	}
	m, k := EstimateParameters(int32(len(ss)), 0.0001)
	b := New(m, k)
	for _, s := range ss {
		b.Add(s)
	}
	for _, s := range ss {
		assert.Equal(t, true, b.Test(s))
	}
	assert.Equal(t, false, b.Test("eeeeee"))
	assert.Equal(t, false, b.Test("ffffff"))
}

func TestNew(t *testing.T) {
	type args struct {
		m int32
		k int32
	}
	tests := []struct {
		name string
		args args
		want *BloomFilter
	}{
		{
			name: "m and k is illegal",
			args: args{m: 0, k: -1},
			want: &BloomFilter{m: 1, k: 1, b: make([]int64, 1)},
		},
		{
			name: "k is illegal",
			args: args{m: 65, k: -1},
			want: &BloomFilter{m: 65, k: 1, b: make([]int64, 2)},
		},
		{
			name: "m is illegal",
			args: args{m: -2, k: 14},
			want: &BloomFilter{m: 1, k: 14, b: make([]int64, 1)},
		},
		{
			name: "m and k legal",
			args: args{m: 64, k: 14},
			want: &BloomFilter{m: 64, k: 14, b: make([]int64, 1)},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				assert.Equalf(t, tt.want, New(tt.args.m, tt.args.k), "New(%v, %v)", tt.args.m, tt.args.k)
			},
		)
	}
}

func TestFrom(t *testing.T) {
	type args struct {
		m      int32
		k      int32
		bitmap []int64
	}
	tests := []struct {
		name string
		args args
		want *BloomFilter
	}{
		{
			name: "m and k is illegal",
			args: args{m: 0, k: -1, bitmap: nil},
			want: &BloomFilter{m: 1, k: 1, b: make([]int64, 1)},
		},
		{
			name: "k is illegal",
			args: args{m: 65, k: -1, bitmap: []int64{}},
			want: &BloomFilter{m: 65, k: 1, b: make([]int64, 2)},
		},
		{
			name: "m is illegal",
			args: args{m: -2, k: 14, bitmap: []int64{111}},
			want: &BloomFilter{m: 1, k: 14, b: []int64{111}},
		},
		{
			name: "m and k legal",
			args: args{m: 64, k: 14, bitmap: []int64{222}},
			want: &BloomFilter{m: 64, k: 14, b: []int64{222}},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				assert.Equalf(
					t, tt.want, From(tt.args.m, tt.args.k, tt.args.bitmap), "From(%v, %v, %v)", tt.args.m, tt.args.k,
					tt.args.bitmap,
				)
			},
		)
	}
}
