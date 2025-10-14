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

// Package limit ...
package limit

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBucket(t *testing.T) {
	Tick = 200 * time.Millisecond
	tb := New(2, 6)
	a := assert.New(t)
	a.True(tb.Consume(6))
	a.False(tb.Consume(1))
	HighResolutionWait(Tick)
	a.True(tb.Consume(1))
	a.True(tb.Consume(1))
	a.False(tb.Consume(1))

	tb = New(2, 6)
	a.False(tb.Consume(7))
}

func TestParallel(t *testing.T) {
	Tick = time.Second * 10
	a := assert.New(t)
	tb := New(500, 500)
	wg := sync.WaitGroup{}
	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.True(tb.Consume(1))
		}()
	}
	wg.Wait()
	a.False(tb.Consume(1))
}
