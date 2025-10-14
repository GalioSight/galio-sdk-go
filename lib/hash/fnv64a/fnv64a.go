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

package fnv64a

// https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function
const (
	offset64 = 14695981039346656037 // FNV offset basis value
	prime64  = 1099511628211        // FNV prime value
)

// New 构造 fnv64a hash 值。
func New() uint64 {
	return offset64
}

// Add 增加一个字符串 s 到 hash 值 h 中，返回更新后的 hash 值。
func Add(h uint64, s string) uint64 {
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= prime64
	}
	return h
}

// AddByte 增加一个字节 b 到 hash 值 h 中，返回更新后的 hash 值。
func AddByte(h uint64, b byte) uint64 {
	h ^= uint64(b)
	h *= prime64
	return h
}

// AddUint64 增加一个 uint64 b 到 hash 值 h 中，返回更新后的 hash 值。
func AddUint64(h, b uint64) uint64 {
	h ^= b
	h *= prime64
	return h
}
