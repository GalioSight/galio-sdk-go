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
	"math/rand"
	"time"
)

// Align 时间 time 对齐到 width（先整除，后翻倍）。
// 例子：time=1646880803 width=10，对齐后 1646880800。
func Align(time, width int64) int64 {
	return time / width * width
}

// WaitAlign 等待时间对齐到 width（单位秒）。
// 例子：当前时间 2022-3-17 12:26:52，width=10，等待到 2022-3-17 12:27:00 函数返回。
func WaitAlign(width int64) {
	if width <= 0 {
		return
	}
	nowSecond := time.Now().Unix()        // 当前时间。
	windowStarted := Align(nowSecond, 60) // 当前分钟的第 0 秒。
	for windowStarted < nowSecond {
		windowStarted += width // 确保每个窗口都是时间对齐的。
	}
	time.Sleep(time.Second * time.Duration(windowStarted-nowSecond))
}

// WaitRandomDuration 随机等待一段时间，sleep [0,d)，用于将不同 SDK 实例的请求打散，
// 避免同一时间重启的机器流量在时间上过于集中。
func WaitRandomDuration(d time.Duration) {
	time.Sleep(RandomDuration(d))
}

// RandomDuration 产生一个随机时间，取值位于 [0,d)，
func RandomDuration(d time.Duration) time.Duration {
	return time.Duration(float64(d) * rand.Float64())
}
