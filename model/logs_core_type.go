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

// CoreType 日志 core 类型
type CoreType int32

// 日志 Core 类型
const (
	IOCore     CoreType = 0 // zap 的 ioCore，默认类型，支持符合级别日志
	SampleCore CoreType = 1 // 支持符合级别日志和所有级别采样日志
)

type LogTracedType string

const (
	LogTracedSample LogTracedType = "sample"
	LogTracedDyeing LogTracedType = "dyeing"
)
