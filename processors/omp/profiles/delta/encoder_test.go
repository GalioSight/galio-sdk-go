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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEecodeField(t *testing.T) {
	tests := []struct {
		name    string
		num     int
		message Message
		want    []byte
	}{
		{
			name: "SampleType",
			num:  int(SampleTypeNum),
			message: &SampleType{
				ValueType{
					Type: 1,
					Unit: 1,
				},
			},
			want: []byte("\n\x04\b\x01\x10\x01"),
		},
		{
			name: "Label",
			num:  3,
			message: &Label{
				Key:     1,
				Str:     1,
				NumUnit: 1,
			},
			want: []byte("\x1a\x06\b\x01\x10\x01 \x01"),
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := &encoder{}
				e.encodeField(tt.num, tt.message)
				assert.Equal(t, e.buf, tt.want)
			},
		)
	}
}

func TestEncodeUint64(t *testing.T) {
	tests := []struct {
		name string
		num  int
		x    uint64
		want []byte
	}{
		{
			name: "value 1 field 1",
			num:  1,
			x:    1,
			want: []byte("\b\x01"),
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := &encoder{}
				e.encodeUint64(tt.num, tt.x)
				assert.Equal(t, e.buf, tt.want)
			},
		)
	}
}

func TestEncodeUint64s(t *testing.T) {
	tests := []struct {
		name string
		num  int
		x    []uint64
		want []byte
	}{
		{
			name: "value 1,2,3 field 1",
			num:  1,
			x:    []uint64{1, 2, 3},
			want: []byte("\n\x03\x01\x02\x03"),
		},
		{
			name: "value 1,2 field 1",
			num:  1,
			x:    []uint64{1, 2},
			want: []byte("\b\x01\b\x02"),
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := &encoder{}
				e.encodeUint64s(tt.num, tt.x)
				assert.Equal(t, tt.want, e.buf)
			},
		)
	}
}

func TestEncodeInt64s(t *testing.T) {
	tests := []struct {
		name string
		num  int
		x    []int64
		want []byte
	}{
		{
			name: "value 1,2,3 field 1",
			num:  1,
			x:    []int64{1, 2, 3},
			want: []byte("\n\x03\x01\x02\x03"),
		},
		{
			name: "value 1,2 field 1",
			num:  1,
			x:    []int64{1, 2},
			want: []byte("\b\x01\b\x02"),
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := &encoder{}
				e.encodeInt64s(tt.num, tt.x)
				assert.Equal(t, tt.want, e.buf)
			},
		)
	}
}

func TestEncodeBytes(t *testing.T) {
	tests := []struct {
		name string
		num  int
		x    []byte
		want []byte
	}{
		{
			name: "value hello field 1",
			num:  1,
			x:    []byte("hello"),
			want: []byte("\n\x05hello"),
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := &encoder{}
				e.encodeBytes(tt.num, tt.x)
				assert.Equal(t, tt.want, e.buf)
			},
		)
	}
}

func TestEncodeBool(t *testing.T) {
	tests := []struct {
		name string
		num  int
		x    bool
		want []byte
	}{
		{
			name: "value true field 1",
			num:  1,
			x:    true,
			want: []byte("\b\x01"),
		},
		{
			name: "value false field 2",
			num:  2,
			x:    false,
			want: []byte("\x10\x00"),
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				e := &encoder{}
				e.encodeBool(tt.num, tt.x)
				assert.Equal(t, tt.want, e.buf)
			},
		)
	}
}
