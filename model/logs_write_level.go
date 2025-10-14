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

// WriteLevel write syncer 日志级别
// 用户打日志 -> core -> write syncer，每个层级都可以有日志级别控制。
type WriteLevel int32

// write syncer 日志级别枚举值
const (
	WriteAll WriteLevel = 0 // 打所有日志。
	// WriteHitTraceSampling 已废弃，前置到 core 层过滤采样日志
	WriteHitTraceSampling WriteLevel = 1 // 只打命中了 trace 采样的日志
	// WriteMustLogTraced 已废弃，前置到 core 层过滤采样日志
	WriteMustLogTraced WriteLevel = 2 // 只要命中了 trace 的日志都打，不管级别（范围仅次于 All，由于历史兼容只能设置为 2）
)
