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

package metrics

import (
	"galiosight.ai/galio-sdk-go/model"
)

type multiToOTPFunc func(m *multi, metrics *model.Metrics) int

// groupConfig 监控组配置。
type groupConfig struct {
	multiToOTP multiToOTPFunc
}

var groupConfigs = newGroupConfigs()

func newGroupConfigs() []groupConfig {
	configs := make([]groupConfig, model.MaxGroup)
	configs[model.ClientGroup] = groupConfig{
		multiToOTP: multiToClientMetricsOTP,
	}
	configs[model.ServerGroup] = groupConfig{
		multiToOTP: multiToServerMetricsOTP,
	}
	configs[model.NormalGroup] = groupConfig{
		multiToOTP: multiToNormalMetricsOTP,
	}
	configs[model.CustomGroup] = groupConfig{
		multiToOTP: multiToCustomMetricsOTP,
	}
	return configs
}

func getMultiToOTPFunc(extractor model.OMPMetric) multiToOTPFunc {
	group := extractor.Group()
	if group < 0 || int(group) >= len(groupConfigs) {
		return nil
	}
	return groupConfigs[group].multiToOTP
}

func multiToClientMetricsOTP(m *multi, metrics *model.Metrics) int {
	clientMetricsOTP := model.NewClientMetricsOTP()
	shallowCopyRPCLabels(m.rpcLabels, clientMetricsOTP.RpcLabels)
	exportCount := 0
	for i := range m.points {
		count, err := m.points[i].ToOTP(clientMetricsOTP, i)
		if err != nil {
			return 0
		}
		exportCount += count
	}
	if exportCount > 0 {
		metrics.ClientMetrics = append(metrics.ClientMetrics, clientMetricsOTP)
	}
	return exportCount
}

func getServerMetricsOTP() *model.ServerMetricsOTP {
	return &model.ServerMetricsOTP{RpcLabels: &model.RPCLabels{}}
}

func multiToServerMetricsOTP(m *multi, metrics *model.Metrics) int {
	serverMetricsOTP := getServerMetricsOTP()
	shallowCopyRPCLabels(m.rpcLabels, serverMetricsOTP.RpcLabels)
	exportCount := 0
	for i := range m.points {
		count, err := m.points[i].ToOTP(serverMetricsOTP, i)
		if err != nil {
			return 0
		}
		exportCount += count
	}
	if exportCount > 0 {
		metrics.ServerMetrics = append(metrics.ServerMetrics, serverMetricsOTP)
	}
	return exportCount
}

func getNormalMetricsOTP() *model.NormalMetricOTP {
	return &model.NormalMetricOTP{
		Metric: &model.MetricOTP{},
	}
}

func multiToNormalMetricsOTP(m *multi, metrics *model.Metrics) int {
	normalMetricsOTP := getNormalMetricsOTP()
	exportCount := 0
	for i := range m.points {
		count, err := m.points[i].ToOTP(normalMetricsOTP, i)
		if err != nil {
			return 0
		}
		exportCount += count
	}
	if exportCount > 0 {
		metrics.NormalMetrics = append(metrics.NormalMetrics, normalMetricsOTP)
	}
	return exportCount
}

func getCustomMetricsOTP(pointCount int) *model.CustomMetricsOTP {
	c := &model.CustomMetricsOTP{
		Metrics: make([]*model.MetricOTP, pointCount),
	}
	for i := 0; i < pointCount; i++ {
		c.Metrics[i] = &model.MetricOTP{}
	}
	return c
}

func multiToCustomMetricsOTP(m *multi, metrics *model.Metrics) int {
	customMetricsOTP := getCustomMetricsOTP(len(m.points)) // TODO(jaimeyang) 内存优化，get from pool。
	customMetricsOTP.CustomLabels = shallowCopyCustomLabels(m.customLabels, customMetricsOTP.CustomLabels)
	customMetricsOTP.MonitorName = m.monitorName
	exportCount := 0
	for i := range m.points {
		count, err := m.points[i].ToOTP(customMetricsOTP, i)
		if err != nil {
			return 0
		}
		exportCount += count
	}
	if exportCount > 0 {
		metrics.CustomMetrics = append(metrics.CustomMetrics, customMetricsOTP)
	}
	return exportCount
}

func shallowCopyRPCLabels(from, to *model.RPCLabels) {
	if from == nil || to == nil {
		return
	}
	if cap(to.Fields) < len(from.Fields) {
		to.Fields = make([]model.RPCLabels_Field, 0, len(from.Fields))
	}
	to.Fields = to.Fields[:len(from.Fields)]
	for i := range from.Fields {
		to.Fields[i].Name = from.Fields[i].Name
		to.Fields[i].Value = from.Fields[i].Value
	}
}

func shallowCopyCustomLabels(from, to []*model.Label) []*model.Label {
	if cap(to) < len(from) {
		to = make([]*model.Label, 0, len(from))
	}
	to = to[:len(from)]
	for i := range from {
		if to[i] == nil {
			to[i] = &model.Label{}
		}
		to[i].Name = from[i].Name
		to[i].Value = from[i].Value
	}
	return to
}
