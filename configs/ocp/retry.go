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

package ocp

import (
	"time"

	"galiosight.ai/galio-sdk-go/errs"
)

// retry 尝试执行 f，直到成功或者超时，重试时不会 sleep。
// 如果调用方需要 sleep，可以在 f 中执行。
// maxRetryCount 最大重试次数，f 函数最多会执行 maxRetryCount+1 次。
// 之前的版本在循环条件入口判断超时，会出现机器负载特别高时，第一次判断就超时的情况，导致循环一次都未执行。
// 现在改成在循环内部判断超时，保证 maxRetryCount>=0 时 至少调用一次 f 函数。
func retry(f func() error, timeout time.Duration, maxRetryCount int) error {
	var err error
	end := time.Now().Add(timeout)
	for i := 0; i <= maxRetryCount; i++ {
		err = f()
		if err == nil {
			return nil
		}
		if time.Now().After(end) {
			return errs.ErrTimeout
		}
	}
	return errs.ErrTimeout
}
