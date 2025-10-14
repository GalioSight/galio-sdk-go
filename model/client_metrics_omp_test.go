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

func TestClientMetrics(t *testing.T) {
	c := &ClientMetrics{}
	assert.Equal(t, "", c.PointName(0))

	c = &ClientMetrics{
		Metrics: []ClientMetrics_Metric{
			{
				Name:        ClientMetrics_rpc_client_started_total,
				Value:       1,
				Aggregation: Aggregation_AGGREGATION_COUNTER,
			},
		},
	}
	assert.Equal(t, "rpc_client_started_total", c.PointName(0))
	assert.Equal(t, Aggregation_AGGREGATION_COUNTER, c.PointAggregation(0))
	assert.Equal(t, float64(1), c.PointValue(0))
	assert.Equal(t, "", c.PointName(1))
	assert.Equal(t, Aggregation_AGGREGATION_NONE, c.PointAggregation(1))
	assert.Equal(t, float64(0), c.PointValue(1))

	c = GetClientMetrics(3)
	defer PutClientMetrics(c)
	c.RpcLabels.Fields[0].Name = RPCLabels_caller_container
	c.RpcLabels.Fields[0].Value = "forwardContainer"
	c.RpcLabels.Fields[1].Name = RPCLabels_callee_container
	c.RpcLabels.Fields[1].Value = "collectorContainer"
	c.RpcLabels.Fields[2].Name = RPCLabels_callee_con_setid
	c.RpcLabels.Fields[2].Value = "collectorSet"
	c.Metrics[ClientMetricStartedTotalPoint].Value = 1
	c.Metrics[ClientMetricHandledTotalPoint].Value = 1
	c.Metrics[ClientMetricHandledSecondsPoint].Value = 0.1

	assert.Equal(t, ClientGroup, c.Group())
	assert.Equal(t, "rpc_client_started_total", c.PointName(0))
	assert.Equal(t, Aggregation_AGGREGATION_COUNTER, c.PointAggregation(0))
	assert.Equal(t, float64(1), c.PointValue(0))
	assert.Equal(t, 3, c.LabelCount())
	assert.Equal(t, "collectorSet", c.LabelValue(2))
}
