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

import "google.golang.org/protobuf/encoding/protowire"

type buffer struct {
	buf []byte
	idx int
	len int

	f *protoField
}

type protoField struct {
	field protowire.Number // field number
	typ   protowire.Type   // proto wire type
	u64   uint64
	data  []byte
}

func newBuffer(buf []byte, f *protoField) *buffer {
	return &buffer{buf: buf, idx: 0, len: len(buf), f: f}
}

func (b *buffer) end() bool {
	return b.idx >= b.len
}

func (b *buffer) decodeTag() (protowire.Number, protowire.Type, error) {
	num, wtype, n := protowire.ConsumeTag(b.buf[b.idx:])
	if n < 0 {
		return 0, 0, protowire.ParseError(n)
	}
	b.idx += n
	return num, wtype, nil
}

func (b *buffer) decodeVarint() (uint64, error) {
	v, n := protowire.ConsumeVarint(b.buf[b.idx:])
	if n < 0 {
		return 0, protowire.ParseError(n)
	}
	b.idx += n
	return uint64(v), nil
}

func (b *buffer) decodeFixed32() (uint64, error) {
	v, n := protowire.ConsumeFixed32(b.buf[b.idx:])
	if n < 0 {
		return 0, protowire.ParseError(n)
	}
	b.idx += n
	return uint64(v), nil
}

func (b *buffer) decodeFixed64() (uint64, error) {
	v, n := protowire.ConsumeFixed64(b.buf[b.idx:])
	if n < 0 {
		return 0, protowire.ParseError(n)
	}
	b.idx += n
	return uint64(v), nil
}

func (b *buffer) decodeRawBytes(alloc bool) ([]byte, error) {
	v, n := protowire.ConsumeBytes(b.buf[b.idx:])
	if n < 0 {
		return nil, protowire.ParseError(n)
	}
	b.idx += n
	if alloc {
		v = append([]byte(nil), v...)
	}
	return v, nil
}
