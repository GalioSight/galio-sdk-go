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

package runtimes

import (
	"fmt"
	"testing"
	"time"

	"galiosight.ai/galio-sdk-go/model"
	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	metrics := &model.Metrics{}
	metrics.TimestampMs = time.Now().Unix() * 1000
	Write(metrics)
	fmt.Println(metrics)
	// TODO 后续完善断言。
}

func TestWriteGalileoMetric(t *testing.T) {
	metrics := &model.Metrics{}
	metrics.TimestampMs = time.Now().Unix() * 1000
	WriteGalileoMetric(metrics)
	assert.NotEqual(t, len(metrics.CustomMetrics), 0)
}
