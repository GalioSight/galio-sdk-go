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
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protowire"
)

func TestDecodeTag(t *testing.T) {
	type want struct {
		num   protowire.Number
		wtype protowire.Type
		idx   int
	}
	var f = &protoField{}
	tests := []struct {
		name    string
		buffer  *buffer
		want    want
		wantErr bool
	}{
		{
			name:   "Fixed32 Type",
			buffer: newBuffer(hexStrToBytes("15"), f),
			want: want{
				num:   2,
				wtype: protowire.Fixed32Type,
				idx:   1,
			},
			wantErr: false,
		},
		{
			name:   "Fixed64 Type",
			buffer: newBuffer(hexStrToBytes("19"), f),
			want: want{
				num:   3,
				wtype: protowire.Fixed64Type,
				idx:   1,
			},
			wantErr: false,
		},
		{
			name:   "Bytes Type",
			buffer: newBuffer(hexStrToBytes("1a"), f),
			want: want{
				num:   3,
				wtype: protowire.BytesType,
				idx:   1,
			},
			wantErr: false,
		},
		{
			name:   "Varint Type",
			buffer: newBuffer(hexStrToBytes("18"), f),
			want: want{
				num:   3,
				wtype: protowire.VarintType,
				idx:   1,
			},
			wantErr: false,
		},
		{
			name:    "error empty",
			buffer:  newBuffer(hexStrToBytes(""), f),
			want:    want{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				num, wtype, err := tt.buffer.decodeTag()
				if err != nil {
					assert.True(t, tt.wantErr)
					return
				}
				assert.NoError(t, err)
				assert.Equal(t, tt.want.num, num)
				assert.Equal(t, tt.want.wtype, wtype)
				assert.Equal(t, tt.want.idx, tt.buffer.idx)
			},
		)
	}
}

func TestDecodeVarint(t *testing.T) {
	var f = &protoField{}
	tests := []struct {
		name    string
		buffer  *buffer
		want    uint64
		wantIdx int
		wantErr bool
	}{
		{
			name:    "value 1",
			buffer:  newBuffer(hexStrToBytes("01"), f),
			want:    1,
			wantIdx: 1,
			wantErr: false,
		},
		{
			name:    "error overflow",
			buffer:  newBuffer(hexStrToBytes("80808080808080808080"), f),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := tt.buffer.decodeVarint()
				if err != nil {
					assert.True(t, tt.wantErr)
					return
				}
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
				assert.Equal(t, tt.wantIdx, tt.buffer.idx)
			},
		)
	}
}

func TestDecodeFixed32(t *testing.T) {
	var f = &protoField{}
	tests := []struct {
		name    string
		buffer  *buffer
		want    uint64
		wantIdx int
		wantErr bool
	}{
		{
			name:    "value 0",
			buffer:  newBuffer(hexStrToBytes("00000000"), f),
			want:    0,
			wantIdx: 4,
			wantErr: false,
		},
		{
			name:    "error empty",
			buffer:  newBuffer(hexStrToBytes(""), f),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := tt.buffer.decodeFixed32()
				if err != nil {
					assert.True(t, tt.wantErr)
					return
				}
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
				assert.Equal(t, tt.wantIdx, tt.buffer.idx)
			},
		)
	}
}

func TestDecodeFixed64(t *testing.T) {
	var f = &protoField{}
	tests := []struct {
		name    string
		buffer  *buffer
		want    uint64
		wantIdx int
		wantErr bool
	}{
		{
			name:    "value 0",
			buffer:  newBuffer(hexStrToBytes("0000000000000000"), f),
			want:    0,
			wantIdx: 8,
			wantErr: false,
		},
		{
			name:    "error empty",
			buffer:  newBuffer(hexStrToBytes(""), f),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := tt.buffer.decodeFixed64()
				if err != nil {
					assert.True(t, tt.wantErr)
					return
				}
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
				assert.Equal(t, tt.wantIdx, tt.buffer.idx)
			},
		)
	}
}

func TestDecodeRawBytes(t *testing.T) {
	var f = &protoField{}
	tests := []struct {
		name    string
		buffer  *buffer
		want    []byte
		wantIdx int
		wantErr bool
	}{
		{
			name:    "value hello",
			buffer:  newBuffer([]byte("\x05hello"), f),
			want:    []byte("hello"),
			wantIdx: 6,
			wantErr: false,
		},
		{
			name:    "error empty",
			buffer:  newBuffer(hexStrToBytes(""), f),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := tt.buffer.decodeRawBytes(false)
				if err != nil {
					assert.True(t, tt.wantErr)
					return
				}
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
				assert.Equal(t, tt.wantIdx, tt.buffer.idx)
			},
		)
	}
}

func hexStrToBytes(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil
	}
	return b
}
