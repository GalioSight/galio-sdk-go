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

func TestServerMetricsOTP(t *testing.T) {
	s := &ServerMetricsOTP{}
	s.SetCount(ServerMetricStartedTotalPoint, 10)
	s.SetCount(ServerMetricHandledTotalPoint, 10)
	s.SetHistogram(ServerMetricHandledSecondsPoint, 50, 10, []string{"000...001", "002...003"}, []int64{5, 5})
	assert.EqualValues(t, 10, s.RpcServerStartedTotal)
	assert.EqualValues(t, 10, s.RpcServerHandledTotal)

	assert.EqualValues(t, 50, s.RpcServerHandledSeconds.Sum)
	assert.EqualValues(t, 10, s.RpcServerHandledSeconds.Count)
	assert.Equal(t, 2, len(s.RpcServerHandledSeconds.Buckets))
}
