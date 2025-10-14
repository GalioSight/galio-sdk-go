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

// Package otelzap ...
package otelzap

import (
	"go.uber.org/zap/zapcore"
)

// coreStrategy core 日志策略
type coreStrategy int8

const (
	DefaultStrategy coreStrategy = 0b00000000 // 默认策略，导出符合级别的所有日志
	OnlyTraceLog    coreStrategy = 0b00000001 // 级别内只打命中了 trace 采样的日志
	MustLogTraced   coreStrategy = 0b00000010 // 命中了 trace 的日志突破级别
)

// MatchCore 是简化的 sampleCore，不固定构造实现，只执行具体的策略
type MatchCore struct {
	zapcore.Core
	matched  bool
	strategy coreStrategy
}

// NewMatchCore 方便外部复用
func NewMatchCore(c zapcore.Core, matched bool, strategy coreStrategy) *MatchCore {
	return &MatchCore{c, matched, strategy}
}

// Enabled 根据传入的日志级别和自身采样标识判断是否支持。
func (c *MatchCore) Enabled(level zapcore.Level) bool {
	switch c.strategy {
	case OnlyTraceLog: // 只导出采样的日志，未突破日志级别
		return c.Core.Enabled(level) && c.matched
	case MustLogTraced: // 导出符合级别的所有日志，且采样日志突破级别
		return c.Core.Enabled(level) || c.matched
	case OnlyTraceLog | MustLogTraced: // 只导出采样的日志，且采样日志突破级别
		return c.matched
	default: // 默认导出符合级别的所有日志
		return c.Core.Enabled(level)
	}
}

func (c *MatchCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *MatchCore) With(fields []zapcore.Field) zapcore.Core {
	return &MatchCore{Core: c.Core.With(fields), matched: c.matched, strategy: c.strategy}
}
