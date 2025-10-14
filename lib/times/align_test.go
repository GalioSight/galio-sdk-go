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

package times

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAlign(t *testing.T) {
	type args struct {
		time  int64
		width int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			"1646880803",
			args{
				1646880803,
				10,
			},
			1646880800,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := Align(tt.args.time, tt.args.width); got != tt.want {
					t.Errorf("Align() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestWaitAlign(t *testing.T) {
	type args struct {
		width int64
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"1",
			args{
				1,
			},
		},
		{
			"3",
			args{
				3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				WaitAlign(tt.args.width)
				assert.Equal(t, int64(0), time.Now().Unix()%tt.args.width)
			},
		)
	}
}

func TestWaitAlignArgs(t *testing.T) {
	start := time.Now()
	WaitAlign(-1)
	assert.Greater(t, time.Second, time.Since(start))
}

func TestWaitRandomDuration(t *testing.T) {
	type args struct {
		d time.Duration
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"time.Second",
			args{
				time.Second,
			},
		},
		{
			"3time.Second",
			args{
				3 * time.Second,
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				start := time.Now()
				WaitRandomDuration(tt.args.d)
				assert.Greater(t, tt.args.d+time.Second, time.Since(start))
			},
		)
	}
}

func TestRandomDuration(t *testing.T) {
	type args struct {
		d time.Duration
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"time.Second",
			args{
				time.Second,
			},
		},
		{
			"3time.Second",
			args{
				3 * time.Second,
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := RandomDuration(tt.args.d)
				assert.Less(t, time.Duration(0), got)
				assert.Less(t, got, tt.args.d)
			},
		)
	}
}
