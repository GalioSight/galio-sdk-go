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

import (
	"sync"
)

// metricsPool metric 对象池。
var metricsPool sync.Pool

// GetMetrics 获取对象。
func GetMetrics() *Metrics {
	if m, ok := metricsPool.Get().(*Metrics); ok {
		return m
	}
	return &Metrics{}
}

// PutMetrics 将对象返回到对象池。
func PutMetrics(m *Metrics) { // TODO(jaimeyang) 内存优化，put to pool，且 reset 所有子对象。
	m.TimestampMs = 0
	m.ClientMetrics = m.ClientMetrics[:0]
	m.ServerMetrics = m.ServerMetrics[:0]
	m.NormalMetrics = m.NormalMetrics[:0]
	m.CustomMetrics = m.CustomMetrics[:0]
	metricsPool.Put(m)
}

// AddNormalMetric 增加属性监控。
func (m *Metrics) AddNormalMetric(name string, aggregation Aggregation, v float64) {
	n := NewNormalMetricsOTP()
	n.SetName(0, name)
	n.SetAggregation(0, aggregation)
	switch aggregation {
	case Aggregation_AGGREGATION_COUNTER:
		n.SetCount(0, int64(v))
		m.NormalMetrics = append(m.NormalMetrics, n)
	case Aggregation_AGGREGATION_SET:
		n.SetValue(0, v)
		m.NormalMetrics = append(m.NormalMetrics, n)
	default:
	}
}

// AddCustomMetric 增加自定义监控。
func (m *Metrics) AddCustomMetric(monitorName, metricName string, aggregation Aggregation, v float64, lvs ...string) {
	labelCount := len(lvs) / 2
	c := NewCustomMetricsOTP(labelCount, 1)
	c.SetMonitorName(monitorName)
	c.SetName(0, metricName)
	c.SetAggregation(0, aggregation)
	for i := 0; i < labelCount; i++ {
		addCustomMetricLabel(c, lvs, i)
	}
	switch aggregation {
	case Aggregation_AGGREGATION_SET:
		c.SetValue(0, v)
		m.CustomMetrics = append(m.CustomMetrics, c)
	default:
	}
}

func addCustomMetricLabel(c *CustomMetricsOTP, lvs []string, i int) {
	idx := i * 2
	if idx+1 >= len(lvs) {
		return
	}
	c.CustomLabels[i].Name = lvs[idx]
	c.CustomLabels[i].Value = lvs[idx+1]
}
