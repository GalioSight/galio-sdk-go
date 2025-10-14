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
	"galiosight.ai/galio-sdk-go/model"
)

func initSum(p *Point) {
	resetSum(p)
	p.toOTPFunc = sumToOTP
	p.updateFunc = updateSum
	p.changeFunc = changeSum
}

func updateSum(p *Point, v float64) {
	p.value += v
	p.incCount()
}

func sumToOTP(p *Point, injector model.OTPMetric, i int) int {
	if !hasData(p) {
		return 0
	}
	v := p.getAndResetSum()
	injector.SetName(i, p.Name())
	injector.SetAggregation(i, p.Aggregation())
	injector.SetValue(i, v)
	return 1
}

func (p *Point) getAndResetSum() float64 {
	v := p.value
	resetSum(p)
	return v
}

func resetSum(p *Point) {
	p.value = 0
	p.resetCount()
}

func changeSum(p *Point, factor float64) {
	p.value *= factor
	p.count = roundInt64(float64(p.count) * factor)
}
