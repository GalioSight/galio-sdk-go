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
	"errors"
	"fmt"

	"google.golang.org/protobuf/encoding/protowire"
)

var ErrFieldNotFound = errors.New("field not found")

// field numbers for SimpleProfile
const (
	SampleTypeNum protowire.Number = iota + 1
	SampleNum
	MappingNum
	LocationNum
	FunctionNum
	StringTableNum
	DropFramesNum
	KeepFramesNum
	TimeNanosNum
	DurationNanosNum
	PeriodTypeNum
	PeriodNum
	CommentNum
	DefaultSampleTypeNum
)

// Field 定义 Profile 的每个字段编解码时，需要的 decode 和 encode 方法
type Field interface {
	decode(b *buffer) (Field, error)
	encode(e *encoder) error
}

// Message 定义 Profile 各字段编解码子字段时需实现的方法
type Message interface {
	encodeInternal(e *encoder)
}

// SimpleProfile 参考 pprof 定义 Profile 结构体，用于 profile 数据解析
// 参考：https://github.com/google/pprof/blob/main/proto/profile.proto
type SimpleProfile struct {
	SampleType        SampleType
	Sample            Sample
	Mapping           Mapping
	Location          Location
	Function          Function
	StringTable       StringTable
	DropFrames        DropFrames
	KeepFrames        KeepFrames
	TimeNanos         TimeNanos
	DurationNanos     DurationNanos
	PeriodType        PeriodType
	Period            Period
	Comment           Comment
	DefaultSampleType DefaultSampleType
}

func (f *SimpleProfile) decode(b *buffer) (Field, error) {
	switch b.f.field {
	case 1: // SampleType
		return f.SampleType.decode(newBuffer(b.f.data, b.f))
	case 2: // Sample
		return f.Sample.decode(newBuffer(b.f.data, b.f))
	case 3: // Mapping
		return f.Mapping.decode(newBuffer(b.f.data, b.f))
	case 4: // Location
		return f.Location.decode(newBuffer(b.f.data, b.f))
	case 5: // Function
		return f.Function.decode(newBuffer(b.f.data, b.f))
	case 6: // StringTable
		return f.StringTable.decode(b)
	case 7: // DropFrames
		return f.DropFrames.decode(b)
	case 8: // KeepFrames
		return f.KeepFrames.decode(b)
	case 9: // TimeNanos
		return f.TimeNanos.decode(b)
	case 10: // DurationNanos
		return f.DurationNanos.decode(b)
	case 11: // PeriodType
		return f.PeriodType.decode(newBuffer(b.f.data, b.f))
	case 12: // Period
		return f.Period.decode(b)
	case 13: // Comment
		return f.Comment.decode(b)
	case 14: // DefaultSampleType
		return f.DefaultSampleType.decode(b)
	default:
		return nil, ErrFieldNotFound
	}
}

func (f *SimpleProfile) encode(e *encoder) error {
	return fmt.Errorf("encode not implememted for SimpleProfile")
}

// SampleType 对应 profile.proto 中的 SampleType 的一个值
type SampleType struct {
	ValueType
}

func (f *SampleType) decode(b *buffer) (Field, error) {
	if _, err := f.ValueType.decode(b); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *SampleType) encode(e *encoder) error {
	e.reset()
	e.encodeField(int(SampleTypeNum), f)
	return e.writeOut(e.buf)
}

func (f *SampleType) encodeInternal(e *encoder) {
	f.ValueType.encodeInternal(e)
}

// Sample 对应 profile.proto 中的 Sample
type Sample struct {
	LocationID []uint64
	Value      []int64
	Label      []Label
}

func (f *Sample) decode(b *buffer) (Field, error) {
	*f = Sample{LocationID: f.LocationID[:0], Value: f.Value[:0], Label: f.Label[:0]}
	for !b.end() {
		if err := decodeField(b); err != nil {
			return nil, err
		}
		switch b.f.field {
		case 1: // LocationID
			if err := decodeUint64s(b, &f.LocationID); err != nil {
				return nil, err
			}
		case 2: // Value
			if err := decodeInt64s(b, &f.Value); err != nil {
				return nil, err
			}
		case 3: // Label
			f.Label = append(f.Label, Label{})
			if _, err := f.Label[len(f.Label)-1].decode(newBuffer(b.f.data, b.f)); err != nil {
				return nil, err
			}
		default:
			return nil, ErrFieldNotFound
		}
	}
	return f, nil
}

func (f *Sample) encode(e *encoder) error {
	e.reset()
	e.encodeField(int(SampleNum), f)
	return e.writeOut(e.buf)
}

func (f *Sample) encodeInternal(e *encoder) {
	e.encodeUint64s(1, f.LocationID)
	e.encodeInt64s(2, f.Value)
	for i := range f.Label {
		e.encodeField(3, &f.Label[i])
	}
}

// Mapping 对应 profile.proto 中的 Mapping
type Mapping struct {
	ID              uint64
	MemoryStart     uint64
	MemoryLimit     uint64
	FileOffset      uint64
	FileName        int64
	BuildID         int64
	HasFunctions    bool
	HasFilenames    bool
	HasLineNumbers  bool
	HasInlineFrames bool
}

func (f *Mapping) decode(b *buffer) (Field, error) {
	*f = Mapping{}
	for !b.end() {
		if err := decodeField(b); err != nil {
			return nil, err
		}
		switch b.f.field {
		case 1: // ID
			f.ID = b.f.u64
		case 2: // MemoryStart
			f.MemoryStart = b.f.u64
		case 3: // MemoryLimit
			f.MemoryLimit = b.f.u64
		case 4: // FileOffset
			f.FileOffset = b.f.u64
		case 5: // FileName
			f.FileName = int64(b.f.u64)
		case 6: // BuildID
			f.BuildID = int64(b.f.u64)
		case 7: // HasFunctions
			f.HasFunctions = b.f.u64 != 0
		case 8: // HasFilenames
			f.HasFilenames = b.f.u64 != 0
		case 9: // HasLineNumbers
			f.HasLineNumbers = b.f.u64 != 0
		case 10: // HasInlineFrames
			f.HasInlineFrames = b.f.u64 != 0
		default:
			return nil, ErrFieldNotFound
		}
	}
	return f, nil
}

func (f *Mapping) encode(e *encoder) error {
	e.reset()
	e.encodeField(int(MappingNum), f)
	return e.writeOut(e.buf)
}

func (f *Mapping) encodeInternal(e *encoder) {
	e.encodeUint64Nonzero(1, f.ID)
	e.encodeUint64Nonzero(2, f.MemoryStart)
	e.encodeUint64Nonzero(3, f.MemoryLimit)
	e.encodeUint64Nonzero(4, f.FileOffset)
	e.encodeInt64Nonzero(5, f.FileName)
	e.encodeInt64Nonzero(6, f.BuildID)
	e.encodeBoolTrue(7, f.HasFunctions)
	e.encodeBoolTrue(8, f.HasFilenames)
	e.encodeBoolTrue(9, f.HasLineNumbers)
	e.encodeBoolTrue(10, f.HasInlineFrames)
}

// Location 对应 profile.proto 中的 Location
type Location struct {
	ID        uint64
	MappingID uint64
	Address   uint64
	Line      []Line
	IsFolded  bool
}

func (f *Location) decode(b *buffer) (Field, error) {
	*f = Location{Line: f.Line[:0]}
	for !b.end() {
		if err := decodeField(b); err != nil {
			return nil, err
		}
		switch b.f.field {
		case 1: // ID
			f.ID = b.f.u64
		case 2: // MappingID
			f.MappingID = b.f.u64
		case 3: // Address
			f.Address = b.f.u64
		case 4: // Line
			f.Line = append(f.Line, Line{})
			if _, err := f.Line[len(f.Line)-1].decode(newBuffer(b.f.data, b.f)); err != nil {
				return nil, err
			}
		case 5: // IsFolded
			f.IsFolded = b.f.u64 != 0
		default:
			return nil, ErrFieldNotFound
		}
	}
	return f, nil
}

func (f *Location) encode(e *encoder) error {
	e.reset()
	e.encodeField(int(LocationNum), f)
	return e.writeOut(e.buf)
}

func (f *Location) encodeInternal(e *encoder) {
	e.encodeUint64Nonzero(1, f.ID)
	e.encodeUint64Nonzero(2, f.MappingID)
	e.encodeUint64Nonzero(3, f.Address)
	for i := range f.Line {
		e.encodeField(4, &f.Line[i])
	}
	e.encodeBoolTrue(5, f.IsFolded)
}

// Function 对应 profile.proto 中的 Function
type Function struct {
	ID         uint64
	Name       int64
	SystemName int64
	FileName   int64
	StartLine  int64
}

func (f *Function) decode(b *buffer) (Field, error) {
	*f = Function{}
	for !b.end() {
		if err := decodeField(b); err != nil {
			return nil, err
		}
		switch b.f.field {
		case 1: // ID
			f.ID = b.f.u64
		case 2: // Name
			f.Name = int64(b.f.u64)
		case 3: // SystemName
			f.SystemName = int64(b.f.u64)
		case 4: // FileName
			f.FileName = int64(b.f.u64)
		case 5: // StartLine
			f.StartLine = int64(b.f.u64)
		default:
			return nil, ErrFieldNotFound
		}
	}
	return f, nil
}

func (f *Function) encode(e *encoder) error {
	e.reset()
	e.encodeField(int(FunctionNum), f)
	return e.writeOut(e.buf)
}

func (f *Function) encodeInternal(e *encoder) {
	e.encodeUint64Nonzero(1, f.ID)
	e.encodeInt64Nonzero(2, f.Name)
	e.encodeInt64Nonzero(3, f.SystemName)
	e.encodeInt64Nonzero(4, f.FileName)
	e.encodeInt64Nonzero(5, f.StartLine)
}

// StringTable 对应 profile.proto 中的 StringTable
type StringTable struct {
	Value []byte
}

func (f *StringTable) decode(b *buffer) (Field, error) {
	f.Value = b.f.data
	return f, nil
}

func (f *StringTable) encode(e *encoder) error {
	e.reset()
	e.encodeBytes(int(StringTableNum), f.Value)
	return e.writeOut(e.buf)
}

// DropFrames 对应 profile.proto 中的 dropFrameX
type DropFrames struct {
	Value int64
}

func (f *DropFrames) decode(b *buffer) (Field, error) {
	f.Value = int64(b.f.u64)
	return f, nil
}

func (f *DropFrames) encode(e *encoder) error {
	e.reset()
	e.encodeInt64Nonzero(int(DropFramesNum), f.Value)
	return e.writeOut(e.buf)
}

// KeepFrames 对应 profile.proto 中的 KeepFrames
type KeepFrames struct {
	Value int64
}

func (f *KeepFrames) decode(b *buffer) (Field, error) {
	f.Value = int64(b.f.u64)
	return f, nil
}

func (f *KeepFrames) encode(e *encoder) error {
	e.reset()
	e.encodeInt64Nonzero(int(KeepFramesNum), f.Value)
	return e.writeOut(e.buf)
}

// TimeNanos 对应 profile.proto 中的 TimeNanos
type TimeNanos struct {
	Value int64
}

func (f *TimeNanos) decode(b *buffer) (Field, error) {
	f.Value = int64(b.f.u64)
	return f, nil
}

func (f *TimeNanos) encode(e *encoder) error {
	e.reset()
	e.encodeInt64Nonzero(int(TimeNanosNum), f.Value)
	return e.writeOut(e.buf)
}

// DurationNanos 对应 profile.proto 中的 DurationNanos
type DurationNanos struct {
	Value int64
}

func (f *DurationNanos) decode(b *buffer) (Field, error) {
	f.Value = int64(b.f.u64)
	return f, nil
}

func (f *DurationNanos) encode(e *encoder) error {
	e.reset()
	e.encodeInt64Nonzero(int(DurationNanosNum), f.Value)
	return e.writeOut(e.buf)
}

// PeriodType 对应 profile.proto 中的 PeriodType
type PeriodType struct {
	ValueType
}

func (f *PeriodType) decode(b *buffer) (Field, error) {
	if _, err := f.ValueType.decode(b); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *PeriodType) encode(e *encoder) error {
	e.reset()
	e.encodeField(int(PeriodTypeNum), f)
	return e.writeOut(e.buf)
}

func (f *PeriodType) encodeInternal(e *encoder) {
	f.ValueType.encodeInternal(e)
}

// Period 对应 profile.proto 中的 Period
type Period struct {
	Value int64
}

func (f Period) num() int { return 12 }

func (f *Period) decode(b *buffer) (Field, error) {
	f.Value = int64(b.f.u64)
	return f, nil
}

func (f *Period) encode(e *encoder) error {
	e.reset()
	e.encodeInt64Nonzero(int(PeriodNum), f.Value)
	return e.writeOut(e.buf)
}

// Comment 对应 profile.proto 中的 commentX
type Comment struct {
	Value int64
}

func (f *Comment) decode(b *buffer) (Field, error) {
	f.Value = int64(b.f.u64)
	return f, nil
}

func (f *Comment) encode(e *encoder) error {
	e.reset()
	e.encodeInt64Nonzero(int(CommentNum), f.Value)
	return e.writeOut(e.buf)
}

// DefaultSampleType 对应 profile.proto 中的 DefaultSampleType
type DefaultSampleType struct {
	Value int64
}

func (f *DefaultSampleType) decode(b *buffer) (Field, error) {
	f.Value = int64(b.f.u64)
	return f, nil
}

func (f *DefaultSampleType) encode(e *encoder) error {
	e.reset()
	e.encodeInt64Nonzero(int(DefaultSampleTypeNum), f.Value)
	return e.writeOut(e.buf)
}

// ValueType 对应的 profile.proto 中的 ValueType
type ValueType struct {
	Type int64
	Unit int64
}

func (f *ValueType) decode(b *buffer) (Field, error) {
	*f = ValueType{}
	for !b.end() {
		if err := decodeField(b); err != nil {
			return nil, err
		}
		switch b.f.field {
		case 1: // Type
			f.Type = int64(b.f.u64)
		case 2: // Unit
			f.Unit = int64(b.f.u64)
		default:
			return nil, ErrFieldNotFound
		}
	}
	return f, nil
}

func (f *ValueType) encode(e *encoder) error {
	return nil
}

func (f *ValueType) encodeInternal(e *encoder) {
	e.encodeInt64Nonzero(1, f.Type)
	e.encodeInt64Nonzero(2, f.Unit)
}

// Label 对应 profile.proto 中的 Label
type Label struct {
	Key     int64
	Str     int64
	Num     int64
	NumUnit int64
}

func (f *Label) decode(b *buffer) (Field, error) {
	*f = Label{}
	for !b.end() {
		if err := decodeField(b); err != nil {
			return nil, err
		}
		switch b.f.field {
		case 1: // Key
			f.Key = int64(b.f.u64)
		case 2: // Str
			f.Str = int64(b.f.u64)
		case 3: // Num
			f.Num = int64(b.f.u64)
		case 4: // NumUint
			f.NumUnit = int64(b.f.u64)
		default:
			return nil, ErrFieldNotFound
		}
	}
	return f, nil
}

func (f *Label) encode(e *encoder) error {
	return nil
}

func (f *Label) encodeInternal(e *encoder) {
	e.encodeInt64Nonzero(1, f.Key)
	e.encodeInt64Nonzero(2, f.Str)
	e.encodeInt64Nonzero(3, f.Num)
	e.encodeInt64Nonzero(4, f.NumUnit)
}

// Line 对应 profile.proto 中的 Line
type Line struct {
	FunctionID uint64
	Line       int64
}

func (f *Line) decode(b *buffer) (Field, error) {
	*f = Line{}
	for !b.end() {
		if err := decodeField(b); err != nil {
			return nil, err
		}
		switch b.f.field {
		case 1: // FunctionID
			f.FunctionID = b.f.u64
		case 2: // Line
			f.Line = int64(b.f.u64)
		default:
			return nil, ErrFieldNotFound
		}
	}
	return f, nil
}

func (f *Line) encode(e *encoder) error {
	return nil
}

func (f *Line) encodeInternal(e *encoder) {
	e.encodeUint64Nonzero(1, f.FunctionID)
	e.encodeInt64Nonzero(2, f.Line)
}
