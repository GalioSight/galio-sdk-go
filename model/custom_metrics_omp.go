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

var _ OMPMetric = (*CustomMetrics)(nil)

// NewCustomMetrics 构造 *CustomMetrics，标签长度 labelCount，数据点个数 pointCount。
func NewCustomMetrics(labelCount, pointCount int) *CustomMetrics {
	return &CustomMetrics{
		Metrics:      make([]Metric, pointCount),
		CustomLabels: make([]Label, labelCount),
	}
}

// Group 监控分组。
func (c *CustomMetrics) Group() MetricGroup {
	return CustomGroup
}

// PointCount 数据点个数（单值监控项 1，多值监控项大于 1）。
func (c *CustomMetrics) PointCount() int {
	return len(c.Metrics)
}

// PointName 第 i 个监控点的名称。
func (c *CustomMetrics) PointName(i int) string {
	if i < 0 || i >= c.PointCount() {
		return ""
	}
	return c.Metrics[i].Name
}

// PointAggregation 第 i 个监控点的策略。
func (c *CustomMetrics) PointAggregation(i int) Aggregation {
	if i < 0 || i >= c.PointCount() {
		return Aggregation_AGGREGATION_NONE
	}
	return c.Metrics[i].Aggregation
}

// PointValue 第 i 个监控点的值。
func (c *CustomMetrics) PointValue(i int) float64 {
	if i < 0 || i >= c.PointCount() {
		return 0
	}
	return c.Metrics[i].Value
}

// LabelCount 标签数量。
func (c *CustomMetrics) LabelCount() int {
	return len(c.CustomLabels)
}

// LabelValue 第 i 个标签的值。
func (c *CustomMetrics) LabelValue(i int) string {
	if i < 0 || i >= c.LabelCount() {
		return ""
	}
	return c.CustomLabels[i].Value
}

// MetricKey 指标 cache 主键，要求此 key 必须唯一，即不能上报同名但是 schema 不同的数据。
type MetricKey struct {
	// 监控项
	MonitorName string
	// 指标名
	MetricName string
	// 聚合方式
	Aggregation Aggregation
}
