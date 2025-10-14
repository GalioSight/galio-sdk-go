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

var _ OMPMetric = (*ClientMetrics)(nil)

const (
	// ClientMetricStartedTotalPoint 客户端（主调方上报）发出的请求量。
	ClientMetricStartedTotalPoint = int(ClientMetrics_rpc_client_started_total)
	// ClientMetricHandledTotalPoint 客户端（主调方上报）处理完成的请求量。
	ClientMetricHandledTotalPoint = int(ClientMetrics_rpc_client_handled_total)
	// ClientMetricHandledSecondsPoint 客户端（主调方上报）处理完成的耗时分布，单位：秒。
	ClientMetricHandledSecondsPoint = int(ClientMetrics_rpc_client_handled_seconds)
	// ClientMetricPointCount 客户端（主调方上报）监控数。
	ClientMetricPointCount = int(ClientMetrics_rpc_client_metrics_point_count)
)

// NewClientMetrics 构造 *ClientMetrics，标签长度 labelCount。
func NewClientMetrics(labelCount int) *ClientMetrics {
	c := &ClientMetrics{
		Metrics:   make([]ClientMetrics_Metric, ClientMetricPointCount),
		RpcLabels: RPCLabels{},
	}
	c.RpcLabels.grow(labelCount)
	for i := range c.Metrics {
		c.Metrics[i].Name = ClientMetrics_MetricName(i)
	}
	c.Metrics[ClientMetricStartedTotalPoint].Aggregation = Aggregation_AGGREGATION_COUNTER
	c.Metrics[ClientMetricHandledTotalPoint].Aggregation = Aggregation_AGGREGATION_COUNTER
	c.Metrics[ClientMetricHandledSecondsPoint].Aggregation = Aggregation_AGGREGATION_HISTOGRAM
	return c
}

// Group 监控分组。
func (c *ClientMetrics) Group() MetricGroup {
	return ClientGroup
}

// PointCount 监控点个数（单值监控项 1，多值监控项大于 1）。
func (c *ClientMetrics) PointCount() int {
	return len(c.Metrics)
}

// clientMetricsPointNames 枚举到字符串映射表，如：0 => rpc_client_started_total
var clientMetricsPointNames []string

// clientMetricsInit 初始化一次就行。
func clientMetricsInit() {
	clientMetricsPointNames = make([]string, ClientMetricPointCount)
	for i := range clientMetricsPointNames {
		clientMetricsPointNames[i] = ClientMetrics_MetricName_name[int32(i)]
	}
}

// PointName 第 i 个监控点的名称。
func (c *ClientMetrics) PointName(i int) string {
	if i < 0 || i >= c.PointCount() {
		return ""
	}
	name := int(c.Metrics[i].Name)
	if name < 0 || name >= len(clientMetricsPointNames) {
		return ""
	}
	return clientMetricsPointNames[name]
}

// PointAggregation 第 i 个监控点的策略。
func (c *ClientMetrics) PointAggregation(i int) Aggregation {
	if i < 0 || i >= c.PointCount() {
		return Aggregation_AGGREGATION_NONE
	}
	return c.Metrics[i].Aggregation
}

// PointValue 第 i 个监控点的值。
func (c *ClientMetrics) PointValue(i int) float64 {
	if i < 0 || i >= c.PointCount() {
		return 0
	}
	return c.Metrics[i].Value
}

// LabelCount 标签数量。
func (c *ClientMetrics) LabelCount() int {
	return len(c.RpcLabels.Fields)
}

// LabelValue 第 i 个标签的值。
func (c *ClientMetrics) LabelValue(i int) string {
	if i < 0 || i >= c.LabelCount() {
		return ""
	}
	return c.RpcLabels.Fields[i].Value
}
