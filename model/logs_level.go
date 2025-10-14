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
	logpb "go.opentelemetry.io/proto/otlp/logs/v1"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogsLevel is a logging priority. Higher levels are more important.
type LogsLevel string

const (
	// TraceLevel A fine-grained debugging event. Typically disabled in default configurations.
	TraceLevel LogsLevel = "TRACE"
	// DebugLevel A debugging event.
	DebugLevel LogsLevel = "DEBUG"
	// InfoLevel An informational event. Indicates that an event happened.
	InfoLevel LogsLevel = "INFO"
	// WarnLevel A warning event. Not an error but is likely more important than an informational event.
	WarnLevel LogsLevel = "WARN"
	// ErrorLevel An error event. Something went wrong.
	ErrorLevel LogsLevel = "ERROR"
	// FatalLevel A fatal error such as application or system crash.
	FatalLevel LogsLevel = "FATAL"
)

// ToZapCoreLevel 字符串日志等级转 zap 日志等级。
func ToZapCoreLevel(level LogsLevel) zapcore.Level {
	switch level {
	case TraceLevel:
		return zap.DebugLevel
	case DebugLevel:
		return zap.DebugLevel
	case InfoLevel:
		return zap.InfoLevel
	case WarnLevel:
		return zap.WarnLevel
	case ErrorLevel:
		return zap.ErrorLevel
	case FatalLevel:
		return zap.FatalLevel
	default:
		return zap.ErrorLevel
	}
}

// ToSeverityNumber 字符串日志等级转 OpenTelemetry 日志等级。
func ToSeverityNumber(level LogsLevel) logpb.SeverityNumber {
	var number logpb.SeverityNumber
	switch level {
	case TraceLevel:
		number = logpb.SeverityNumber_SEVERITY_NUMBER_TRACE
	case DebugLevel:
		number = logpb.SeverityNumber_SEVERITY_NUMBER_DEBUG
	case InfoLevel:
		number = logpb.SeverityNumber_SEVERITY_NUMBER_INFO
	case WarnLevel:
		number = logpb.SeverityNumber_SEVERITY_NUMBER_WARN
	case ErrorLevel:
		number = logpb.SeverityNumber_SEVERITY_NUMBER_ERROR
	case FatalLevel:
		number = logpb.SeverityNumber_SEVERITY_NUMBER_FATAL
	}
	return number
}
