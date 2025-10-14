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

package self

import (
	"testing"

	logs2 "galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
	selfmetric "galiosight.ai/galio-sdk-go/self/metric"
	"github.com/stretchr/testify/assert"
)

func TestSetupObserver(t *testing.T) {
	SetupObserver(
		&model.Resource{Target: "a"},
		logs2.DefaultWrapper(),
		model.SelfMonitor{
			Collector: model.Collector{
				Addr: "b",
			},
		},
		"c",
		selfmetric.WithAPIKey("d"),
	)
	assert.Equal(t, logs2.LevelError, logs2.DefaultWrapper().GetLevel())
}
