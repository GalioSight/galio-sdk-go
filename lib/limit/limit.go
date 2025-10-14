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

// Package limiter a token bucket implement inspired by
// https://raw.githubusercontent.com/rigtorp/TokenBucket/master/TokenBucket.h
// 对比 golang.org/x/time/rate 无锁，性能更高
// 对比 uber ratelimit，非阻塞
// 对比 https://github.com/hedzr/rate 没有额外的线程
package limit

import (
	"sync/atomic"
	"time"
)

var Tick time.Duration = time.Second

type TokenBucket struct {
	timePerToken time.Duration
	timePerBurst time.Duration
	times        atomic.Value
}

func New(rate, burst uint64) *TokenBucket {
	timePerToken := Tick / time.Duration(rate)
	ret := &TokenBucket{
		timePerToken: timePerToken,
		timePerBurst: time.Duration(burst) * timePerToken,
	}
	ret.times.Store(time.Time{})
	return ret
}

func (tb *TokenBucket) Consume(tokens uint64) bool {
	now := time.Now()
	timeNeeded := time.Duration(tokens) * tb.timePerToken
	minTime := now.Add(-tb.timePerBurst)

	for {
		// 在 C++ 的实现中，使用 std::memory_order_relaxed 的 atomic 实现，即使不在
		// for 循环内部每次更新，也能访问到最新的 oldTime 在 go 中，atomic 是强同步的操
		// 作，如果在 for 外面读取 oldTime，会导致别的协程修改了 atomic 后，自己的
		// oldTime Compare 一定会失败，造成死循环
		// 因此简单的做法是，for 循环内每次取出最新的 oldTime 即可。当然，其性能会比 C++ 的更差
		oldTime := tb.times.Load().(time.Time)
		newTime := oldTime
		if minTime.After(newTime) {
			newTime = minTime
		}
		newTime = newTime.Add(timeNeeded)
		if newTime.After(now) {
			return false
		}
		if tb.times.CompareAndSwap(oldTime, newTime) {
			return true
		}
	}
}

// HighResolutionWait 用于单测，高精度等待，消耗 CPU 资源
func HighResolutionWait(d time.Duration) {
	until := time.Now().Add(d)
	for {
		if time.Now().After(until) {
			break
		}
	}
}
