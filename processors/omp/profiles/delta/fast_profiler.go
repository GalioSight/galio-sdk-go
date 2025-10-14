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
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/kelindar/bitmap"
	"github.com/spaolacci/murmur3"
	"google.golang.org/protobuf/encoding/protowire"
)

// 目前支持 heap、mutex 和 block 三种 profile 数据的 delta 计算
// heap:
//   - SampleTypes: {"alloc_objects", "count"}, {"alloc_space", "bytes"},
//     {"inuse_objects", "count"}, {"inuse_space", "bytes"}
//   - Delta SampleTypes: {"alloc_objects", "count"}, {"alloc_space", "bytes"}
//
// mutex:
//   - SampleTypes: {"contentions", "count"}, {"delay", "nanoseconds"}
//   - Delta SampleTypes: {"contentions", "count"}, {"delay", "nanoseconds"}
//
// block:
//   - SampleTypes: {"contentions", "count"}, {"delay", "nanoseconds"}
//   - Delta SampleTypes: {"contentions", "count"}, {"delay", "nanoseconds"}
//
// 所以对应 sample values 最多包含 4种（heap)，所有类型的 profile 都支持 2 种 delta。
const (
	maxSampleValues      = 4
	maxDeltaSampleValues = 2
)

type HashKey [16]byte

// fieldMask 标记 profile 中哪些字段需要处理
// fieldMask[0]：true 则表示所有字段
// fieldMask[1..14]：对应 profile 中的各字段
type fieldMask [15]bool

// FastProfiler 优化过的 delta profiler
type FastProfiler struct {
	c   *FastDeltaCaculator
	buf *bytes.Buffer
	r   *gzip.Reader
	w   *gzip.Writer
}

// FastDeltaCaculator 实现 delta 计算逻辑
type FastDeltaCaculator struct {
	decoder  *decoder
	enconder *encoder

	// 记录该类型 profile 数据所有的 SampleType,
	// Type 和 Unit 用 stringTable 的 index 表示
	sampleTypesAll [][2]int
	// 记录哪些 SampleType 需要计算 delta
	sampleTypesDelta []ValueTypeBytes
	// 记录 sample 中的第 n 个 value 是否需要计算 delta，
	// 需要为 true，不需要为 false，
	// n 为 0..[2 - 4]
	supportDeltaIdx []bool
	// 根据 sample 计算出的 hash key，记录每种 sample 每轮采集的值
	samplesByKey map[HashKey]sampleValueTracker

	// 记录全部的 Loction，slice index 为 LocationID，元素为 Location.Address
	locationAddressIdx []uint64
	// 记录有效的 location id，
	// 如果 location 只关联全 0 的 sample，则被认为无效
	validLocationIDs bitmap.Bitmap

	// 记录已经处理过的 string
	validStringIdx bitmap.Bitmap
	// 记录已经处理过的 Location
	validFunctionIDs bitmap.Bitmap

	currentTimeNanos  int64
	durationTimeNanos int64

	// hash 相关
	h                 murmur3.Hash128
	hashedLabels      []HashKey
	hashedStringTable []HashKey
	tmp64             [8]byte
	tmpKey            HashKey

	isBroken bool

	fieldMask fieldMask

	tmpProtoField *protoField
}

// NewFastProfiler 创建新的 FastProfiler
func NewFastProfiler(st []DeltaValueType) Profiler {
	p := &FastProfiler{
		c:   newFastDeltaCaculator(st),
		r:   new(gzip.Reader),
		buf: new(bytes.Buffer),
	}
	p.w = gzip.NewWriter(p.buf)
	return p
}

// Delta 实现 profile 数据的增量计算
func (p *FastProfiler) Delta(data []byte) ([]byte, error) {
	if isGzipData(data) {
		err := p.r.Reset(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		data, err = io.ReadAll(p.r)
		if err != nil {
			return nil, fmt.Errorf("failed to read profile data: %v", err)
		}
	}
	p.buf.Reset()
	p.w.Reset(p.buf)

	if err := p.c.CalculateDeltas(data, p.w); err != nil {
		return nil, fmt.Errorf("failed to caculate profile delta: %w", err)
	}
	if err := p.w.Close(); err != nil {
		return nil, err
	}
	out := make([]byte, len(p.buf.Bytes()))
	copy(out, p.buf.Bytes())
	return out, nil
}

func newFastDeltaCaculator(st []DeltaValueType) *FastDeltaCaculator {
	p := &FastDeltaCaculator{
		h:                murmur3.New128(),
		decoder:          &decoder{},
		enconder:         &encoder{},
		sampleTypesDelta: newValueTypeBytes(st),
		samplesByKey:     map[HashKey]sampleValueTracker{},
		tmpProtoField:    &protoField{},
	}
	return p
}

func (p *FastDeltaCaculator) reset(in []byte, out io.Writer) {
	if p.isBroken {
		p.samplesByKey = map[HashKey]sampleValueTracker{}
		p.currentTimeNanos = 0
		p.durationTimeNanos = 0
	}

	p.decoder.set(in)
	p.enconder.set(out)

	p.hashedLabels = p.hashedLabels[:0]
	p.hashedStringTable = p.hashedStringTable[:0]
	p.locationAddressIdx = p.locationAddressIdx[:0]
	p.validLocationIDs.Clear()

	p.sampleTypesAll = p.sampleTypesAll[:0]
	p.supportDeltaIdx = p.supportDeltaIdx[:0]

	p.validFunctionIDs.Clear()
	p.validStringIdx.Clear()
}

type processor func(f Field) error

func processField(b *buffer, f Field, m *fieldMask, p processor) error {
	for !b.end() {
		if err := decodeField(b); err != nil {
			return err
		}
		if !m.filter(b.f.field) {
			continue
		}
		sub, err := f.decode(b)
		if err != nil {
			return err
		}
		if p == nil {
			continue
		}
		if err := p(sub); err != nil {
			return err
		}
	}
	return nil
}

func (m fieldMask) filter(f protowire.Number) bool {
	if m[0] {
		return true
	}
	return m[int(f)]
}

func (m *fieldMask) set(fields ...protowire.Number) {
	m.reset()
	for _, f := range fields {
		(*m)[int(f)] = true
	}
}

func (m *fieldMask) setAll() {
	m[0] = true
}

func (m *fieldMask) reset() {
	for i := range *m {
		(*m)[i] = false
	}
}

type sampleValueTracker struct {
	oldDeltaValues [maxDeltaSampleValues]int64
	newFullValues  [maxSampleValues]int64
	haveCalculated bool
}

// ValueTypeBytes 用 []byte 表示 ValueType 中的 string
type ValueTypeBytes struct {
	Type []byte
	Unit []byte
}

func newValueTypeBytes(st []DeltaValueType) []ValueTypeBytes {
	var vtb []ValueTypeBytes
	for _, v := range st {
		vtb = append(
			vtb, ValueTypeBytes{
				Type: []byte(v.Type),
				Unit: []byte(v.Unit),
			},
		)
	}
	return vtb
}
