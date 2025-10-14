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

	"galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/lib/strings"
	"galiosight.ai/galio-sdk-go/model"
)

// BucketFunc 分桶获取函数。
type BucketFunc func() *configs.Bucket

func initHistogram(p *Point) {
	p.toOTPFunc = histogramToOTP
	p.updateFunc = updateHistogram
	p.changeFunc = changeHistogram
}

// SetBucket 设置人工分桶。
func (p *Point) SetBucket(bucketFunc BucketFunc) {
	if bucketFunc == nil {
		return
	}
	p.bucketCfgFunc = bucketFunc
	p.resetHistogram(p.bucketCfgFunc())
}

func updateHistogram(p *Point, v float64) {
	p.updateBucket(v)
	p.value += v
	p.incCount()
}

func (p *Point) updateBucket(v float64) {
	if !p.hasBucket() {
		return
	}
	bucketIdx := p.searchBucket(v)
	atomic.AddInt64(&p.counts[bucketIdx], 1)
}

func (p *Point) searchBucket(v float64) int {
	bucket := p.bucketCfg
	lb := lowerBound(bucket.Values, v)
	count := len(bucket.Values)
	if lb >= count { // 比最大的桶还大。
		return count - 1
	}
	if bucket.Values[lb] == v { // 刚刚好命中某个桶。
		return lb
	}
	lb--
	if lb < 0 { // 比最小的还小。
		return 0
	}
	return lb // 两个桶中间值。
}

func lowerBound(array []float64, target float64) int {
	low, high, mid := 0, len(array)-1, 0
	for low <= high {
		mid = (low + high) / 2
		if array[mid] >= target {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	return low
}

// histogramToOTP 如果返回 0，说明没有数据需要上报。
// 此时调用方需要判断，避免上报空数据。
// 否则会导致后台收到一个指标名为空的数据，带来干扰。
// go_gc_pause_seconds 指标需要判断，因为时间窗口内可能没有发生 gc。
func histogramToOTP(p *Point, injector model.OTPMetric, i int) int {
	// 一般的 histogram 指标，只有用户有上报，才会运行到这里，所以肯定就有数据，不会命中这个分支。
	// 但是 go_gc_pause_seconds 指标，是固定周期上报，如果此时间段没有 gc，就没有数据，会命中这个分支。
	if !hasData(p) {
		return 0
	}
	ranges, counts, sum, count := p.getAndClearHistogram()
	injector.SetName(i, p.Name())
	injector.SetAggregation(i, p.Aggregation())
	injector.SetHistogram(i, sum, count, ranges, counts)
	// 当前数据导出后，检查分桶变化。
	p.handleBucketChange()
	return 2 + len(ranges)
}

func (p *Point) handleBucketChange() {
	if p.bucketCfgFunc == nil {
		return
	}
	cfg := p.bucketCfgFunc()
	if cfg == nil {
		return
	}
	if !p.equalBucket(cfg.Key) {
		p.resetHistogram(cfg)
	}
}

// equalBucket 传入分桶 key，判断与当前桶是否一样。
func (p *Point) equalBucket(k string) bool {
	if !p.hasBucket() { // 当前没有桶。
		return false
	}
	return p.bucketCfg.Key == k
}

// resetHistogram 重新设置直方图。
func (p *Point) resetHistogram(cfg *configs.Bucket) {
	if cfg == nil { // 配置是空的，不更新。
		return
	}
	p.bucketCfg = cfg
	values := cfg.Values
	if cap(p.counts) < len(values) {
		p.counts = make([]int64, 0, len(values))
	}
	p.counts = p.counts[:len(values)]
	if cap(p.ranges) < len(values) {
		p.ranges = make([]string, 0, len(values))
	}
	p.ranges = p.ranges[:len(values)]
	for i := 0; i < len(values); i++ { // 重新配置每个桶的 vmrange，赋 0。
		start := strings.VMRangeFloatToString(values[i])
		end := strings.VMRangeMax
		if i+1 < len(values) {
			end = strings.VMRangeFloatToString(values[i+1])
		}
		p.ranges[i] = start + strings.VMRangeSeparator + end
		p.counts[i] = 0
	}
	p.value = 0 // 重新设置 sum + count。
	p.count = 0
}

func (p *Point) getAndClearHistogram() ([]string, []int64, float64, int64) {
	var ranges []string
	var counts []int64
	if len(p.ranges) > 0 && len(p.ranges) == len(p.counts) {
		ranges, counts = p.getAndClearBucket()
	}
	sum := p.value
	count := p.count
	p.value = 0
	p.count = 0
	return ranges, counts, sum, count
}

func (p *Point) getAndClearBucket() ([]string, []int64) {
	ranges := make([]string, 0, len(p.ranges))
	counts := make([]int64, 0, len(p.ranges))
	for i := range p.ranges {
		c := atomic.SwapInt64(&p.counts[i], 0)
		if c != 0 {
			ranges = append(ranges, p.ranges[i])
			counts = append(counts, c)
		}
	}
	return ranges, counts
}

// hasBucket 是否存在正确的人工分桶。
func (p *Point) hasBucket() bool {
	return p.bucketCfg != nil && len(p.bucketCfg.Values) != 0
}

func changeHistogram(p *Point, factor float64) {
	if p.hasBucket() {
		for i := range p.counts {
			p.counts[i] = roundInt64(float64(p.counts[i]) * factor)
		}
	}
	p.value *= factor
	p.count = roundInt64(float64(p.count) * factor)
}
