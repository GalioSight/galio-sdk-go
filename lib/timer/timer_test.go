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

package timer

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSafeTimer(t *testing.T) {
	const numRoutines = 10
	const numIterations = 100

	safeTimer := NewSafeTimer(time.Millisecond * 1)
	wg := sync.WaitGroup{}
	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				time.Sleep(time.Millisecond * 1)
				start := time.Now()
				safeTimer.Reset(time.Millisecond * 1)
				select {
				case <-safeTimer.C():
					assert.Greater(t, time.Since(start), time.Microsecond)
				case <-time.After(time.Millisecond * 100):
					break
				}
			}
		}()
	}

	wg.Wait()
}

func TestSafeTimer2(t *testing.T) {
	d := 100 * time.Millisecond
	timer := NewSafeTimer(d)

	// 测试 Reset 方法
	start := time.Now()
	if !timer.Reset(d) {
		t.Error("Expected timer.Reset to return true, got false")
	}
	<-timer.C()
	if time.Since(start) < d {
		t.Error("Timer expired too soon")
	}

	// 测试 Stop 方法
	timer.Reset(d)
	if !timer.Stop() {
		t.Error("Expected timer.Stop to return true, got false")
	}
	select {
	case <-timer.C():
		t.Error("Timer did not stop")
	case <-time.After(d + 50*time.Millisecond):
		// Timer stopped successfully
	}
}
