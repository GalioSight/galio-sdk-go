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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetMaxStringLength(t *testing.T) {
	// 测试正常设置最大长度
	SetMaxStringLength(1000)
	assert.Equal(t, 1000, maxStringLength)

	// 测试设置为 2MB 限制
	SetMaxStringLength(maxStringLimit)
	assert.Equal(t, maxStringLimit, maxStringLength)

	// 测试设置超过 2MB 限制
	SetMaxStringLength(maxStringLimit + 1)
	assert.Equal(t, maxStringLimit, maxStringLength)

	// 测试设置小于截断后缀长度
	SetMaxStringLength(len(truncatedSuffix) - 1)
	assert.Equal(t, maxStringLength, maxStringLength)
}

func TestTruncatePurify(t *testing.T) {
	old := maxStringLength
	defer func() {
		maxStringLength = old
	}()
	cases := []struct {
		Name      string
		MaxLength int
		S         string
		Want      string
	}{
		{"普通字符串不超长", 30, "ABC", "ABC"},
		{"普通字符串不超长", 10, "1234567890", "1234567890"},
		{"不超长非法 UTF8", 30, "Geeks\xc5Geeks", "Geeks\xc5Geeks"},
		{
			"普通超长字符串",
			30,
			"1234567890123456789012345678901",
			"12345678901234" + truncatedSuffix,
		},
		{
			"普通超长非法 UTF8 字符串",
			30,
			"\xc51234567890123456789012345678901",
			"\xc51234567890123" + truncatedSuffix,
		},
		{
			"截断超长非法 UTF8 字符串",
			30,
			"一二三四五一二三四五 1",
			"一二三四" + truncatedSuffix,
		},
		{
			"截断超长非法 UTF8 字符串",
			29,
			"一二三四五一二三四五 1",
			"一二三四" + truncatedSuffix,
		},
		{
			"截断超长非法 UTF8 字符串",
			28,
			"一二三四五一二三四五 1",
			"一二三四" + truncatedSuffix,
		},
	}
	for _, c := range cases {
		t.Run(
			c.Name, func(t *testing.T) {
				t.Log(maxStringLength)
				maxStringLength = c.MaxLength
				res := truncatePurify(c.S)
				assert.Equal(t, c.Want, res)
			},
		)
	}
}

func TestGetMaxStringLength(t *testing.T) {
	assert.Equal(t, maxStringLength, GetMaxStringLength())
}
