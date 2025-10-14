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
	"strings"

	"galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/model"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func parseLevel(text string) zapcore.Level {
	if level, err := zapcore.ParseLevel(text); err == nil {
		return level
	}
	if strings.ToLower(text) == "trace" {
		return zapcore.DebugLevel
	}
	return zapcore.InfoLevel
}

// NewLogger 获取日志对象
func NewLogger(cfg *configs.Logs, options ...zap.Option) (*zap.Logger, error) {
	// cfg.Log 是 selflog，所以 cfg.Log.Level 是 selflog 的 Level，这里不应该使用它来初始化 zl
	zl := zap.NewAtomicLevelAt(parseLevel(cfg.Processor.GetLevel()))
	core, err := newZapCore(cfg, zl)
	if err != nil {
		return nil, err
	}
	options = append(options, toZapOptions(cfg)...)
	logger := zap.New(core, options...)
	return logger, nil
}

// 额外设置 WithContextSampleLevel
func toZapOptions(cfg *configs.Logs) []zap.Option {
	var ret []zap.Option
	logsProcessor := &cfg.Processor
	coreType := model.IOCore
	strategy := DefaultStrategy
	if logsProcessor.MustLogTraced {
		coreType = model.SampleCore
		strategy |= MustLogTraced
	}
	if logsProcessor.OnlyTraceLog {
		coreType = model.SampleCore
		strategy |= OnlyTraceLog
	}
	if coreType == model.SampleCore {
		// 只有当 cfg 配置了 MustLogTraced or OnlyTraceLog 时，这段代码才生效，即才会使用配置中的 LogTracedType 字段
		opt := WithContextSampleLevel(strategy)
		if logsProcessor.LogTracedType == string(model.LogTracedDyeing) {
			opt = WithContextDyeingLevel(strategy)
		}
		return append(ret, opt)
	}
	return ret
}
