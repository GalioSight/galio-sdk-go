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

//go:build !((windows || darwin || linux) && amd64)

// Package traces ...
package traces

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/nanmu42/limitio"
)

// json 方式序列化为 string 相比 pb 序列化为 string 有如下优点：
//
// 1. 包体包含中文可以正常显示可读性好的中文字符串而不是\346\267\261 这样的以`%03o`格式打印原始的 rune, 便于查看。
// 2. 便于复制数据提供给 trpc cli 作为请求包体，便于调试。

// DefaultMarshalToString 函数将给定的结构体序列化为字符串。
// 当 message 内部有循环引用时，JSON 无法正确序列化。
// 所以对于 JSON 序列化出错的情况，使用 fmt 来进行序列化。
// 业务可以通过自定义 JSON Marshal 或 message.String() 方法来控制数据格式。
func DefaultMarshalToString(message interface{}) string {
	buf, err := json.Marshal(message)
	if err != nil {
		b := bytes.NewBuffer(nil)
		fmt.Fprintf(limitio.NewWriter(b, maxStringLength+1, false), "%+v", message)
		buf = b.Bytes()
	}
	return truncateBytes(buf)
}

// SetSonicFastest 此处仅为了在 !windows && !darwin && !linux 的情况下编译通过
func SetSonicFastest(fastest bool) {
}
