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

func TestNormalMetric(t *testing.T) {
	n := &NormalMetric{}
	assert.Equal(t, "", n.PointName(0))

	n = GetNormalMetric()
	defer PutNormalMetric(n)
	n.Metric.Name = "TestNormalMetric"
	n.Metric.Aggregation = Aggregation_AGGREGATION_MAX
	n.Metric.Value = 1

	assert.Equal(t, NormalGroup, n.Group())
	assert.EqualValues(t, 1, n.PointCount())
	assert.Equal(t, "TestNormalMetric", n.PointName(0))
	assert.Equal(t, Aggregation_AGGREGATION_MAX, n.PointAggregation(0))
	assert.EqualValues(t, 1, n.PointValue(1))
}
