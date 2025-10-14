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

var _ OTPMetric = (*CustomMetricsOTP)(nil)

// NewCustomMetricsOTP 构造 *CustomMetricsOTP，标签长度 labelCount，数据点个数 pointCount。
func NewCustomMetricsOTP(labelCount, pointCount int) *CustomMetricsOTP {
	c := &CustomMetricsOTP{
		Metrics:      make([]*MetricOTP, pointCount),
		CustomLabels: make([]*Label, labelCount),
	}
	for i := range c.Metrics {
		c.Metrics[i] = &MetricOTP{}
	}
	for i := range c.CustomLabels {
		c.CustomLabels[i] = &Label{}
	}
	return c
}

// SetMonitorName 设置自定义指标的 monitor name
func (c *CustomMetricsOTP) SetMonitorName(monitorName string) {
	c.MonitorName = monitorName
}

// SetName 设置第 i 个数据的 name，（单值监控项 i 填 0）。
func (c *CustomMetricsOTP) SetName(i int, name string) {
	c.Metrics[i].Name = name
}

// SetAggregation 设置第 i 个数据的策略，（单值监控项 i 填 0）。
func (c *CustomMetricsOTP) SetAggregation(i int, a Aggregation) {
	c.Metrics[i].Aggregation = a
}

// SetHistogram 设置第 i 个数据的 histogram 值，（单值监控项 i 填 0）。
func (c *CustomMetricsOTP) SetHistogram(i int, sum float64, count int64, ranges []string, counts []int64) {
	c.Metrics[i].V = NewOTPHistogram(sum, count, counts, ranges) // TODO(jaimeyang) 内存优化，get from pool。
}

// SetAvg 设置第 i 个数据的 avg 值，（单值监控项 i 填 0）。
func (c *CustomMetricsOTP) SetAvg(i int, sum float64, count int64) {
	c.Metrics[i].V = NewOTPAvg(sum, count)
}

// SetCount 设置第 i 个数据的 counter 值，（单值监控项 i 填 0）。
func (c *CustomMetricsOTP) SetCount(i int, count int64) {
	c.SetValue(i, float64(count))
}

// SetValue 设置第 i 个数据的其他类型值（如：min、max、sum、set），（单值监控项 i 填 0）。
func (c *CustomMetricsOTP) SetValue(i int, value float64) {
	c.Metrics[i].V = NewOTPValue(value)
}
