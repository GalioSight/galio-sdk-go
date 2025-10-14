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
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		text  string
		level zapcore.Level
	}{
		{"trace", zapcore.DebugLevel},
		{"Trace", zapcore.DebugLevel},
		{"trace", zapcore.DebugLevel},
		{"", zapcore.InfoLevel},
		{"bad", zapcore.InfoLevel},
		{"debug", zapcore.DebugLevel},
		{"Debug", zapcore.DebugLevel},
		{"info", zapcore.InfoLevel},
		{"fatal", zapcore.FatalLevel},
	}

	a := assert.New(t)
	for _, test := range tests {
		t.Run(
			"", func(t *testing.T) {
				a.Equal(test.level, parseLevel(test.text))
			},
		)
	}
}
