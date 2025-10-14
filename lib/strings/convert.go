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
	"reflect"
	"unsafe"
)

// copy from prometheus source code

// NoAllocString convert []byte to string
func NoAllocString(buf []byte) string {
	return *(*string)(unsafe.Pointer(&buf))
}

// NoAllocBytes convert string to []byte
func NoAllocBytes(buf string) []byte {
	// not safe: return *(*[]byte)(unsafe.Pointer(&buf))
	x := (*reflect.StringHeader)(unsafe.Pointer(&buf))
	h := reflect.SliceHeader{Data: x.Data, Len: x.Len, Cap: x.Len}
	return *(*[]byte)(unsafe.Pointer(&h))
}
