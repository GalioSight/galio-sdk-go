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
	"io"

	"google.golang.org/protobuf/encoding/protowire"
)

type encoder struct {
	buf []byte
	out io.Writer
	tmp [16]byte
}

func (e *encoder) set(w io.Writer) {
	e.out = w
	e.buf = e.buf[:0]
}

func (e *encoder) encodeField(num int, m Message) {
	n1 := len(e.buf)
	m.encodeInternal(e)
	n2 := len(e.buf)
	e.append(num, n2, n1)
}

func (e *encoder) encodeUint64Nonzero(num int, x uint64) {
	if x == 0 {
		return
	}
	e.encodeUint64(num, x)
}

func (e *encoder) encodeUint64(num int, x uint64) {
	e.buf = protowire.AppendTag(e.buf, protowire.Number(num), protowire.VarintType)
	e.buf = protowire.AppendVarint(e.buf, x)
}

func (e *encoder) encodeInt64Nonzero(num int, x int64) {
	if x == 0 {
		return
	}
	e.encodeInt64(num, x)
}

func (e *encoder) encodeInt64(num int, x int64) {
	u := uint64(x)
	e.encodeUint64(num, u)
}

func (e *encoder) encodeUint64s(num int, x []uint64) {
	if len(x) > 2 {
		n1 := len(e.buf)
		for _, v := range x {
			e.buf = protowire.AppendVarint(e.buf, v)
		}
		n2 := len(e.buf)
		// 添加 tag(field number + wire type)
		e.append(num, n2, n1)
	} else {
		for _, v := range x {
			e.encodeUint64(num, v)
		}
	}
}

func (e *encoder) append(num int, n2 int, n1 int) {
	e.buf = protowire.AppendTag(e.buf, protowire.Number(num), protowire.BytesType)
	// 添加 length 长度
	e.buf = protowire.AppendVarint(e.buf, uint64(n2-n1))
	n3 := len(e.buf)
	// field number + wire type + length 放到 e.buf 前面
	// internal 数据放到 e.buf 后面
	copy(e.tmp[:], e.buf[n2:n3])
	copy(e.buf[n1+(n3-n2):], e.buf[n1:n2])
	copy(e.buf[n1:], e.tmp[:n3-n2])
}

func (e *encoder) encodeInt64s(num int, x []int64) {
	if len(x) == 0 {
		return
	}
	if len(x) > 2 {
		n1 := len(e.buf)
		for _, v := range x {
			e.buf = protowire.AppendVarint(e.buf, uint64(v))
		}
		n2 := len(e.buf)
		e.append(num, n2, n1)
	} else {
		for _, v := range x {
			e.encodeInt64(num, v)
		}
	}
}

func (e *encoder) encodeBytes(num int, x []byte) {
	e.buf = protowire.AppendTag(e.buf, protowire.Number(num), protowire.BytesType)
	e.buf = protowire.AppendVarint(e.buf, uint64(len(x)))
	e.buf = append(e.buf, x...)
}

func (e *encoder) encodeBoolTrue(num int, x bool) {
	if x {
		e.encodeBool(num, x)
	}
}

func (e *encoder) encodeBool(num int, x bool) {
	if x {
		e.encodeUint64(num, 1)
	} else {
		e.encodeUint64(num, 0)
	}
}

func (e *encoder) reset() {
	e.buf = e.buf[:0]
}

func (e *encoder) writeOut(b []byte) error {
	for len(b) > 0 {
		var (
			n   int
			err error
		)
		if n, err = e.out.Write(b); err != nil {
			return err
		}
		b = b[n:]
	}
	return nil
}
