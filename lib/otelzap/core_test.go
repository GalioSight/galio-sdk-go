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

package otelzap

import (
	"testing"

	baseconfigs "galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/self/metric"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewZapCore(t *testing.T) {
	res := model.Resource{
		Target:   "a.b.c",
		TenantId: "debug",
	}
	baseLogsCfg := &baseconfigs.Logs{
		Log:       logs.NopWrapper(),
		Resource:  res,
		Processor: model.LogsProcessor{},
		Exporter: model.LogsExporter{
			Protocol:      "otlp",
			WindowSeconds: 1,
		},
		SelfMonitor: model.SelfMonitor{
			ReportSeconds: 1,
		},
		Stats: metric.GetSelfMonitor().Stats,
	}
	zl := zap.NewAtomicLevelAt(zapcore.Level(baseLogsCfg.Log.GetLevel()))
	_, err := newZapCore(baseLogsCfg, zl)
	assert.Nil(t, err)
	assert.Nil(t, err)
	baseLogsCfg.Processor.MustLogTraced = true
	_, err = newZapCore(baseLogsCfg, zl)
	assert.Nil(t, err)
}
