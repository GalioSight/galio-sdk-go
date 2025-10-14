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

//go:build linux
// +build linux

package runtimes

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadProcStat(t *testing.T) {
	tests := []struct {
		name       string
		fileName   string
		assertFunc func(*ProcStat)
		wantErr    bool
	}{
		{
			name:     "正常解析",
			fileName: "./testdata/proc_self_stat.txt",
			assertFunc: func(stat *ProcStat) {
				assert.Equal(t, byte('R'), stat.State)
				assert.Equal(t, 23041, stat.Ppid)
			},
			wantErr: false,
		},
		{
			name:       "异常解析，缺少（）",
			fileName:   "./testdata/proc_self_stat_wrong.txt",
			assertFunc: func(stat *ProcStat) {},
			wantErr:    true,
		},
		{
			name:       "异常解析，数据被截断",
			fileName:   "./testdata/proc_self_stat_short.txt",
			assertFunc: func(stat *ProcStat) {},
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := ReadProcStat(tt.fileName)
				if (err != nil) != tt.wantErr {
					t.Errorf("parseProcStat() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				tt.assertFunc(got)
			},
		)
	}
}

func TestReadMemStat(t *testing.T) {
	tests := []struct {
		name       string
		fileName   string
		assertFunc func(*MemStat)
		wantErr    bool
	}{
		{
			name:     "正常解析",
			fileName: "./testdata/proc_self_status.txt",
			assertFunc: func(stat *MemStat) {
				assert.Equal(t, uint64(122100*1024), stat.VMPeak)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := ReadMemStat(tt.fileName)
				if (err != nil) != tt.wantErr {
					t.Errorf("ReadMemStat() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				tt.assertFunc(got)
			},
		)
	}
}

func TestReadIOStat(t *testing.T) {
	tests := []struct {
		name       string
		fileName   string
		assertFunc func(*IOStat)
		wantErr    bool
	}{
		{
			name:     "正常解析",
			fileName: "./testdata/proc_self_io.txt",
			assertFunc: func(stat *IOStat) {
				assert.Equal(t, int64(8468), stat.Rchar)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := ReadIOStat(tt.fileName)
				if (err != nil) != tt.wantErr {
					t.Errorf("ReadIOStat() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				tt.assertFunc(got)
			},
		)
	}
}

func TestReadK8SStat(t *testing.T) {
	tests := []struct {
		name                string
		memoryUsageFileName string
		memoryLimitFileName string
		want                *K8SStat
		wantErr             bool
	}{
		{
			name:                "正常解析",
			memoryUsageFileName: "./testdata/memory_usage_in_bytes.txt",
			memoryLimitFileName: "./testdata/memory_limit_in_bytes.txt",
			want: &K8SStat{
				MemoryUsageInBytes: 23437545472,
				MemoryLimitInBytes: 9223372036854775807,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := ReadK8SStat(tt.memoryUsageFileName, tt.memoryLimitFileName)
				if (err != nil) != tt.wantErr {
					t.Errorf("ReadK8SStat() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ReadK8SStat() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
