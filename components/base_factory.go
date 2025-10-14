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

package components

// BaseFactory 基础工厂，由更具体的工厂实现（处理器工厂、导出器工厂）。
type BaseFactory interface {
	// Protocol 协议，如：omp、otp、otlp。
	Protocol() string
}

type baseFactory struct {
	protocol string
}

// Protocol 协议，如：omp、otp、otlp。
func (b baseFactory) Protocol() string {
	return b.protocol
}
