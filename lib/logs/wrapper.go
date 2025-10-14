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

// Package logs 封装自监控日志组件
package logs

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 定义了一个日志接口
type Logger interface {
	Debug(msg string, fields ...zapcore.Field)
	Info(msg string, fields ...zapcore.Field)
	Error(msg string, fields ...zapcore.Field)
}

// Wrapper 定义日志包装器
type Wrapper struct {
	Level  Level  // 用户设置的日志级别，用于提前判断，减少 fmt.Sprintf 的调用
	logger Logger // 使用接口替代具体的 zap.Logger
}

func newZapLogger(logPath string) *zap.Logger {
	w := zapcore.AddSync(
		&lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    5,
			MaxBackups: 1,
			LocalTime:  true,
		},
	)

	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeCaller = zapcore.FullCallerEncoder
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncodeLevel = zapcore.CapitalLevelEncoder

	// 默认 zap 的日志级别为 debug，依靠 Wrapper.Level 来控制日志级别
	level := zap.NewAtomicLevelAt(zap.DebugLevel)
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(cfg), w, level)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return logger
}

var wrapper = &Wrapper{
	Level:  LevelError,
	logger: newZapLogger("./galileo/galileo.log"),
}

// DefaultWrapper 返回默认的日志包装器
func DefaultWrapper() *Wrapper {
	return wrapper
}

// NopWrapper 返回一个不打印任何日志的包装器
func NopWrapper() *Wrapper {
	return &Wrapper{
		Level:  LevelNone,
		logger: zap.NewNop(),
	}
}

// NewWrapper 创建一个新的日志包装器，后面三个参数仅用于兼容历史代码，没有实际效果。此方法只能获得一个不同日志级别的 wrapper
// Deprecated 此方法不应该被调用，直接调用 DefaultWrapper 即可
func NewWrapper(level Level, debugf, infof, errorf func(format string, args ...interface{})) *Wrapper {
	return &Wrapper{
		Level:  level,
		logger: wrapper.logger,
	}
}

// Enable 判断是否开启了对应级别的日志
func (w *Wrapper) Enable(level Level) bool {
	return w != nil && w.Level <= level
}

// SetLevel 设置日志级别
func (w *Wrapper) SetLevel(level Level) {
	w.Level = level
}

// GetLevel 获取日志级别
func (w *Wrapper) GetLevel() Level {
	return w.Level
}

// SetPath 设置日志路径
func (w *Wrapper) SetPath(logPath string) {
	w.logger = newZapLogger(logPath)
}

// Sync 同步，立即输出日志，方便测试和观察
func (w *Wrapper) Sync() error {
	return w.logger.(*zap.Logger).Sync()
}

// AddCallerSkip 增加 CallerSkip。
func (w *Wrapper) AddCallerSkip(skip int) *Wrapper {
	return &Wrapper{
		Level:  w.Level,
		logger: w.logger.(*zap.Logger).WithOptions(zap.AddCallerSkip(skip)),
	}
}

// Debugf 打印调试级别的日志
func (w *Wrapper) Debugf(format string, args ...interface{}) {
	if w.Enable(LevelDebug) {
		w.logger.Debug(fmt.Sprintf(format, args...))
	}
}

// Infof 打印信息级别的日志
func (w *Wrapper) Infof(format string, args ...interface{}) {
	if w.Enable(LevelInfo) {
		w.logger.Info(fmt.Sprintf(format, args...))
	}
}

// Errorf 打印错误级别的日志
func (w *Wrapper) Errorf(format string, args ...interface{}) {
	if w.Enable(LevelError) {
		w.logger.Error(fmt.Sprintf(format, args...))
	}
}
