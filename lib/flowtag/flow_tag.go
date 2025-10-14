// Copyright 2024 Tencent Galileo Authors
//
// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// template file: https://github.com/open-telemetry/opentelemetry-go/blob/main/semconv/template.j2

// Package flowtag 定义流量标签，如灰度、降级、重试等
package flowtag

import (
	"strings"

	semconv "galiosight.ai/galio-sdk-go/semconv/v1.0.0"
)

// FlowTag 定义 FlowTag 类型
type FlowTag uint32

// 定义可用的 FlowTag 枚举值，用于 CD 可观测。
// 当前限定为以下值，以避免数据出现不规范的情况，如有需要，可以联系伽利略小助手添加新的标签。
const (
	// Gray 灰度
	Gray FlowTag = 1 << iota // 1
	// Downgrade 降级
	Downgrade // 2
	// Retry 重试
	Retry // 4
	// maxFlowTag FlowTag 的最大值，随着版本变化，此值会发生变化，不是固定常量，不导出
	maxFlowTag
)

// singleTagNames 用于存储 FlowTag 的名称，与前面的常量定义对应
var singleTagNames = []string{
	semconv.FlowTagGrayValue,
	semconv.FlowTagDowngradeValue,
	semconv.FlowTagRetryValue,
}

// flowTagNameCache 组合流量标签值的 cache，用于加速 FlowTag.String 函数，避免重复拼接字符串
var flowTagNameCache = [maxFlowTag]string{}

func init() {
	for i := FlowTag(0); i < maxFlowTag; i++ {
		flowTagNameCache[i] = i.toString()
	}
}

// String 返回 FlowTag 的字符串表示
func (f FlowTag) String() string {
	if f >= maxFlowTag {
		return ""
	}
	return flowTagNameCache[f]
}

// String 返回 FlowTag 的字符串表示
func (f FlowTag) toString() string {
	var tags []string
	for i := 0; i < len(singleTagNames); i++ {
		if f.Has(FlowTag(1 << i)) {
			tags = append(tags, singleTagNames[i])
		}
	}
	return strings.Join(tags, ",")
}

// Has 检查是否包含特定的 FlowTag
func (f FlowTag) Has(tag FlowTag) bool {
	return f&tag != 0
}
