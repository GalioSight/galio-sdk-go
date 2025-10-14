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

	"go.uber.org/atomic"
)

// StatNamePrefix 自监控 name 前缀。
const StatNamePrefix = "self_"

// MetricsStats 指标处理器统计。
type MetricsStats struct {
	// MultiCount 多值点个数（主调、被调、属性、自定义都算 1 个多值点）。
	MultiCount atomic.Int64 `aggregation:"AGGREGATION_SET"`
	// PointCount 单值点个数（多值点的单值点均统计在内，如：主调监控的 3 个点均统计在内）。
	PointCount atomic.Int64 `aggregation:"AGGREGATION_SET"`
	// ExportCount 最终导出个数，比如一个 point 是 histogram，导出个数就是 1+1+N。
	ExportCount atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	// ClearMultiCount 过期清理多值点个数。
	ClearMultiCount atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	// DiscardMultiCount 丢弃多值点个数（超量控制，拒绝写入后的统计）。
	DiscardMultiCount atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	// SelfMonitorError 自监控失败次数。
	SelfMonitorError atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	// SelfMonitorCount 自监控上报次数。
	SelfMonitorCount atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	// ReportErrorTotal 上报出现 error 的次数。
	ReportErrorTotal atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	// ReportHandledTotal 上报完成的次数
	ReportHandledTotal atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	// ReportErrorTotal 上报出现 error 的次数。
	ReportErrorRowsTotal atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	// ReportHandledTotal 上报完成的次数
	ReportHandledRowsTotal atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	// DoubleBufferChangeSlow 双 buffer 切换缓慢次数（正常应该是 0，出现非 0 需要关注）。
	DoubleBufferChangeSlow atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	// MaxPointCount 最大单值点（按分钟滑动窗口统计）。
	MaxPointCount atomic.Int64 `aggregation:"AGGREGATION_MAX"`
}

// TracesStats 追踪导出器统计。
type TracesStats struct {
	InitErrorTotal           atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	FailedExportCounter      atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	SucceededExportCounter   atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	BatchByCountCounter      atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	BatchByPacketSizeCounter atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	BatchByTimerCounter      atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	EnqueueCounter           atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	DropCounter              atomic.Int64 `aggregation:"AGGREGATION_COUNTER"` // 超过上报队列长度丢弃
	SucceededWriteByteSize   atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	FailedWriteByteSize      atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	WorkflowSampledCounter   atomic.Int64 `aggregation:"AGGREGATION_COUNTER"` // workflow 默认采样统计
	WorkflowBreakCounter     atomic.Int64 `aggregation:"AGGREGATION_COUNTER"` // workflow 默认采样触发熔断统计
	WorkflowPathCounter      atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	LimitDropCounter         atomic.Int64 `aggregation:"AGGREGATION_COUNTER"` // 熔断丢弃 span 统计
}

// LogsStats 日志导出器统计。
type LogsStats struct {
	InitErrorTotal         atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	FailedExportCounter    atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	SucceededExportCounter atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	EnqueueCounter         atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	DropCounter            atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	SucceededWriteByteSize atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	FailedWriteByteSize    atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	RawWriteByteSize       atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
}

// ProfilesStats 性能导出器统计
type ProfilesStats struct {
	InitErrorTotal         atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	FailedExportCounter    atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	SucceededExportCounter atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	EnqueueCounter         atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	DropCounter            atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	SucceededWriteByteSize atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	FailedWriteByteSize    atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
}

// PrometheusPushStats prometheus push 自监控指标
type PrometheusPushStats struct {
	InitErrorTotal         atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	FailedExportCounter    atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
	SucceededExportCounter atomic.Int64 `aggregation:"AGGREGATION_COUNTER"`
}

// SelfMonitorStats 监控统计。
type SelfMonitorStats struct {
	// MetricsStats 监控处理器统计。
	MetricsStats
	// TracesStats 追踪导出器统计。
	TracesStats
	// LogsStats 日志导出器统计。
	LogsStats
	// ProfilesStats 性能导出器统计
	ProfilesStats
	// PrometheusPushStats prometheus push 自监控指标
	PrometheusPushStats
}

// GetDeltaMetrics 获取增量自监控数据。
// 此方法不是线程安全的。
// lastStats 上次上报时的指标数据，用于和当前指标数据计算增量值。
// 每次调用完成后，lastStats 会更新成当前指标。
// 自监控上报调用，频率很低，允许反射调用，减少重复代码。
// 此方法只会被单线程定时调用，不会有并发调用。
// 每次上报自监控数据时，会和上次上报的数据计算增量。
func GetDeltaMetrics(
	last, current *SelfMonitorStats, target string,
) *Metrics {
	metrics := &Metrics{}
	_, lastValue := typeAndValue(last)
	curType, curValue := typeAndValue(current)
	groupCount := curType.NumField()
	for i := 0; i < groupCount; i++ {
		groupType := curType.Field(i)
		groupName := groupType.Name
		lastGroupValue := lastValue.Field(i)
		curGroupValue := curValue.Field(i)
		metricOTPS := buildGroupMetric(
			groupName, curGroupValue, lastGroupValue, groupType,
		)
		metrics.CustomMetrics = append(
			metrics.CustomMetrics, &CustomMetricsOTP{
				MonitorName: groupName,
				CustomLabels: []*Label{
					{
						"SdkTarget",
						target,
					},
				},
				Metrics: metricOTPS,
			},
		)
	}
	return metrics
}

func buildGroupMetric(
	groupName string, curGroupValue reflect.Value, lastGroupValue reflect.Value,
	groupType reflect.StructField,
) []*MetricOTP {
	numField := curGroupValue.NumField()
	metricOTPS := make([]*MetricOTP, numField)
	for j := 0; j < numField; j++ {
		curField := curGroupValue.Field(j)
		lastField := lastGroupValue.Field(j)
		field := groupType.Type.Field(j)
		aggregation := Aggregation(Aggregation_value[field.Tag.Get("aggregation")])
		metricOTPS[j] = &MetricOTP{
			Name: CustomName(groupName, field.Name, aggregation),
			V: &MetricOTP_Value{
				Value: getMetricValue(
					aggregation, curField, lastField,
				),
			},
			Aggregation: aggregation,
		}
	}
	return metricOTPS
}

// getMetricValue 获取指标值。
// 指标暂时只支持 2 种类型，
// set 类型直接取最新值。
// counter 类型，取增量值。
func getMetricValue(
	aggregation Aggregation, curField reflect.Value, lastField reflect.Value,
) float64 {
	switch aggregation {
	case Aggregation_AGGREGATION_SET, Aggregation_AGGREGATION_MAX:
		return get(curField)
	case Aggregation_AGGREGATION_COUNTER:
		return sub(curField, lastField)
	default:
		return 0
	}
}

func typeAndValue(stats *SelfMonitorStats) (reflect.Type, reflect.Value) {
	statsType := reflect.TypeOf(stats).Elem()
	statsValue := reflect.ValueOf(stats).Elem()
	return statsType, statsValue
}

// sub 通过反射计算 a-b。
// a,b 的原始类型必须是 atomic.Int64。
func sub(a, b reflect.Value) float64 {
	a1, ok1 := a.Interface().(atomic.Int64)
	b1, ok2 := b.Interface().(atomic.Int64)
	if ok1 && ok2 {
		return float64(a1.Load() - b1.Load())
	}
	return 0
}

// sub 获取数值
// a 的原始类型必须是 atomic.Int64。
func get(a reflect.Value) float64 {
	a1, ok1 := a.Interface().(atomic.Int64)
	if ok1 {
		return float64(a1.Load())
	}
	return 0
}
