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

type ProfileType string

const (
	// CPUProfile 收集 CPU 消耗
	CPUProfile ProfileType = "cpu"
	// HeapProfile 收集内存分配采样，用于检查内存使用情况，帮助检查内存泄漏
	HeapProfile ProfileType = "heap"
	// BlockProfile 收集 goroutines 在 mutex 和 channel 操作上的阻塞等待
	// 可能会导致明显的 CPU 开销。默认不开启
	BlockProfile ProfileType = "block"
	// MutexProfile 收集锁竞争，帮助判断 CPU 是否浪费在因为互斥锁竞争，默认不开启
	MutexProfile ProfileType = "mutex"
	// GoroutineProfile 收集当前所有 goroutines 的 stack traces
	GoroutineProfile ProfileType = "goroutine"
)
