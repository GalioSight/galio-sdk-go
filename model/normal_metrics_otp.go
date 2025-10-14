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

var _ OTPMetric = (*NormalMetricOTP)(nil)

// NewNormalMetricsOTP 构造属性监控 otp 对象。
func NewNormalMetricsOTP() *NormalMetricOTP {
	return &NormalMetricOTP{
		Metric: &MetricOTP{},
	}
}

// SetName 设置第 i 个数据的 name，（单值监控项 i 填 0）。
func (n *NormalMetricOTP) SetName(i int, name string) {
	n.Metric.Name = name
}

// SetAggregation 设置第 i 个数据的策略，（单值监控项 i 填 0）。
func (n *NormalMetricOTP) SetAggregation(i int, a Aggregation) {
	n.Metric.Aggregation = a
}

// NewOTPValue 构造 otp 值类型（如：min、max、sum）。
func NewOTPValue(value float64) *MetricOTP_Value {
	return &MetricOTP_Value{Value: value}
}

// NewOTPHistogram 构造 otp histogram 类型。
func NewOTPHistogram(sum float64, count int64, counts []int64, ranges []string) *MetricOTP_Histogram {
	h := &MetricOTP_Histogram{
		Histogram: &Histogram{
			Sum:   sum,
			Count: count,
		},
	}
	h.Histogram.Buckets = make([]*Bucket, len(counts))
	for i := range counts {
		h.Histogram.Buckets[i] = &Bucket{Range: ranges[i], Count: counts[i]}
	}
	return h
}

// NewOTPAvg 构造 otp avg 类型。
func NewOTPAvg(sum float64, count int64) *MetricOTP_Avg {
	return &MetricOTP_Avg{
		Avg: &Avg{
			Sum:   sum,
			Count: count,
		},
	}
}

// SetHistogram 设置第 i 个数据的 histogram 值，（单值监控项 i 填 0）。
func (n *NormalMetricOTP) SetHistogram(i int, sum float64, count int64, ranges []string, counts []int64) {
	n.Metric.V = NewOTPHistogram(sum, count, counts, ranges)
}

// SetAvg 设置第 i 个数据的 avg 值，（单值监控项 i 填 0）。
func (n *NormalMetricOTP) SetAvg(i int, sum float64, count int64) {
	n.Metric.V = NewOTPAvg(sum, count)
}

// SetCount 设置第 i 个数据的 counter 值，（单值监控项 i 填 0）。
func (n *NormalMetricOTP) SetCount(i int, count int64) {
	n.SetValue(i, float64(count))
}

// SetValue 设置第 i 个数据的其他类型值（如：min、max、sum、set），（单值监控项 i 填 0）。
func (n *NormalMetricOTP) SetValue(i int, value float64) {
	n.Metric.V = NewOTPValue(value)
}
