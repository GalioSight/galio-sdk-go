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

func TestCustomMetricsOTP(t *testing.T) {
	c := NewCustomMetricsOTP(0, 1)
	c.SetName(0, "TestCustomMetricsOTP")
	c.SetAggregation(0, Aggregation_AGGREGATION_COUNTER)
	c.SetCount(0, 1)

	assert.Equal(t, "TestCustomMetricsOTP", c.Metrics[0].Name)
	assert.Equal(t, Aggregation_AGGREGATION_COUNTER, c.Metrics[0].Aggregation)
	assert.Equal(t, NewOTPValue(1), c.Metrics[0].V)
}
