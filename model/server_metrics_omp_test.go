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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerMetrics(t *testing.T) {
	s := &ServerMetrics{}
	assert.Equal(t, "", s.PointName(0))

	s = &ServerMetrics{
		Metrics: []ServerMetrics_Metric{
			{
				Name:        ServerMetrics_rpc_server_started_total,
				Value:       1,
				Aggregation: Aggregation_AGGREGATION_COUNTER,
			},
		},
	}
	assert.Equal(t, "rpc_server_started_total", s.PointName(0))
	assert.Equal(t, Aggregation_AGGREGATION_COUNTER, s.PointAggregation(0))
	assert.Equal(t, float64(1), s.PointValue(0))
	assert.Equal(t, "", s.PointName(1))
	assert.Equal(t, Aggregation_AGGREGATION_NONE, s.PointAggregation(1))
	assert.Equal(t, float64(0), s.PointValue(1))

	s = GetServerMetrics(3)
	defer PutServerMetrics(s)
	s.RpcLabels.Fields[0].Name = RPCLabels_caller_container
	s.RpcLabels.Fields[0].Value = "caller_container_1"
	s.RpcLabels.Fields[1].Name = RPCLabels_callee_container
	s.RpcLabels.Fields[1].Value = "callee_container_2"
	s.RpcLabels.Fields[2].Name = RPCLabels_callee_con_setid
	s.RpcLabels.Fields[2].Value = "callee_set_3"
	s.Metrics[ServerMetricStartedTotalPoint].Value = 1
	s.Metrics[ServerMetricHandledTotalPoint].Value = 1
	s.Metrics[ServerMetricHandledSecondsPoint].Value = 0.1

	assert.Equal(t, ServerGroup, s.Group())
	assert.Equal(t, "rpc_server_started_total", s.PointName(0))
	assert.Equal(t, Aggregation_AGGREGATION_COUNTER, s.PointAggregation(0))
	assert.Equal(t, float64(1), s.PointValue(0))
	assert.Equal(t, 3, s.LabelCount())
	assert.Equal(t, "callee_set_3", s.LabelValue(2))
}
