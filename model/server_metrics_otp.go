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

var _ OTPMetric = (*ServerMetricsOTP)(nil)

// SetName 设置第 i 个数据的 name，（单值监控项 i 填 0）。
func (s *ServerMetricsOTP) SetName(i int, name string) {
	// ServerMetricsOTP 明确语义，不需要设置 name。
}

// SetAggregation 设置第 i 个数据的策略，（单值监控项 i 填 0）。
func (s *ServerMetricsOTP) SetAggregation(i int, a Aggregation) {
	// ServerMetricsOTP 明确语义，不需要设置数据策略。
}

// SetHistogram 设置第 i 个数据的 histogram 值，（单值监控项 i 填 0）。
func (s *ServerMetricsOTP) SetHistogram(i int, sum float64, count int64, ranges []string, counts []int64) {
	if i == ServerMetricHandledSecondsPoint {
		s.RpcServerHandledSeconds = NewHistogram(count, sum, counts, ranges)
	}
}

// SetAvg 设置第 i 个数据的 avg 值，（单值监控项 i 填 0）。
func (s *ServerMetricsOTP) SetAvg(i int, sum float64, count int64) {
	// ServerMetricsOTP 明确语义，只有 counter 和 histogram 点，不需要设置 avg。
}

// SetCount 设置第 i 个数据的 counter 值，（单值监控项 i 填 0）。
func (s *ServerMetricsOTP) SetCount(i int, count int64) {
	if i == ServerMetricStartedTotalPoint {
		s.RpcServerStartedTotal = int64(count)
	} else if i == ServerMetricHandledTotalPoint {
		s.RpcServerHandledTotal = int64(count)
	}
}

// SetValue 设置第 i 个数据的其他类型值（如：min、max、sum、set），（单值监控项 i 填 0）。
func (s *ServerMetricsOTP) SetValue(i int, value float64) {
	if i == ServerMetricStartedTotalPoint {
		s.RpcServerStartedTotal = int64(value)
	} else if i == ServerMetricHandledTotalPoint {
		s.RpcServerHandledTotal = int64(value)
	}
}
