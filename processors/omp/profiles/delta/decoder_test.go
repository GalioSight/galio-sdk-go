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
	"google.golang.org/protobuf/encoding/protowire"
)

func TestDecodeField(t *testing.T) {
	var f = &protoField{}
	tests := []struct {
		name    string
		buffer  *buffer
		want    *protoField
		wantIdx int
		wantErr bool
	}{
		{
			name:   "value fixed64 0",
			buffer: newBuffer(hexStrToBytes("190000000000000000"), f),
			want: &protoField{
				field: 3,
				typ:   protowire.Fixed64Type,
				u64:   0,
			},
			wantIdx: 9,
			wantErr: false,
		},
		{
			name:   "value bytes hello",
			buffer: newBuffer([]byte("\x1a\x05hello"), f),
			want: &protoField{
				field: 3,
				typ:   protowire.BytesType,
				u64:   0,
				data:  []byte("hello"),
			},
			wantIdx: 7,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				err := decodeField(tt.buffer)
				if err != nil {
					assert.True(t, tt.wantErr)
					return
				}
				assert.NoError(t, err)
				assert.Equal(t, tt.want.field, tt.buffer.f.field)
				assert.Equal(t, tt.want.typ, tt.buffer.f.typ)
				assert.Equal(t, tt.want.u64, tt.buffer.f.u64)
				assert.Equal(t, tt.want.data, tt.buffer.f.data)
				assert.Equal(t, tt.wantIdx, tt.buffer.idx)
			},
		)
	}
}

func TestDecodeInt64s(t *testing.T) {
	tests := []struct {
		name    string
		buffer  *buffer
		want    *[]int64
		wantErr bool
	}{
		{
			name: "Bytes Type",
			buffer: newBuffer(
				[]byte{}, &protoField{
					typ:  protowire.BytesType,
					data: hexStrToBytes("010203"),
				},
			),
			want:    &[]int64{1, 2, 3},
			wantErr: false,
		},
		{
			name: "Varint Type",
			buffer: newBuffer(
				[]byte{}, &protoField{
					typ: protowire.VarintType,
					u64: 1,
				},
			),
			want:    &[]int64{1},
			wantErr: false,
		},
		{
			name: "Invalid Type",
			buffer: newBuffer(
				[]byte{}, &protoField{
					typ: protowire.Fixed64Type,
				},
			),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := &[]int64{}
				err := decodeInt64s(tt.buffer, got)
				if err != nil {
					assert.True(t, tt.wantErr)
					return
				}
				t.Logf("got: %v", got)
				t.Logf("want: %v", tt.want)
				assert.NoError(t, err)
				assert.Equal(t, *tt.want, *got)
			},
		)
	}
}

func TestDecodeUint64s(t *testing.T) {
	tests := []struct {
		name    string
		buffer  *buffer
		want    *[]uint64
		wantErr bool
	}{
		{
			name: "Bytes Type",
			buffer: newBuffer(
				[]byte{}, &protoField{
					typ:  protowire.BytesType,
					data: hexStrToBytes("010203"),
				},
			),
			want:    &[]uint64{1, 2, 3},
			wantErr: false,
		},
		{
			name: "Varint Type",
			buffer: newBuffer(
				[]byte{}, &protoField{
					typ: protowire.VarintType,
					u64: 1,
				},
			),
			want:    &[]uint64{1},
			wantErr: false,
		},
		{
			name: "Invalid Type",
			buffer: newBuffer(
				[]byte{}, &protoField{
					typ: protowire.Fixed64Type,
				},
			),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := &[]uint64{}
				err := decodeUint64s(tt.buffer, got)
				if err != nil {
					assert.True(t, tt.wantErr)
					return
				}
				assert.NoError(t, err)
				assert.Equal(t, *tt.want, *got)
			},
		)
	}
}
