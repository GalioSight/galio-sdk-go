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

//go:build (windows || darwin || linux) && amd64

package traces

import (
	"bytes"
	"fmt"
	"runtime"
	"sync/atomic"

	selflog "galiosight.ai/galio-sdk-go/self/log"
	"github.com/bytedance/sonic"
	"github.com/nanmu42/limitio"
)

// json 方式序列化为 string 相比 pb 序列化为 string 有如下优点：
//
// 1. 包体包含中文可以正常显示可读性好的中文字符串而不是\346\267\261 这样的以`%03o`格式打印原始的 rune, 便于查看。
// 2. 便于复制数据提供给 trpc cli 作为请求包体，便于调试。
// 3. 使用 sonic 序列化为 JSON string 时性能高于 proto.MarshalText.

var (
	// 是否使用 sonic fastest 模式进行序列化，使用 atomic 访问以保证并发安全
	sonicFastest uint32 = 1
)

// DefaultMarshalToString 函数将给定的结构体序列化为字符串。
// 当序列化得到的字符串超出长度时，会进行截断，可以通过 SetMaxStringLength 方法调整字符串长度。
// 当 message 内部有循环引用时，JSON 序列化会出错，这时候会使用 fmt 来进行序列化。
// 业务可以通过自定义 JSON Marshal 或 message.String() 方法来控制数据格式。
func DefaultMarshalToString(message interface{}) string {
	buf, err := sonicMarshal(message)
	if err != nil {
		b := bytes.NewBuffer(nil)
		// 允许超过最大长度，方便后续统一截断，以保证最后一个字符不会乱码。
		_, _ = fmt.Fprintf(limitio.NewWriter(b, maxStringLength+1, false), "%+v", message)
		buf = b.Bytes()
	}
	return truncateBytes(buf)
}

func sonicMarshal(message interface{}) ([]byte, error) {
	defer handlePanic("message:", message)
	if atomic.LoadUint32(&sonicFastest) == 1 {
		return sonic.ConfigFastest.Marshal(message)
	} else {
		return sonic.ConfigStd.Marshal(message)
	}
}

// SetSonicFastest 函数用于设置 sonicFastest 变量的值。
// 当设置为 false 时，将使用 sonic.ConfigStd 进行序列化。
func SetSonicFastest(fastest bool) {
	if fastest {
		atomic.StoreUint32(&sonicFastest, 1)
	} else {
		atomic.StoreUint32(&sonicFastest, 0)
	}
}

func handlePanic(labels ...interface{}) {
	if err := recover(); err != nil {
		const (
			panicGalileoPrefix = "[PANIC][GALILEO]"
			PanicBufLen        = 1024 * 1024
		)
		buf := make([]byte, PanicBufLen)
		buf = buf[:runtime.Stack(buf, false)]
		msg := fmt.Sprintf(panicGalileoPrefix+"[%v]%v\n%s\n", labels, err, buf)
		selflog.Errorf(msg)
	}
}
