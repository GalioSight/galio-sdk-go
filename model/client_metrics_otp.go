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

package model

var _ OTPMetric = (*ClientMetricsOTP)(nil)

// NewClientMetricsOTP 构造 *ClientMetricsOTP
func NewClientMetricsOTP() *ClientMetricsOTP {
	return &ClientMetricsOTP{RpcLabels: &RPCLabels{}}
}

// SetName 设置第 i 个数据的 name，（单值监控项 i 填 0）。
func (c *ClientMetricsOTP) SetName(i int, name string) {
	// ClientMetricsOTP 明确语义，不需要设置 name。
}

// SetAggregation 设置第 i 个数据的策略，（单值监控项 i 填 0）。
func (c *ClientMetricsOTP) SetAggregation(i int, a Aggregation) {
	// ClientMetricsOTP 明确语义，不需要设置数据策略。
}

// NewHistogram 构造 histogram。
// count：监控点的次数，sum：监控点的总数（比如：总耗时）。
// counts：每个桶的次数，ranges：每个桶的标签（比如："1.896e-01...2.154e-01"）。
func NewHistogram(count int64, sum float64, counts []int64, ranges []string) *Histogram {
	h := &Histogram{}
	h.Count = count
	h.Sum = sum
	h.Buckets = make([]*Bucket, len(counts))
	for i := range counts {
		h.Buckets[i] = &Bucket{Range: ranges[i], Count: counts[i]}
	}
	return h
}

// SetHistogram 设置第 i 个数据的 histogram 值，（单值监控项 i 填 0）。
func (c *ClientMetricsOTP) SetHistogram(i int, sum float64, count int64, ranges []string, counts []int64) {
	if i == ClientMetricHandledSecondsPoint {
		c.RpcClientHandledSeconds = NewHistogram(count, sum, counts, ranges)
	}
}

// SetAvg 设置第 i 个数据的 avg 值，（单值监控项 i 填 0）。
func (c *ClientMetricsOTP) SetAvg(i int, sum float64, count int64) {
	// ClientMetricsOTP 明确语义，只有 counter 和 histogram 点，不需要设置 avg。
}

// SetCount 设置第 i 个数据的 counter 值，（单值监控项 i 填 0）。
func (c *ClientMetricsOTP) SetCount(i int, count int64) {
	if i == ClientMetricStartedTotalPoint {
		c.RpcClientStartedTotal = int64(count)
	} else if i == ClientMetricHandledTotalPoint {
		c.RpcClientHandledTotal = int64(count)
	}
}

// SetValue 设置第 i 个数据的其他类型值（如：min、max、sum、set），（单值监控项 i 填 0）。
func (c *ClientMetricsOTP) SetValue(i int, value float64) {
	if i == ClientMetricStartedTotalPoint {
		c.RpcClientStartedTotal = int64(value)
	} else if i == ClientMetricHandledTotalPoint {
		c.RpcClientHandledTotal = int64(value)
	}
}
