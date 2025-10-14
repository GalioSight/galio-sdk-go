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

//go:build sonic
// +build sonic

package profiles

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"testing"
	"time"
	_ "unsafe"

	"galiosight.ai/galio-sdk-go/self/log"
	"github.com/bytedance/sonic"
)

//go:linkname stopProfiling github.com/bytedance/sonic/internal/rt.StopProfiling
var stopProfiling bool

func TestSonicPprofConflict(t *testing.T) {
	stopProfiling = true
	log.Infof("stopProfiling = %v", stopProfiling)

	// ctx, cancelFunc := context.WithTimeout(context.Background(), 25*time.Second)
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancelFunc()

	var wg sync.WaitGroup
	go startCPUProfile()

	goroutineCount := runtime.NumCPU() * 2
	wg.Add(goroutineCount)
	for i := 0; i < goroutineCount; i++ {
		go func(ctx context.Context) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					sonicDemo()
				}
			}
		}(ctx)
	}
	wg.Wait()
}

func startCPUProfile() {
	var f bytes.Buffer
	for {
		if err := pprof.StartCPUProfile(&f); err != nil {
			log.Fatal(err)
		}
		time.Sleep(3 * time.Second)
		pprof.StopCPUProfile()
	}
}

// sonicDemo 无限循环 sonic 解码随机 JSON 字符串，触发即时编译，复现 pprof 和 sonic 的冲突
func sonicDemo() {
	var data interface{}
	jsonStr := randomJSON()
	for i := 0; i < 1000000; i++ {
		if err := sonic.UnmarshalString(jsonStr, &data); err != nil {
			log.Error(err)
		}
	}
}

// randomJSON 随机生成 JSON 字符串
func randomJSON() string {
	var builder strings.Builder
	builder.WriteString("{")
	for i := 0; i < rand.Intn(30); i++ {
		randStr := randomString(10)
		builder.WriteString(fmt.Sprintf("\"%s\": \"%s\",", randStr, randStr))
	}
	randStr := randomString(10)
	builder.WriteString(fmt.Sprintf("\"%s\": \"%s\"", randStr, randStr))
	builder.WriteString("}")
	return builder.String()
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var letterCount = len(letters)

// randomString 生成随机字符串
func randomString(n int) string {
	offset := rand.Intn(letterCount)
	step := rand.Intn(letterCount)

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[offset%letterCount]
		offset += step
	}

	return string(s)
}
