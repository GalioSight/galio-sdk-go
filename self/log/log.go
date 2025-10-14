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

// Package log 自监控日志
package log

import (
	"galiosight.ai/galio-sdk-go/lib/logs"
)

// SetLogger 修改伽利略自己的日志对象，用于自身的调试。
// Deprecated 当前自监控日志不允许修改成其他对象，只能调整日志级别。
// 使用 SetLogLevel 代替。
func SetLogger(log *logs.Wrapper) {
	logs.DefaultWrapper().SetLevel(log.GetLevel())
}

// SetLogLevel 设置自监控日志的级别。
func SetLogLevel(level string) {
	logs.DefaultWrapper().SetLevel(logs.ToLevel(level))
}

// SetLogPath 设置自监控日志的路径。
func SetLogPath(logPath string) {
	logs.DefaultWrapper().SetPath(logPath)
}

// addCallerSkip 增加调用栈
func addCallerSkip(skip int) *logs.Wrapper {
	return logs.DefaultWrapper().AddCallerSkip(skip)
}

// Debugf logs to DEBUG log. Arguments are handled in the manner of fmt.Printf.
func Debugf(format string, args ...interface{}) {
	if Enable(logs.LevelDebug) {
		addCallerSkip(1).Debugf(format, args...)
	}
}

// Infof logs to INFO log. Arguments are handled in the manner of fmt.Printf.
func Infof(format string, args ...interface{}) {
	if Enable(logs.LevelInfo) {
		addCallerSkip(1).Infof(format, args...)
	}
}

// Errorf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
func Errorf(format string, args ...interface{}) {
	if Enable(logs.LevelError) {
		addCallerSkip(1).Errorf(format, args...)
	}
}

// Enable 是否开启了对应级别的日志。
func Enable(level logs.Level) bool {
	return logs.DefaultWrapper().Enable(level)
}
