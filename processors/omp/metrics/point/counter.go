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

package point

import (
	"sync/atomic"

	"galiosight.ai/galio-sdk-go/model"
)

func initCounter(p *Point) {
	resetCounter(p)
	p.toOTPFunc = counterToOTP
	p.updateFunc = updateCounter
	p.changeFunc = changeCounter
}

func updateCounter(p *Point, v float64) {
	atomic.AddInt64(&p.counter, int64(v))
	p.incCount()
}

func counterToOTP(p *Point, injector model.OTPMetric, i int) int {
	if !hasData(p) {
		return 0
	}
	v := p.getAndResetCounter()
	injector.SetName(i, p.Name())
	injector.SetAggregation(i, p.Aggregation())
	injector.SetCount(i, v)
	return 1
}

func (p *Point) getAndResetCounter() int64 {
	v := atomic.LoadInt64(&p.counter)
	resetCounter(p)
	return v
}

func resetCounter(p *Point) {
	atomic.StoreInt64(&p.counter, 0)
	p.resetCount()
}

func changeCounter(p *Point, factor float64) {
	p.counter = roundInt64(float64(p.counter) * factor)
	p.count = roundInt64(float64(p.count) * factor)
}
