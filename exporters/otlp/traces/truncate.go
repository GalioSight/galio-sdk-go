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

package traces

import (
	"unicode/utf8"

	"galiosight.ai/galio-sdk-go/lib/strings"
)

const (
	maxStringLimit = 2 * 1024 * 1024 // 2MB
	// truncatedSuffix 用于 message 包体过长截断时的后缀
	truncatedSuffix = "...stringTooLong"
)

var maxStringLength = 32766

// SetMaxStringLength 允许用户自己调整最大长度，但不能超过 2MB 硬性限制
func SetMaxStringLength(limit int) {
	if limit > maxStringLimit {
		return
	}
	if limit < len(truncatedSuffix) {
		return
	}
	maxStringLength = limit
}

// GetMaxStringLength 获取最大长度
func GetMaxStringLength() int {
	return maxStringLength
}

// truncateBytes 将原始字符串在最长范围内截断
func truncateBytes(buf []byte) string {
	if len(buf) <= maxStringLength {
		return strings.NoAllocString(buf)
	}
	i := maxStringLength - len(truncatedSuffix)
	// 从最后面开始找到 UTF-8 字符的开始字节，保证截断位置刚好在字符开始处，避免将中文截断成乱码
	for i > 0 {
		if utf8.RuneStart(buf[i]) {
			break
		}
		i--
	}
	copy(buf[i:], truncatedSuffix)
	return strings.NoAllocString(buf[:i+len(truncatedSuffix)])
}

// truncatePurify 将原始字符串在最长范围内截断
func truncatePurify(s string) string {
	if len(s) > maxStringLength {
		return truncateBytes([]byte(s))
	}
	return s
}
