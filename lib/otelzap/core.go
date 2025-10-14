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
	"errors"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	baseconfigs "galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/helper"
	expres "galiosight.ai/galio-sdk-go/internal/resource"
	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/self/metric"
)

// newZapCore 创建日志 core 对象。
func newZapCore(baseLogsCfg *baseconfigs.Logs, zl zap.AtomicLevel) (zapcore.Core, error) {
	res := &baseLogsCfg.Resource
	metric.Init(res, baseLogsCfg.SelfMonitor, baseLogsCfg.Log, metric.WithAPIKey(baseLogsCfg.APIKey))
	exporter, err := helper.GetLogsExporter(baseLogsCfg) // 假设这里已经创建了带 schemaURL 的 exporter
	if err != nil {
		return nil, errors.New("NewLogsExporter error: " + err.Error())
	}

	// core 选项。
	logsExporter := baseLogsCfg.Exporter
	core := newCore(
		NewWriteSyncer(
			exporter,
			expres.GenResource(baseLogsCfg.SchemaURL, res, expres.SchemaTypeLog),
			toSyncerOptions(logsExporter, baseLogsCfg.Log)...,
		), zl,
	)
	return core, err
}

// toSyncerOptions 根据日志配置的 exporter 等转换成 writerSyncer 的选项
func toSyncerOptions(
	logsExporter model.LogsExporter,
	logger *logs.Wrapper,
) []WriteSyncerOption {
	var options = []WriteSyncerOption{
		WithStats(metric.GetSelfMonitor().Stats),
	}
	options = append(
		options,
		WithMaxQueueSize(int(logsExporter.BufferSize)),
		WithMaxExportBatchSize(int(logsExporter.PageSize)),
		WithMaxPacketSize(int(logsExporter.PacketSize)),
		WithBatchTimeout(time.Duration(logsExporter.WindowSeconds)*time.Second),
		WithExportToFile(logsExporter.ExportToFile),
		WithLog(logger),
	)
	return options
}

// newCore 构造 zap core 实例。
func newCore(syncer zapcore.WriteSyncer, zl zap.AtomicLevel) zapcore.Core {
	// 默认增加 ctxCore，吃掉_ctx field
	return &ctxCore{Core: zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig()), syncer, zl)}
}

func encoderConfig() zapcore.EncoderConfig {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeCaller = zapcore.FullCallerEncoder
	return cfg
}
