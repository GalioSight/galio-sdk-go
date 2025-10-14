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

// Package bloom 布隆过滤器实现。
// 参考 https://github.com/bits-and-blooms/bloom
package bloom

import (
	"math"

	"github.com/spaolacci/murmur3"
)

// wordSize bitmap 实现数组中每个元素的大小。
const wordSize = 64

// log2WordSize bitmap 实现数组中每个元素大小取 log2 的对数。
const log2WordSize = 6

// wordsNeeded 根据 bit 空间大小计算所需要的数组大小。
func wordsNeeded(i int32) int {
	return int((i + (wordSize - 1)) >> log2WordSize)
}

// wordsIndex 计算 bitmap 中第 i 个位置在对应 byte 中的序号。
func wordsIndex(i int32) int32 {
	return i & (wordSize - 1)
}

// BloomFilter 布隆过滤器。
type BloomFilter struct {
	m int32
	k int32
	b []int64
}

// EstimateParameters 根据元素数量 n 和误判率 p 估算哈希空间大小 m 和哈希函数个数 k.
func EstimateParameters(n int32, p float64) (m int32, k int32) {
	m = int32(math.Ceil(-1 * float64(n) * math.Log(p) / math.Pow(math.Log(2), 2)))
	k = int32(math.Ceil(math.Log(2) * float64(m) / float64(n)))
	return
}

// New 新建布隆过滤器。
func New(m int32, k int32) *BloomFilter {
	bitmap := make([]int64, wordsNeeded(max(1, m)))
	return &BloomFilter{max(1, m), max(1, k), bitmap}
}

// From 根据已有数据创建布隆过滤器。
func From(m int32, k int32, bitmap []int64) *BloomFilter {
	if bitmap == nil || len(bitmap) == 0 {
		return New(m, k)
	}
	return &BloomFilter{max(1, m), max(1, k), bitmap}
}

func max(x, y int32) int32 {
	if x > y {
		return x
	}
	return y
}

// Cap 返回布隆过滤器容量 m.
func (f *BloomFilter) Cap() int32 {
	return f.m
}

// K 返回布隆过滤器哈希函数数量 k.
func (f *BloomFilter) K() int32 {
	return f.k
}

// Bitmap 返回布隆过滤器数据 bitmap.
func (f *BloomFilter) Bitmap() []int64 {
	return f.b
}

// Add 向布隆过滤器中加入元素。
func (f *BloomFilter) Add(data string) {
	for i := int32(0); i < f.k; i++ {
		location := f.location(data, i)
		f.b[location>>log2WordSize] |= 1 << wordsIndex(location)
	}
}

// Test 测试一个元素是否在布隆过滤器中。
// 返回 true，则元素可能在布隆过滤器中。
// 返回 false，则元素一定不存在布隆过滤器中。
func (f *BloomFilter) Test(data string) bool {
	for i := int32(0); i < f.k; i++ {
		location := f.location(data, i)
		if f.b[location>>log2WordSize]&(1<<wordsIndex(location)) == 0 {
			return false
		}
	}
	return true
}

// location 根据数据和哈希函数序号计算在 bitmap 中的位置。
func (f *BloomFilter) location(data string, i int32) int32 {
	h, _ := murmur3.Sum128WithSeed([]byte(data), uint32(i))
	return int32(h % uint64(f.m))
}
