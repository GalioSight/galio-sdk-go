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

var _ OMPMetric = (*ServerMetrics)(nil)

const (
	// ServerMetricStartedTotalPoint 服务端（被调方上报）接收到的请求量。
	ServerMetricStartedTotalPoint = int(ServerMetrics_rpc_server_started_total)
	// ServerMetricHandledTotalPoint 服务端（被调方上报）处理完成的请求量。
	ServerMetricHandledTotalPoint = int(ServerMetrics_rpc_server_handled_total)
	// ServerMetricHandledSecondsPoint 服务端（被调方上报）处理完成的耗时分布，单位：秒。
	ServerMetricHandledSecondsPoint = int(ServerMetrics_rpc_server_handled_seconds)
	// ServerMetricPointCount 服务端（被调方上报）监控数。
	ServerMetricPointCount = int(ServerMetrics_rpc_server_metrics_point_count)
)

// NewServerMetrics 构造 *ServerMetrics，标签长度 labelCount。
func NewServerMetrics(labelCount int) *ServerMetrics {
	s := &ServerMetrics{
		Metrics:   make([]ServerMetrics_Metric, ServerMetricPointCount),
		RpcLabels: RPCLabels{},
	}
	s.RpcLabels.grow(labelCount)
	for i := range s.Metrics {
		s.Metrics[i].Name = ServerMetrics_MetricName(i)
	}
	s.Metrics[ServerMetricStartedTotalPoint].Aggregation = Aggregation_AGGREGATION_COUNTER
	s.Metrics[ServerMetricHandledTotalPoint].Aggregation = Aggregation_AGGREGATION_COUNTER
	s.Metrics[ServerMetricHandledSecondsPoint].Aggregation = Aggregation_AGGREGATION_HISTOGRAM
	return s
}

// Group 监控分组。
func (s *ServerMetrics) Group() MetricGroup {
	return ServerGroup
}

// PointCount 数据点个数（单值监控项 1，多值监控项大于 1）。
func (s *ServerMetrics) PointCount() int {
	return len(s.Metrics)
}

// serverMetricsPointNames 枚举到字符串映射表，如：0 => rpc_server_started_total
var serverMetricsPointNames []string

// serverMetricsInit 初始化一次就行。
func serverMetricsInit() {
	serverMetricsPointNames = make([]string, ServerMetricPointCount)
	for i := range serverMetricsPointNames {
		serverMetricsPointNames[i] = ServerMetrics_MetricName_name[int32(i)]
	}
}

// PointName 第 i 个监控点的名称。
func (s *ServerMetrics) PointName(i int) string {
	if i < 0 || i >= s.PointCount() {
		return ""
	}
	name := int(s.Metrics[i].Name)
	if name < 0 || name >= len(serverMetricsPointNames) {
		return ""
	}
	return serverMetricsPointNames[name]
}

// PointAggregation 第 i 个监控点的策略。
func (s *ServerMetrics) PointAggregation(i int) Aggregation {
	if i < 0 || i >= s.PointCount() {
		return Aggregation_AGGREGATION_NONE
	}
	return s.Metrics[i].Aggregation
}

// PointValue 第 i 个监控点的值。
func (s *ServerMetrics) PointValue(i int) float64 {
	if i < 0 || i >= s.PointCount() {
		return 0
	}
	return s.Metrics[i].Value
}

// LabelCount 标签数量。
func (s *ServerMetrics) LabelCount() int {
	return len(s.RpcLabels.Fields)
}

// LabelValue 第 i 个标签的值。
func (s *ServerMetrics) LabelValue(i int) string {
	if i < 0 || i >= s.LabelCount() {
		return ""
	}
	return s.RpcLabels.Fields[i].Value
}
