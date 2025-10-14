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

// Package traces ...
package traces

import (
	attrutil "galiosight.ai/galio-sdk-go/exporters/otlp/traces/attribute"
	"github.com/alphadose/haxmap"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/atomic"
)

// WindowSize 窗口个数，保留 2 个窗口，用于时间窗到期后无锁切换
const WindowSize int8 = 2

type samplerWindow struct {
	keys        *haxmap.Map[string, int32]
	len         atomic.Int32
	maxKeyCount int32
}

type minCountWindow struct {
	windows   [WindowSize]*samplerWindow
	windowIdx int8
}

func (s *samplerWindow) getSampledCount(key string) (val int32, succ bool) {
	val, ok := s.keys.Get(key)
	if ok {
		return val, ok
	}
	// key 总数超过最大限制，不再走最小采样逻辑
	if s.len.Load() > s.maxKeyCount {
		return 0, false
	}
	// 当前 key 首次出现，计数并返回计数器
	val, loaded := s.keys.GetOrSet(key, 0)
	if loaded {
		return val, false
	}
	// 新 key 插入成功
	s.len.Add(1)
	return val, true
}

// shiftWindow 切换到下一个时间窗并清空上一个时间窗的数据
func (a *adaptiveSampler) shiftWindow() {
	nextIdx := getNextIndex(a.windowIdx)
	n := []uintptr{a.windows[a.windowIdx].keys.Len()} // prealloc size
	if n[0] == 0 {
		n = n[:0] // ignore size param to let haxmap use default size
	}
	a.windows[nextIdx] = &samplerWindow{maxKeyCount: a.opts.windowMaxKeyCount, keys: haxmap.New[string, int32](n...)}
	a.windowIdx = nextIdx
}

// minCountSampler 最小采样逻辑，确保每个接口组合都能被采集到
func (a *adaptiveSampler) minCountSampler(p *sdktrace.SamplingParameters) bool {
	if !a.opts.enableMinSample {
		return false
	}
	// keyIndex 是高效的返回 4 元组，我们只用 callee 的前两个元素
	return a.minCount(minCountKey(p))
}

func minCountKey(p *sdktrace.SamplingParameters) string {
	keys := attrutil.NewCalleeKeys(p.Attributes)
	return keys.CalleeMethod + "_" + keys.CalleeService
}

// minCount 最小采样逻辑，确保每个接口组合都能被采集到至少 minSampleCount 个 trace
// 但是这个接口不保证每个组合采样到的个数一定等于 minSampleCount 个
func (a *adaptiveSampler) minCount(key string) bool {
	if a.opts.minSampleCount <= 0 {
		return false
	}
	count, ok := a.windows[a.windowIdx].getSampledCount(key)
	if ok && count < a.opts.minSampleCount {
		return a.windows[a.windowIdx].keys.CompareAndSwap(key, count, count+1)
	}
	return false
}

func getNextIndex(index int8) int8 {
	index++
	if index >= WindowSize {
		return 0
	}
	return index
}
