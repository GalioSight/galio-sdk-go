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
	"bytes"
	"fmt"
	"math"
	"strconv"
	"sync"
)

const (
	initialNumBufSize = 24
)

var (
	floatBytesPool sync.Pool // 用来存结果
	convBytesPool  sync.Pool // 用来 strconv.AppendFloat
)

func getFloatBytes() *bytes.Buffer {
	if fb, ok := floatBytesPool.Get().(*bytes.Buffer); ok {
		return fb
	}
	return bytes.NewBuffer(make([]byte, 0, initialNumBufSize))
}

func putFloatBytes(fb *bytes.Buffer) {
	fb.Reset()
	floatBytesPool.Put(fb)
}

func getConvBytes() *[]byte {
	if cb, ok := convBytesPool.Get().(*[]byte); ok {
		return cb
	}
	cb := make([]byte, 0, initialNumBufSize)
	return &cb
}

func putConvBytes(cb *[]byte) {
	*cb = (*cb)[:0]
	convBytesPool.Put(cb)
}

// LeFloatToString Prometheus le 分桶浮点数转字符串。
// 与下面的 FloatToString 函数功能相同。LeFloatToString 专门做了性能优化。
func LeFloatToString(f float64) string {
	fb := getFloatBytes()
	defer putFloatBytes(fb)

	if _, err := writeFloat(fb, f); err == nil {
		return fb.String()
	}
	return strconv.FormatFloat(f, 'f', 3, 64)
}

func writeFloat(b *bytes.Buffer, f float64) (int, error) {
	switch {
	case f == 1:
		return 1, b.WriteByte('1')
	case f == 0:
		return 1, b.WriteByte('0')
	case f == -1:
		return b.WriteString("-1")
	case math.IsNaN(f):
		return b.WriteString("NaN")
	case math.IsInf(f, +1):
		return b.WriteString(VMRangeMax)
	case math.IsInf(f, -1):
		return b.WriteString(VMRangeMin)
	default:
		bp := getConvBytes()
		*bp = strconv.AppendFloat((*bp)[:0], f, 'g', -1, 64)
		written, err := b.Write(*bp)
		putConvBytes(bp)
		return written, err
	}
}

// FloatToString 浮点数转字符串，与 LeFloatToString 功能相同。
// 与 LeFloatToString 功能相同，LeFloatToString 专门做了性能优化。
func FloatToString(f float64) string {
	return fmt.Sprintf("%g", f)
}
