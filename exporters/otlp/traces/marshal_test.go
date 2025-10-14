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

package traces

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"galiosight.ai/galio-sdk-go/model"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

var (
	testData = &model.Metric{
		Name:        strings.Repeat("a", 1000),
		Value:       0,
		Aggregation: model.Aggregation_AGGREGATION_AVG,
	}
)

func BenchmarkProtoMarshalTextString(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		proto.Marshal(testData)
	}
}

func BenchmarkProtoMarshalCompactTextString(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = testData.String()
	}
}

func BenchmarkProtoMarshalPb(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(testData)
	}
}

// BenchmarkProtoSize-12    	 7018797	       165 ns/op	      16 B/op	       1 allocs/op
func BenchmarkProtoSize(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = proto.Size(testData)
	}
}

// 测试 DefaultMarshalToString 函数在非 UTF-8 字符下 SetSonicFastest(false) 和 SetSonicFastest(true) 的表现
func TestDefaultMarshalToString_ProtoInvalidUTF8(t *testing.T) {
	// 测试 DefaultMarshalToString 函数
	tests := []struct {
		name    string
		message interface{}
		want    string
	}{
		{
			name: "proto message",
			message: &model.Metric{
				Name:        string([]byte{0xff, 0xfe, 0xfd}),
				Value:       0,
				Aggregation: model.Aggregation_AGGREGATION_AVG,
			},
			want: `{"name":"\ufffd\ufffd\ufffd","aggregation":3}`,
		},
	}
	SetSonicFastest(false)
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := DefaultMarshalToString(tt.message)
				assert.True(t, utf8.ValidString(got))
				assert.Equal(t, tt.want, got)
			},
		)
	}
	SetSonicFastest(true)
}

// 测试 DefaultMarshalToString 函数
func TestDefaultMarshalToString_NonProto(t *testing.T) {
	type Person struct {
		Name    string
		Age     int
		Friends []*Person
	}

	// 创建一个循环引用的结构体
	p1 := &Person{Name: strings.Repeat("Alice", 10), Age: 20}
	p2 := &Person{Name: "Bob", Age: 21}
	p1.Friends = []*Person{p2}
	p2.Friends = []*Person{p1}

	// 测试 DefaultMarshalToString 函数
	tests := []struct {
		name    string
		message interface{}
	}{
		{
			name:    "non-proto message",
			message: p1,
		},
		{
			name:    "non-proto message",
			message: p2,
		},
		{
			name:    "non-proto message",
			message: p1,
		},
	}
	a := assert.New(t)

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := DefaultMarshalToString(tt.message)
				want := fmt.Sprintf("%+v", tt.message)
				a.Equal(want, got)
			},
		)
	}
}
