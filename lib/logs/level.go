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

// Package logs 日志组件
package logs

import (
	"strings"
)

// Level is the log Level.
type Level int

// Enums log Level constants.
const (
	LevelDebug Level = iota
	LevelInfo
	LevelError
	LevelNone
)

// LevelStrings is the map from log Level to its string representation.
var LevelStrings = map[Level]string{
	LevelDebug: "debug",
	LevelInfo:  "info",
	LevelError: "error",
	LevelNone:  "none",
}

// LevelNames is the map from string to log Level.
var LevelNames = map[string]Level{
	"debug": LevelDebug,
	"info":  LevelInfo,
	"error": LevelError,
	"none":  LevelNone,
}

// ToLevel 将字符串转成 Level 类型。
func ToLevel(s string) Level {
	level, ok := LevelNames[strings.ToLower(s)]
	if !ok {
		level = LevelNone
	}
	return level
}
