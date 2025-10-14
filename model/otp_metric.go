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

// OTPMetric otp 监控项抽象。
// ClientMetricsOTP、ServerMetricsOTP、CustomMetricsOTP 是多值语义。
// NormalMetricOTP 是单值语义。
type OTPMetric interface {
	// SetName 设置第 i 个数据的 name，（单值监控项 i 填 0）。
	SetName(i int, name string)
	// SetAggregation 设置第 i 个数据的策略，（单值监控项 i 填 0）。
	SetAggregation(i int, a Aggregation)
	// SetHistogram 设置第 i 个数据的 histogram 值，（单值监控项 i 填 0）。
	SetHistogram(i int, sum float64, count int64, ranges []string, counts []int64)
	// SetAvg 设置第 i 个数据的 avg 值，（单值监控项 i 填 0）。
	SetAvg(i int, sum float64, count int64)
	// SetCount 设置第 i 个数据的 counter 值，（单值监控项 i 填 0）。
	SetCount(i int, c int64)
	// SetValue 设置第 i 个数据的其他类型值（如：min、max、sum、set），（单值监控项 i 填 0）。
	SetValue(i int, v float64)
}
