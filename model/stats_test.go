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
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
)

func TestMetricsStats_GetDeltaMetrics(t *testing.T) {
	stats := &SelfMonitorStats{}
	lastStats := &SelfMonitorStats{}
	target := "a.b.c"
	// 新增加字段在 expected 和 inc 函数里面都添加下。
	// 这里明确的列出期望，避免误删除导致也断言成功
	expected := []*CustomMetricsOTP{
		{
			MonitorName: "MetricsStats", CustomLabels: []*Label{{"SdkTarget", target}}, Metrics: []*MetricOTP{
				{
					Name: "custom_gauge_MetricsStats_MultiCount_set", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_SET,
				}, {
					Name: "custom_gauge_MetricsStats_PointCount_set", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_SET,
				}, {
					Name: "custom_counter_MetricsStats_ExportCount_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_MetricsStats_ClearMultiCount_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_MetricsStats_DiscardMultiCount_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_MetricsStats_SelfMonitorError_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_MetricsStats_SelfMonitorCount_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_MetricsStats_ReportErrorTotal_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_MetricsStats_ReportHandledTotal_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_MetricsStats_ReportErrorRowsTotal_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_MetricsStats_ReportHandledRowsTotal_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_MetricsStats_DoubleBufferChangeSlow_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_gauge_MetricsStats_MaxPointCount_max", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_MAX,
				},
			},
		}, {
			MonitorName: "TracesStats", CustomLabels: []*Label{{"SdkTarget", target}}, Metrics: []*MetricOTP{
				{
					Name: "custom_counter_TracesStats_InitErrorTotal_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_TracesStats_FailedExportCounter_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_TracesStats_SucceededExportCounter_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_TracesStats_BatchByCountCounter_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_TracesStats_BatchByPacketSizeCounter_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_TracesStats_BatchByTimerCounter_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_TracesStats_EnqueueCounter_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_TracesStats_DropCounter_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_TracesStats_SucceededWriteByteSize_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_TracesStats_FailedWriteByteSize_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_TracesStats_WorkflowSampledCounter_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_TracesStats_WorkflowBreakCounter_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_TracesStats_WorkflowPathCounter_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_TracesStats_LimitDropCounter_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				},
			},
		}, {
			MonitorName: "LogsStats", CustomLabels: []*Label{{"SdkTarget", target}}, Metrics: []*MetricOTP{
				{
					Name: "custom_counter_LogsStats_InitErrorTotal_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_LogsStats_FailedExportCounter_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_LogsStats_SucceededExportCounter_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_LogsStats_EnqueueCounter_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_LogsStats_DropCounter_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_LogsStats_SucceededWriteByteSize_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_LogsStats_FailedWriteByteSize_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_LogsStats_RawWriteByteSize_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				},
			},
		}, {
			MonitorName: "ProfilesStats", CustomLabels: []*Label{{"SdkTarget", target}}, Metrics: []*MetricOTP{
				{
					Name: "custom_counter_ProfilesStats_InitErrorTotal_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_ProfilesStats_FailedExportCounter_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_ProfilesStats_SucceededExportCounter_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_ProfilesStats_EnqueueCounter_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_ProfilesStats_DropCounter_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_ProfilesStats_SucceededWriteByteSize_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_ProfilesStats_FailedWriteByteSize_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				},
			},
		}, {
			MonitorName: "PrometheusPushStats", CustomLabels: []*Label{{"SdkTarget", target}}, Metrics: []*MetricOTP{
				{
					Name: "custom_counter_PrometheusPushStats_InitErrorTotal_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_PrometheusPushStats_FailedExportCounter_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				}, {
					Name: "custom_counter_PrometheusPushStats_SucceededExportCounter_total", V: NewOTPValue(1),
					Aggregation: Aggregation_AGGREGATION_COUNTER,
				},
			},
		},
	}
	for ii := 0; ii < 3; ii++ {
		t.Run(
			strconv.Itoa(ii), func(t *testing.T) {
				inc(stats)
				curStats := *stats
				metrics := GetDeltaMetrics(lastStats, &curStats, target)
				lastStats = &curStats
				actual := metrics.CustomMetrics
				require.Equal(t, len(expected), len(actual))
				for i := 0; i < len(expected); i++ {
					require.Equal(t, len(expected[i].Metrics), len(actual[i].Metrics))
					for j := 0; j < len(expected[i].Metrics); j++ {
						require.Equal(t, expected[i].Metrics[j].V, actual[i].Metrics[j].V)
						require.Equal(t, expected[i].Metrics[j].Name, actual[i].Metrics[j].Name)
					}
					require.Equal(t, expected[i], actual[i])
				}
				require.Equal(t, expected, actual)
			},
		)
	}
}

func inc(stats *SelfMonitorStats) {
	stats.MultiCount.Store(1)
	stats.PointCount.Store(1)
	stats.ExportCount.Inc()
	stats.ClearMultiCount.Inc()
	stats.DiscardMultiCount.Inc()
	stats.SelfMonitorError.Inc()
	stats.SelfMonitorCount.Inc()
	stats.ReportErrorTotal.Inc()
	stats.ReportHandledTotal.Inc()
	stats.ReportErrorRowsTotal.Inc()
	stats.ReportHandledRowsTotal.Inc()
	stats.DoubleBufferChangeSlow.Inc()
	stats.MaxPointCount.Store(1)
	inc64 := func(v *atomic.Int64) {
		v.Inc()
	}

	walk(&stats.TracesStats, inc64)
	walk(&stats.LogsStats, inc64)
	walk(&stats.ProfilesStats, inc64)
	walk(&stats.PrometheusPushStats, inc64)
}

func walk(stats interface{}, cb func(*atomic.Int64)) {
	rv := reflect.ValueOf(stats)
	for i := 0; i < rv.Elem().NumField(); i++ {
		cb(rv.Elem().Field(i).Addr().Interface().(*atomic.Int64))
	}
}
