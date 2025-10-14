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
	"encoding/binary"
	"fmt"
	"io"
	"sort"
)

// CalculateDeltas 计算 profile deltas
func (p *FastDeltaCaculator) CalculateDeltas(in []byte, out io.Writer) error {
	p.reset(in, out)
	if err := p.doCalculate(); err != nil {
		p.isBroken = true
		return err
	}
	// 上一轮采集失败了，本轮数据会被作为下一轮计算的基础数据，但是不会上传
	if p.isBroken {
		p.isBroken = false
		return fmt.Errorf("last round is broken, won't upload the data of this round")
	}
	return nil
}

func (p *FastDeltaCaculator) doCalculate() error {
	defer func() error {
		if e := recover(); e != nil {
			return fmt.Errorf("got panic when caclulating deltas : %v", e)
		}
		return nil
	}()

	if err := p.preCheck(); err != nil {
		return err
	}

	if err := p.BuildIndex(); err != nil {
		return err
	}
	if err := p.AggregateSamples(); err != nil {
		return err
	}
	if err := p.CalculateDeltasAndWriteSamples(); err != nil {
		return err
	}
	// 上一轮采集失败了，本轮采集的数据不会上传，只作为基础数据供下一轮计算，
	// 所以不需要后续写操作
	if p.isBroken {
		return nil
	}
	if err := p.WriteOtherProfileFields(); err != nil {
		return err
	}
	if err := p.WriteFunctions(); err != nil {
		return err
	}
	if err := p.WriteStringTable(); err != nil {
		return err
	}
	return nil
}

// BuildIndex 第一轮遍历，遍历所有的 SampleType、Location、StringTable，
// SampleType：记录所有的 types，全部遍历完后，与 sampleTypesDelta 对比，确定 supportDeltaIdx。
// Location: 记录所有 location 到数组 locationAddressIdx，Location ID 为数组下标，address 为数组元素
// StringTable：按顺序哈希所有的 string，并记录到 hashedStringTable 数组，
// 后面需要通过通过 string 的 index 找到对应 string 计算 sample 哈希。
func (p *FastDeltaCaculator) BuildIndex() error {
	processor := func(f Field) error {
		switch t := f.(type) {
		case *SampleType:
			p.addSampleType(t)
		case *Location:
			p.addLocation(t)
		case *StringTable:
			p.addHashString(t)
			if len(t.Value) == 0 {
				p.validStringIdx.Set(0)
			}
		default:
			return fmt.Errorf("field mismatch: %T", f)
		}
		return nil
	}
	p.fieldMask.set(SampleTypeNum, LocationNum, StringTableNum)
	if err := processField(
		newBuffer(p.decoder.raw, p.tmpProtoField), &p.decoder.prof, &p.fieldMask, processor,
	); err != nil {
		return err
	}
	if len(p.sampleTypesAll) > maxSampleValues {
		return fmt.Errorf("too many sample type, max is %d", maxSampleValues)
	}
	return p.buildSupportDeltaIdx()
}

// AggregateSamples 第二轮遍历，遍历所有 Sample，合并相同 sample key（哈希）的 sample 值
func (p *FastDeltaCaculator) AggregateSamples() error {
	processor := func(f Field) error {
		sample, ok := f.(*Sample)
		if !ok {
			return fmt.Errorf("field mismatch: %T", f)
		}
		if err := p.validateSample(sample); err != nil {
			return err
		}
		if err := p.aggregateSampleValues(sample); err != nil {
			return err
		}
		return nil
	}
	p.fieldMask.set(SampleNum)
	if err := processField(
		newBuffer(p.decoder.raw, p.tmpProtoField), &p.decoder.prof, &p.fieldMask, processor,
	); err != nil {
		return err
	}
	return nil
}

// CalculateDeltasAndWriteSamples 第三轮遍历，再次遍历 Sample，通过 Sample Key 找到对应 sample 上一轮采集的值，
// 计算 diff，并更新对应的samplesByKey，并将 diff 后的 sample 进行 encode 并写入结果中。
func (p *FastDeltaCaculator) CalculateDeltasAndWriteSamples() error {
	processor := func(f Field) error {
		sample, ok := f.(*Sample)
		if !ok {
			return fmt.Errorf("field mismatch: %T", f)
		}
		valid, err := p.calculateDeltas(sample)
		if err != nil {
			return err
		}
		if !valid {
			return nil
		}
		for _, locationID := range sample.LocationID {
			p.validLocationIDs.Set(uint32(locationID))
		}
		for _, l := range sample.Label {
			p.validStringIdx.Set(uint32(l.Key))
			p.validStringIdx.Set(uint32(l.Str))
			p.validStringIdx.Set(uint32(l.NumUnit))
		}
		return f.encode(p.enconder)
	}
	p.fieldMask.set(SampleNum)
	if err := processField(
		newBuffer(p.decoder.raw, p.tmpProtoField), &p.decoder.prof, &p.fieldMask, processor,
	); err != nil {
		return err
	}
	return nil
}

// WriteOtherProfileFields 第四轮遍历，遍历其他字段，encode 后写入结果，只记录与之前遍历出的有效 sample 有关联的值
func (p *FastDeltaCaculator) WriteOtherProfileFields() error {
	processor := func(f Field) error {
		switch t := f.(type) {
		case *SampleType:
			p.validStringIdx.Set(uint32(t.Type))
			p.validStringIdx.Set(uint32(t.Unit))
		case *Mapping:
			p.validStringIdx.Set(uint32(t.FileName))
			p.validStringIdx.Set(uint32(t.BuildID))
		case *Location:
			if !p.validLocationIDs.Contains(uint32(t.ID)) {
				return nil
			}
			for _, line := range t.Line {
				p.validFunctionIDs.Set(uint32(line.FunctionID))
			}
		case *DropFrames:
			p.validStringIdx.Set(uint32(t.Value))
		case *KeepFrames:
			p.validStringIdx.Set(uint32(t.Value))
		case *TimeNanos:
			// 首次采集或者上一轮采集有问题
			if p.currentTimeNanos == 0 || t.Value < p.currentTimeNanos {
				p.currentTimeNanos = t.Value
			} else {
				p.durationTimeNanos = t.Value - p.currentTimeNanos
				p.currentTimeNanos = t.Value
			}
			return nil
		case *DurationNanos:
			return nil
		case *PeriodType:
			p.validStringIdx.Set(uint32(t.Type))
			p.validStringIdx.Set(uint32(t.Unit))
		case *Period:
		case *Comment:
			p.validStringIdx.Set(uint32(t.Value))
		case *DefaultSampleType:
			p.validStringIdx.Set(uint32(t.Value))
		default:
			return fmt.Errorf("field mismatch: %T", f)
		}
		return f.encode(p.enconder)
	}
	p.fieldMask.set(
		SampleTypeNum, MappingNum, LocationNum, DropFramesNum, KeepFramesNum, TimeNanosNum,
		DurationNanosNum, PeriodTypeNum, PeriodNum, CommentNum, DefaultSampleTypeNum,
	)
	if err := processField(
		newBuffer(p.decoder.raw, p.tmpProtoField), &p.decoder.prof, &p.fieldMask, processor,
	); err != nil {
		return err
	}

	return p.writeTimeAndDuration()
}

// WriteFunctions 第五轮遍历，将上一轮遍历根据 locations 过滤出的有效 functions 进行 encode 并写入结果
func (p *FastDeltaCaculator) WriteFunctions() error {
	processor := func(f Field) error {
		function, ok := f.(*Function)
		if !ok {
			return fmt.Errorf("field mismatch: %T", f)
		}
		if !p.validFunctionIDs.Contains(uint32(function.ID)) {
			return nil
		}
		p.validStringIdx.Set(uint32(function.Name))
		p.validStringIdx.Set(uint32(function.SystemName))
		p.validStringIdx.Set(uint32(function.FileName))
		return f.encode(p.enconder)
	}
	p.fieldMask.set(FunctionNum)
	if err := processField(
		newBuffer(p.decoder.raw, p.tmpProtoField), &p.decoder.prof, &p.fieldMask, processor,
	); err != nil {
		return err
	}
	return nil
}

// WriteStringTable 第六轮遍历，将前几轮遍历各个字段过滤出的有效 string 进行 encode 并写入结果
func (p *FastDeltaCaculator) WriteStringTable() error {
	strIdx := 0
	processor := func(f Field) error {
		str, ok := f.(*StringTable)
		if !ok {
			return fmt.Errorf("field mismatch: %T", f)
		}
		if !p.validStringIdx.Contains(uint32(strIdx)) {
			str.Value = nil
		}
		strIdx++
		return f.encode(p.enconder)
	}
	p.fieldMask.set(StringTableNum)
	if err := processField(
		newBuffer(p.decoder.raw, p.tmpProtoField), &p.decoder.prof, &p.fieldMask, processor,
	); err != nil {
		return err
	}
	return nil
}

func (p *FastDeltaCaculator) preCheck() error {
	if len(p.sampleTypesDelta) > maxSampleValues {
		return fmt.Errorf("too many delta sample types, max is %d", maxDeltaSampleValues)
	}
	return nil
}

func (p *FastDeltaCaculator) addSampleType(st *SampleType) {
	p.sampleTypesAll = append(p.sampleTypesAll, [2]int{int(st.Type), int(st.Unit)})
}

func (p *FastDeltaCaculator) addLocation(l *Location) {
	for len(p.locationAddressIdx) <= int(l.ID) {
		p.locationAddressIdx = append(p.locationAddressIdx, 0)
	}
	p.locationAddressIdx[l.ID] = l.Address
}

func (p *FastDeltaCaculator) addHashString(st *StringTable) {
	p.hashBytes(st.Value)
	p.hashedStringTable = append(p.hashedStringTable, p.tmpKey)
}

func (p *FastDeltaCaculator) buildSupportDeltaIdx() error {
	var (
		t   HashKey
		u   HashKey
		err error
	)
	for len(p.supportDeltaIdx) < len(p.sampleTypesAll) {
		p.supportDeltaIdx = append(p.supportDeltaIdx, false)
	}
	for _, dst := range p.sampleTypesDelta {
		if t, err = p.hashBytes(dst.Type); err != nil {
			return err
		}
		if u, err = p.hashBytes(dst.Unit); err != nil {
			return err
		}
		for i, st := range p.sampleTypesAll {
			if t == p.hashedStringTable[st[0]] &&
				u == p.hashedStringTable[st[1]] {
				p.supportDeltaIdx[i] = true
				break
			}
		}
	}
	return nil
}

func (p *FastDeltaCaculator) validateSample(s *Sample) error {
	numOfStrings := int64(len(p.hashedStringTable))
	for _, l := range s.Label {
		if numOfStrings <= l.Key {
			return fmt.Errorf("invalid Label Key index %d", l.Key)
		}
		if numOfStrings <= l.Str {
			return fmt.Errorf("invalid Label Str index %d", l.Str)
		}
		if numOfStrings < l.NumUnit {
			return fmt.Errorf("invalid Lable NumUnit index %d", l.NumUnit)
		}
	}
	return nil
}

func (p *FastDeltaCaculator) aggregateSampleValues(s *Sample) error {
	sk, err := p.sampleKey(s)
	if err != nil {
		return err
	}
	var sample sampleValueTracker
	prev := p.samplesByKey[sk]
	sample.oldDeltaValues = prev.oldDeltaValues
	if len(s.Value) > maxSampleValues || len(sample.newFullValues) > maxSampleValues {
		return fmt.Errorf("too many sample values, max is %d", maxSampleValues)
	}
	for i, v := range s.Value {
		sample.newFullValues[i] = sample.newFullValues[i] + v
	}
	p.samplesByKey[sk] = sample
	return nil
}

func (p *FastDeltaCaculator) calculateDeltas(s *Sample) (bool, error) {
	sk, err := p.sampleKey(s)
	if err != nil {
		return false, err
	}
	st, ok := p.samplesByKey[sk]
	if !ok {
		return false, fmt.Errorf("sample not found by key %v", sk)
	}
	// 因为同一个 sampleKey 可能有多个 sample，避免重复计算
	if st.haveCalculated {
		return false, nil
	}

	zeroValueCount := 0
	j := 0
	for i := range s.Value {
		if p.supportDeltaIdx[i] {
			// 计算 delta value
			s.Value[i] = st.newFullValues[i] - st.oldDeltaValues[j]
			st.oldDeltaValues[j] = st.newFullValues[i]
			j++
		} else {
			s.Value[i] = st.newFullValues[i]
		}
		if s.Value[i] == 0 {
			zeroValueCount++
		}
	}
	// 全是 0 值
	if len(s.Value) == zeroValueCount {
		return false, nil
	}

	st.haveCalculated = true
	// 清零 newFullValues
	for i := range st.newFullValues {
		st.newFullValues[i] = 0
	}
	p.samplesByKey[sk] = st
	return true, nil
}

func (p *FastDeltaCaculator) sampleKey(s *Sample) (HashKey, error) {
	p.hashedLabels = p.hashedLabels[:0]
	if err := p.hashLabel(s); err != nil {
		return HashKey{}, err
	}
	if len(p.hashedLabels) > 1 {
		sort.Slice(
			p.hashedLabels, func(i, j int) bool {
				return bytes.Compare(p.hashedLabels[i][:], p.hashedLabels[j][:]) == -1
			},
		)
	}

	p.h.Reset()
	// write location to hash
	for _, id := range s.LocationID {
		if id >= uint64(len(p.locationAddressIdx)) {
			return HashKey{}, fmt.Errorf("location ID %d not found in profiler", id)
		}
		// get location.Address
		binary.LittleEndian.PutUint64(p.tmp64[:], p.locationAddressIdx[id])
		p.h.Write(p.tmp64[:8])
	}
	// write hashed label to hash
	for _, l := range p.hashedLabels {
		copy(p.tmpKey[:], l[:])
		p.h.Write(p.tmpKey[:])
	}
	p.h.Sum(p.tmpKey[:0])

	return p.tmpKey, nil
}

func (p *FastDeltaCaculator) hashBytes(b []byte) (HashKey, error) {
	p.h.Reset()
	if _, err := p.h.Write(b); err != nil {
		return p.tmpKey, err
	}
	p.h.Sum(p.tmpKey[:0])
	return p.tmpKey, nil
}

func (p *FastDeltaCaculator) hashLabel(s *Sample) error {
	p.h.Reset()
	for _, l := range s.Label {
		if _, err := p.h.Write(p.hashedStringTable[l.Key][:]); err != nil {
			return err
		}
		if _, err := p.h.Write(p.hashedStringTable[l.NumUnit][:]); err != nil {
			return err
		}
		binary.BigEndian.PutUint64(p.tmp64[:], uint64(l.Num))
		if _, err := p.h.Write(p.tmp64[:8]); err != nil {
			return err
		}
		if _, err := p.h.Write(p.hashedStringTable[l.Str][:]); err != nil {
			return err
		}
		p.h.Sum(p.tmpKey[:0])
		p.hashedLabels = append(p.hashedLabels, p.tmpKey)
	}
	return nil
}

func (p FastDeltaCaculator) writeTimeAndDuration() error {
	if p.currentTimeNanos != 0 {
		f := &TimeNanos{Value: p.currentTimeNanos}
		if err := f.encode(p.enconder); err != nil {
			return err
		}
	}
	if p.durationTimeNanos != 0 {
		f := &DurationNanos{Value: p.durationTimeNanos}
		if err := f.encode(p.enconder); err != nil {
			return err
		}
	}
	return nil
}
