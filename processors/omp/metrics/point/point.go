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

// Package point 数据点统一抽象，实现。
// 实现类型：counter、histogram、avg、set、sum、max、min。
package point

import (
	"fmt"
	"sync"
	"sync/atomic"

	"galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/model"
)

type (
	updateFunc     func(p *Point, v float64)
	pointToOTPFunc func(p *Point, injector model.OTPMetric, i int) int
	changeFunc     func(p *Point, factor float64)
)

// Point 数据点。
type Point struct {
	bucketCfgFunc BucketFunc      // 分桶配置，更新函数。
	bucketCfg     *configs.Bucket // 分桶配置。
	updateFunc    updateFunc      // 数据更新函数。
	toOTPFunc     pointToOTPFunc  // 数据转 otp 函数。
	changeFunc    changeFunc      // 修改数据函数，根据放大系数修改数据，仅命中采样且采样率不为 1 时需要调用该函数。
	name          string
	ranges        []string          // histogram 分桶范围列表，如 1...2、2...4。
	counts        []int64           // histogram 分桶计数列表，如 10、20。
	counter       int64             // counter。
	value         float64           // avg、histogram、max、min、set、sum。
	count         int64             // 计数，数据更新 1 次，+1。
	aggregation   model.Aggregation // 数据点聚合类型。
}

// get 构造数据点
func get(name string, aggregation model.Aggregation, defaultOpts []option) *Point {
	p := getOrNewPoint(aggregation)
	p.name = name
	p.aggregation = aggregation
	for i := range defaultOpts {
		defaultOpts[i](p)
	}
	return p
}

var (
	pointPool          sync.Pool
	pointHistogramPool sync.Pool
)

func getOrNewPoint(aggregation model.Aggregation) *Point {
	if aggregation == model.Aggregation_AGGREGATION_HISTOGRAM {
		if p, ok := pointHistogramPool.Get().(*Point); ok {
			return p
		}
		return &Point{}
	}
	if p, ok := pointPool.Get().(*Point); ok {
		return p
	}
	return &Point{}
}

func (p *Point) reset() {
	p.name = ""
	p.aggregation = model.Aggregation_AGGREGATION_NONE
	p.counter = 0
	p.value = 0
	p.count = 0
	p.bucketCfgFunc = nil
	p.bucketCfg = nil
	for i := range p.counts {
		p.counts[i] = 0
	}
	p.counts = p.counts[:0]
	for i := range p.ranges {
		p.ranges[i] = ""
	}
	p.ranges = p.ranges[:0]
	p.updateFunc = nil
	p.toOTPFunc = nil
	p.changeFunc = nil
}

// Get 根据数据策略 aggregation 构造数据点。
func Get(aggregation model.Aggregation, name string) *Point {
	defaultOpts := getDefaultOptions(aggregation)
	return get(name, aggregation, defaultOpts)
}

// Put Point 过期清理后返回对象池。
func Put(p *Point) {
	if p == nil {
		return
	}
	a := p.aggregation
	p.reset()
	if a == model.Aggregation_AGGREGATION_HISTOGRAM {
		pointHistogramPool.Put(p)
	} else {
		pointPool.Put(p)
	}
}

// Update 更新数据。
func (p *Point) Update(v float64) {
	if p == nil {
		return
	}
	if p.updateFunc == nil { // TODO 重构这里的更新逻辑，不用函数指针。
		return
	}
	p.updateFunc(p, v) // 30 ns/op
}

func hasData(p *Point) bool {
	return p.Count() != 0
}

// ToOTP 导出到 OTPMetric 的第 i 个数据点。
func (p *Point) ToOTP(injector model.OTPMetric, i int) (int, error) {
	if p.toOTPFunc != nil {
		return p.toOTPFunc(p, injector, i), nil
	}
	return 0, fmt.Errorf("exportFunc nil")
}

func (p *Point) resetCount() {
	atomic.StoreInt64(&p.count, 0)
}

// Count 原始数据条数。
func (p *Point) Count() int64 {
	return atomic.LoadInt64(&p.count)
}

func (p *Point) incCount() {
	atomic.AddInt64(&p.count, 1)
}

// Name 监控项名称。
func (p *Point) Name() string {
	return p.name
}

// Aggregation 数据策略。
func (p *Point) Aggregation() model.Aggregation {
	return p.aggregation
}

// Change 根据放大系数修改数据。
func (p *Point) Change(factor float64) error {
	if p.changeFunc == nil {
		return fmt.Errorf("changeFunc nil")
	}
	p.changeFunc(p, factor)
	return nil
}

// roundInt64 四舍五入。正数结果符合预期，负数存在误差，此处只由正数调用。
func roundInt64(v float64) int64 {
	return int64(v + 0.5)
}
