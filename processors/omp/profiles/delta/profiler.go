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

// Package delta 实现 profile 的增量计算
package delta

import "bytes"

// Profiler 定义 delta profiler 的接口
type Profiler interface {
	Delta(data []byte) ([]byte, error)
}

// DeltaValueType 定义支持 delta 的 SampleType
type DeltaValueType struct {
	Type string
	Unit string
}

func isGzipData(data []byte) bool {
	return bytes.HasPrefix(data, []byte{0x1f, 0x8b})
}
