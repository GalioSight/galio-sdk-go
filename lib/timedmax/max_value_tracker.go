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

// Package timedmax 定时最大值。
package timedmax

import (
	"sync"
	"time"
)

// MaxValueTracker 最大值跟踪器。
type MaxValueTracker struct {
	currentMaxExpiry time.Time     // 当前窗口过期时间。
	currentMax       float64       // 当前窗口最大值。
	minMax           float64       // 最小的最大值。
	window           time.Duration // 滑动窗口大小。
	mu               sync.Mutex
}

// NewMaxValueTracker 构造最大值跟踪器，滑动窗口大小：window，最小的最大值可配：minMax。
func NewMaxValueTracker(window time.Duration, minMax float64) *MaxValueTracker {
	return &MaxValueTracker{
		currentMaxExpiry: time.Now().Add(window),
		currentMax:       minMax,
		minMax:           minMax,
		window:           window,
	}
}

// Update 更新最大值。
func (m *MaxValueTracker) Update(value float64) float64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	if value > m.currentMax { // 如果值大于当前值，直接更新。
		m.currentMax = value
	}
	now := time.Now()
	newExpiry := now.Add(m.window)
	if now.After(m.currentMaxExpiry) { // 如果窗口变化，更新值，更新窗口。
		m.currentMax = value
		m.currentMaxExpiry = newExpiry
	}
	return m.currentMax
}

// Get 获取最大值。
func (m *MaxValueTracker) Get() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	if time.Now().After(m.currentMaxExpiry) { // 如果窗口外，用最小的最大值。
		return m.minMax
	}
	return m.currentMax
}
