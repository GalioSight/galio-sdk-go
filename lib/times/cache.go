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

package times

import (
	"sync"
	"sync/atomic"
	"time"
)

var (
	secondPrecisionUnix     int64
	secondPrecisionUnixOnce sync.Once
)

// SecondPrecisionUnix 秒级精度的时间戳。
func SecondPrecisionUnix() int64 {
	secondPrecisionUnixOnce.Do(
		func() {
			atomic.StoreInt64(&secondPrecisionUnix, time.Now().Unix())
			go secondPrecisionUnixUpdate()
		},
	)
	return atomic.LoadInt64(&secondPrecisionUnix)
}

func secondPrecisionUnixUpdate() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for t := range ticker.C {
		atomic.StoreInt64(&secondPrecisionUnix, t.Unix())
	}
}
