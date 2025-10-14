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

package model

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomName(t *testing.T) {
	type args struct {
		group       string
		name        string
		aggregation Aggregation
		check       bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "none",
			args: args{
				group:       "test",
				name:        "a",
				aggregation: Aggregation_AGGREGATION_NONE,
				check:       true,
			},
			want: "custom_gauge_test_a_set",
		},
		{
			name: "set",
			args: args{
				group:       "test",
				name:        "a",
				aggregation: Aggregation_AGGREGATION_SET,
				check:       true,
			},
			want: "custom_gauge_test_a_set",
		},
		{
			name: "sum",
			args: args{
				group:       "test",
				name:        "a",
				aggregation: Aggregation_AGGREGATION_SUM,
				check:       true,
			},
			want: "custom_counter_test_a_total",
		},
		{
			name: "avg",
			args: args{
				group:       "test",
				name:        "a",
				aggregation: Aggregation_AGGREGATION_AVG,
				check:       true,
			},
			want: "custom_counter_test_a",
		},
		{
			name: "max",
			args: args{
				group:       "test",
				name:        "a",
				aggregation: Aggregation_AGGREGATION_MAX,
				check:       true,
			},
			want: "custom_gauge_test_a_max",
		},
		{
			name: "min",
			args: args{
				group:       "test",
				name:        "a",
				aggregation: Aggregation_AGGREGATION_MIN,
				check:       true,
			},
			want: "custom_gauge_test_a_min",
		},
		{
			name: "histogram",
			args: args{
				group:       "test",
				name:        "a",
				aggregation: Aggregation_AGGREGATION_HISTOGRAM,
				check:       true,
			},
			want: "custom_histogram_test_a",
		},
		{
			name: "counter",
			args: args{
				group:       "test",
				name:        "a",
				aggregation: Aggregation_AGGREGATION_COUNTER,
				check:       true,
			},
			want: "custom_counter_test_a_total",
		},
		{
			name: "errorGroup",
			args: args{
				group:       "test-中国",
				name:        "a",
				aggregation: Aggregation_AGGREGATION_COUNTER,
				check:       true,
			},
			want: "custom_counter_base629uZ5tiL5tQ3clRH_a_total",
		},
		{
			name: "errorName",
			args: args{
				group:       "test",
				name:        "a-中国",
				aggregation: Aggregation_AGGREGATION_COUNTER,
				check:       true,
			},
			want: "custom_counter_test_base629uZ5tiL5tEG_total",
		}, {
			name: "groupEmpty",
			args: args{
				group:       "",
				name:        "a",
				aggregation: Aggregation_AGGREGATION_SET,
				check:       true,
			},
			want: "custom_gauge_default_a_set",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := CustomName(tt.args.group, tt.args.name, tt.args.aggregation)
				if got != tt.want {
					t.Errorf("CustomName() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

// BenchmarkCustomName CustomName 性能测试。
// BenchmarkNewMetricSchema-10    	 3333951	       355.0 ns/op
func BenchmarkCustomName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CustomName("中文监控项", "根目录的磁盘使用率", Aggregation_AGGREGATION_COUNTER)
	}
}

// BenchmarkNewMetricSchema 创建元数据性能测试。
// BenchmarkNewMetricSchema-10        	 3368247	       351.0 ns/op
func BenchmarkNewMetricSchema(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewMetricSchema("中文监控项", "根目录的磁盘使用率", Aggregation_AGGREGATION_COUNTER)
	}
}

// BenchmarkGetMetricSchema 从获取元数据
// BenchmarkGetMetricSchema-10        	71901318	        17.37 ns/op
func BenchmarkGetMetricSchema(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetMetricSchema("中文监控项", "根目录的磁盘使用率", Aggregation_AGGREGATION_COUNTER)
	}
}

// BenchmarkNameToIdentifier name to id 转换性能测试
// BenchmarkNameToIdentifier-10    	 5449281	       213.3 ns/op
func BenchmarkNameToIdentifier(b *testing.B) {
	name := "测试名字 - 中文监控项 - 目录的磁盘使用率"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NameToIdentifier(name)
	}
}

// BenchmarkIdentifierToName id to name 转换性能测试
// BenchmarkIdentifierToName-10    	 9221800	       127.0 ns/op
func BenchmarkIdentifierToName(b *testing.B) {
	id := "base62H645oS55ftXyxczzDc0zJUzzrs3yddzzBpFQyN00P1RzjczzP0SzbFXyBpFQus1ybEyyr8V0XsWzB"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IdentifierToName(id)
	}
}

func TestToValidName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"trpc.DiskUsed(GB)",
			args{
				"trpc.DiskUsed(GB)",
			},
			"trpc_DiskUsed_GB_",
		},
		{
			"trpc.DiskUsedFraction(%)",
			args{
				"trpc.DiskUsedFraction(%)",
			},
			"trpc_DiskUsedFraction___",
		},
		{
			"abc",
			args{
				"abc",
			},
			"abc",
		},
		{
			"我是中国人，I Love China!",
			args{
				"我是中国人，I Love China!",
			},
			"______I_Love_China_",
		},
		{
			"空字符串",
			args{
				"",
			},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := ToValidName(tt.args.name); got != tt.want {
					t.Errorf("ToValidName() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestNameToIdentifier(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"empty",
			args{
				"",
			},
			"",
		},
		{
			"first",
			args{
				"first",
			},
			"first",
		},
		{
			"first_second",
			args{
				"first_second",
			},
			"first_second",
		},
		{
			"中文",
			args{
				"中文",
			},
			"base62Hap5tiL5",
		},

		{
			"特殊符号",
			args{
				"(%)",
			},
			"base62pUCK",
		},

		{
			"中文 + 特殊符号",
			args{
				"cpu 使用率 (%)",
			},
			"base62pUCKgcojnjKlnf3eSOI1B3Y",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				for i := 0; i < 1000; i++ {
					identifier := NameToIdentifier(tt.args.name)
					runtime.GC()
					assert.Equalf(t, tt.want, identifier, "NameToIdentifier(%v)", tt.args.name)
				}
			},
		)
	}
}

func TestIdentifierToName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"empty",
			args{
				"",
			},
			"",
		},
		{
			"first",
			args{
				"first",
			},
			"first",
		},
		{
			"first_second",
			args{
				"first_second",
			},
			"first_second",
		},
		{
			"中文",
			args{
				"base62Hap5tiL5",
			},
			"中文",
		},

		{
			"特殊符号",
			args{
				"base62pUCK",
			},
			"(%)",
		},
		{
			"中文 + 特殊符号",
			args{
				"base62pUCKgcojnjKlnf3eSOI1B3Y",
			},
			"cpu 使用率 (%)",
		},
		{
			"解码错误",
			args{
				"base62 中文",
			},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				for i := 0; i < 1000; i++ {
					name := IdentifierToName(tt.args.name)
					runtime.GC()
					assert.Equalf(t, tt.want, name, "NameToIdentifier(%v)", tt.args.name)
				}
			},
		)
	}
}

func TestParseCustomName(t *testing.T) {
	type args struct {
		customName string
	}
	tests := []struct {
		name    string
		args    args
		want    MetricSchema
		wantErr assert.ErrorAssertionFunc
	}{
		{
			"gauge_test_a_set",
			args{
				"gauge_test_a_set",
			},
			MetricSchema{},
			func(t assert.TestingT, err error, i ...interface{}) bool {
				return err != nil
			},
		},
		{
			"custom_gauge_test",
			args{
				"custom_gauge_test",
			},
			MetricSchema{},
			func(t assert.TestingT, err error, i ...interface{}) bool {
				return err != nil
			},
		},
		{
			name: "custom_gauge_test_a_set",
			args: args{
				"custom_gauge_test_a_set",
			},
			want: MetricSchema{
				MetricType:  "gauge",
				MonitorName: "test",
				MetricName:  "a",
				Aggregation: "set",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err == nil
			},
		},
		{
			name: "custom_gauge_a_set",
			args: args{
				"custom_gauge_a_set",
			},
			want: MetricSchema{
				MetricType:  "gauge",
				MetricName:  "a",
				Aggregation: "set",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err == nil
			},
		},
		{
			name: "custom_counter_base629uZ5tiL5tQ3clRH_a_total",
			args: args{
				"custom_counter_base629uZ5tiL5tQ3clRH_a_total",
			},
			want: MetricSchema{
				MetricType:  "counter",
				MonitorName: "test-中国",
				MetricName:  "a",
				Aggregation: "total",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err == nil
			},
		},
		{
			name: "custom_counter_test_base629uZ5tiL5tEG_total",
			args: args{
				"custom_counter_test_base629uZ5tiL5tEG_total",
			},
			want: MetricSchema{
				MetricType:  "counter",
				MonitorName: "test",
				MetricName:  "a-中国",
				Aggregation: "total",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err == nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := ParseCustomName(tt.args.customName)
				if !tt.wantErr(t, err, fmt.Sprintf("ParseCustomName(%v)", tt.args.customName)) {
					return
				}
				assert.Equalf(t, tt.want, got, "ParseCustomName(%v)", tt.args.customName)
			},
		)
	}
}
