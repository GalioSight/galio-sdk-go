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

//go:build apitest
// +build apitest

package metrics

import (
	"testing"
	"time"

	"galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func Test_otp_api_test(t *testing.T) {
	exporter, err := NewExporter(
		&configs.Metrics{
			Log: logs.DefaultWrapper(),
			Exporter: model.MetricsExporter{
				Protocol:  "otp",
				Collector: model.Collector{Addr: ""},

				ThreadCount:   0,
				BufferSize:    0,
				WindowSeconds: 1,
				PageSize:      0,
			},
		},
	)
	assert.Nil(t, err)
	exporter.Export(proto.Clone(data).(*model.Metrics))
	// 等 5 秒，让数据异步报完，期望数据在 5 秒内报完
	time.Sleep(time.Second * 5)
	stats := exporter.(*metricsExporter).stats
	assert.Equal(t, int64(0), stats.ReportErrorTotal.Load())
	assert.Equal(t, int64(1), stats.ReportHandledTotal.Load())
}
