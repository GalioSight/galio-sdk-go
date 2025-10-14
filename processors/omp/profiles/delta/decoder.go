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
	"fmt"

	"google.golang.org/protobuf/encoding/protowire"
)

type decoder struct {
	raw  []byte
	prof SimpleProfile
}

func (d *decoder) set(p []byte) {
	d.raw = p
}

func decodeField(b *buffer) error {
	var err error
	b.f.field, b.f.typ, err = b.decodeTag()
	if err != nil {
		return err
	}
	switch b.f.typ {
	case protowire.VarintType:
		b.f.u64, err = b.decodeVarint()
	case protowire.Fixed32Type:
		b.f.u64, err = b.decodeFixed32()
	case protowire.Fixed64Type:
		b.f.u64, err = b.decodeFixed64()
	case protowire.BytesType:
		b.f.data, err = b.decodeRawBytes(false)
	default:
		return fmt.Errorf("unknow wire type: %d", b.f.typ)
	}
	if err != nil {
		return err
	}
	return nil
}

func decodeInt64s(b *buffer, x *[]int64) error {
	switch b.f.typ {
	case protowire.BytesType:
		var (
			u   uint64
			err error
		)
		buf := newBuffer(b.f.data, b.f)
		for !buf.end() {
			u, err = buf.decodeVarint()
			if err != nil {
				return err
			}
			*x = append(*x, int64(u))
		}
	case protowire.VarintType:
		*x = append(*x, int64(b.f.u64))
	default:
		return fmt.Errorf("unknow wire type: %d", b.f.typ)
	}
	return nil
}

func decodeUint64s(b *buffer, x *[]uint64) error {
	switch b.f.typ {
	case protowire.BytesType:
		var (
			u   uint64
			err error
		)
		buf := newBuffer(b.f.data, b.f)
		for !buf.end() {
			u, err = buf.decodeVarint()
			if err != nil {
				return err
			}
			*x = append(*x, u)
		}
	case protowire.VarintType:
		*x = append(*x, b.f.u64)
	default:
		return fmt.Errorf("unknow wire type: %d", b.f.typ)
	}
	return nil
}
