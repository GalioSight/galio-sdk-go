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

package timedmax

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMaxValueTracker(t *testing.T) {
	m := NewMaxValueTracker(time.Second, 0)
	tests := []struct {
		updateValue float64
		needUpdate  bool
		updateRet   float64
		wait        time.Duration
		expectedMax float64
	}{
		{updateValue: 10, needUpdate: true, updateRet: 10, wait: 0, expectedMax: 10},
		{updateValue: 5, needUpdate: true, updateRet: 10, wait: 0, expectedMax: 10},
		{updateValue: 20, needUpdate: true, updateRet: 20, wait: 0, expectedMax: 20},
		// 900ms，还在窗口内 20
		{updateValue: 15, needUpdate: true, updateRet: 20, wait: 900 * time.Millisecond, expectedMax: 20},
		// 1s，窗口外，拿到的是 0
		{needUpdate: false, wait: time.Second, expectedMax: 0},
		// 新窗口
		{updateValue: 10, needUpdate: true, updateRet: 10, wait: 0, expectedMax: 10},
		{updateValue: 30, needUpdate: true, updateRet: 30, wait: 0, expectedMax: 30},
	}
	for i, tt := range tests {
		if tt.needUpdate {
			require.Equalf(t, tt.updateRet, m.Update(tt.updateValue), "Test case %d", i+1)
		}
		time.Sleep(tt.wait)
		require.Equalf(t, tt.expectedMax, m.Get(), "Test case %d", i+1)
	}
}
