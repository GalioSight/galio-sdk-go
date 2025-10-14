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

func TestCustomMetrics(t *testing.T) {
	c := &CustomMetrics{}
	assert.Equal(t, "", c.PointName(0))

	c = &CustomMetrics{
		Metrics: []Metric{
			{
				Name:        "TestCustomMetrics_1",
				Value:       1,
				Aggregation: Aggregation_AGGREGATION_MAX,
			},
		},
	}
	assert.Equal(t, "TestCustomMetrics_1", c.PointName(0))
	assert.Equal(t, Aggregation_AGGREGATION_MAX, c.PointAggregation(0))
	assert.Equal(t, float64(1), c.PointValue(0))
	assert.Equal(t, "", c.PointName(1))
	assert.Equal(t, Aggregation_AGGREGATION_NONE, c.PointAggregation(1))
	assert.Equal(t, float64(0), c.PointValue(1))

	c = GetCustomMetrics(1, 1)
	defer PutCustomMetrics(c)
	c.CustomLabels[0].Name = "TestCustomMetrics_label1"
	c.CustomLabels[0].Value = "TestCustomMetrics_value1"
	c.Metrics[0].Value = 1
	c.Metrics[0].Aggregation = Aggregation_AGGREGATION_MIN
	c.Metrics[0].Name = "TestCustomMetrics_metric1"

	assert.Equal(t, CustomGroup, c.Group())
	assert.Equal(t, "TestCustomMetrics_metric1", c.PointName(0))
	assert.Equal(t, Aggregation_AGGREGATION_MIN, c.PointAggregation(0))
	assert.Equal(t, float64(1), c.PointValue(0))
	assert.Equal(t, 1, c.LabelCount())
	assert.Equal(t, "TestCustomMetrics_value1", c.LabelValue(0))
}
