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

var _ OMPMetric = (*NormalMetric)(nil)

// NewNormalMetric 构造 *NormalMetric。
func NewNormalMetric() *NormalMetric {
	return &NormalMetric{}
}

// Group 监控分组。
func (n *NormalMetric) Group() MetricGroup {
	return NormalGroup
}

// PointCount 数据点个数（单值监控项 1，多值监控项大于 1）。
func (n *NormalMetric) PointCount() int {
	return 1 // 属性监控只支持单值语义，1 个数据点。
}

// PointName 第 i 个监控点的名称。
func (n *NormalMetric) PointName(i int) string {
	return n.Metric.Name
}

// PointAggregation 第 i 个监控点的策略。
func (n *NormalMetric) PointAggregation(i int) Aggregation {
	return n.Metric.Aggregation
}

// PointValue 第 i 个监控点的值。
func (n *NormalMetric) PointValue(i int) float64 {
	return n.Metric.Value
}

// LabelCount 标签数量。
func (n *NormalMetric) LabelCount() int {
	return 0
}

// LabelValue 第 i 个标签的值。
func (n *NormalMetric) LabelValue(i int) string {
	return ""
}
