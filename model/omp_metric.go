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

func init() {
	clientMetricsInit()
	serverMetricsInit()
}

// OMPMetric omp 监控项抽象（多值语义）。
type OMPMetric interface {
	// Group 监控分组。
	Group() MetricGroup

	// PointCount 监控点个数（单值监控项 1，多值监控项大于 1）。
	PointCount() int
	// PointName 第 i 个监控点的名称。
	PointName(i int) string
	// PointAggregation 第 i 个监控点的策略。
	PointAggregation(i int) Aggregation
	// PointValue 第 i 个监控点的值。
	PointValue(i int) float64

	// LabelCount 标签数量。
	LabelCount() int
	// LabelValue 第 i 个标签的值。
	LabelValue(i int) string
}
