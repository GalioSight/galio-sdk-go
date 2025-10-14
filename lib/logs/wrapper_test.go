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

package logs

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

// mockLogger 是一个模拟的 Logger，用于测试
type mockLogger struct {
	lastMessage string
}

func (l *mockLogger) Debug(msg string, fields ...zapcore.Field) {
	l.lastMessage = msg
}

func (l *mockLogger) Info(msg string, fields ...zapcore.Field) {
	l.lastMessage = msg
}

func (l *mockLogger) Error(msg string, fields ...zapcore.Field) {
	l.lastMessage = msg
}

func TestWrapper_Debugf(t *testing.T) {
	logger := &mockLogger{}
	wrapper := &Wrapper{
		Level:  LevelDebug,
		logger: logger,
	}

	wrapper.Debugf("hello %s", "world")
	if logger.lastMessage != "hello world" {
		t.Errorf("unexpected message: %s", logger.lastMessage)
	}
}

func TestWrapper_Infof(t *testing.T) {
	logger := &mockLogger{}
	wrapper := &Wrapper{
		Level:  LevelInfo,
		logger: logger,
	}

	wrapper.Infof("hello %s", "world")
	if logger.lastMessage != "hello world" {
		t.Errorf("unexpected message: %s", logger.lastMessage)
	}
}

func TestWrapper_Errorf(t *testing.T) {
	logger := &mockLogger{}
	wrapper := &Wrapper{
		Level:  LevelError,
		logger: logger,
	}

	wrapper.Errorf("hello %s", "world")
	if logger.lastMessage != "hello world" {
		t.Errorf("unexpected message: %s", logger.lastMessage)
	}
}

func TestWrapper_SetLevel(t *testing.T) {
	logger := &mockLogger{}
	wrapper := &Wrapper{
		Level:  LevelDebug,
		logger: logger,
	}

	wrapper.SetLevel(LevelInfo)
	if wrapper.GetLevel() != LevelInfo {
		t.Errorf("unexpected Level: %v", wrapper.Level)
	}

	wrapper.SetLevel(LevelError)
	if wrapper.GetLevel() != LevelError {
		t.Errorf("unexpected Level: %v", wrapper.Level)
	}
}

func TestWrapper_SetPath(t *testing.T) {
	// 设置新的日志路径
	logPath := "./a/b.log"
	_ = os.Remove(logPath)
	w := &Wrapper{
		Level:  LevelError,
		logger: newZapLogger("./galileo/galileo.log"),
	}
	w.SetPath(logPath)
	w.Errorf("test")
	err := w.Sync()
	assert.NoError(t, err)
	assert.FileExists(t, logPath)
}

func TestWrapper_AddCaller(t *testing.T) {
	w := &Wrapper{
		Level:  LevelError,
		logger: newZapLogger("./galileo/galileo.log"),
	}
	w.AddCallerSkip(0)
	assert.Equal(t, LevelError, w.Level)
}

func TestWrapper(t *testing.T) {
	logPath := "./galileo/galileo.log"
	_ = os.Remove(logPath)
	w := NewWrapper(
		LevelDebug,
		func(format string, args ...interface{}) {
			fmt.Printf(format, args...)
		}, func(format string, args ...interface{}) {
			fmt.Printf(format, args...)
		}, func(format string, args ...interface{}) {
			fmt.Printf(format, args...)
		},
	)
	w.Debugf("test_debug\n")
	w.Infof("test_info\n")
	w.Errorf("test_error\n")

	d := DefaultWrapper()
	d.Debugf("default_debug\n")
	d.Infof("default_info\n")
	d.Errorf("default_error\n")

	n := NopWrapper()
	n.Debugf("nop_debug\n")
	n.Infof("nop_info\n")
	n.Errorf("nop_error\n")

	err := d.Sync()
	assert.NoError(t, err)

	assert.FileExists(t, logPath)
	// 读取日志文件内容
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	// 检查日志文件内容
	assert.Contains(t, string(content), "test_debug")
	assert.Contains(t, string(content), "test_info")
	assert.Contains(t, string(content), "test_error")

	assert.NotContains(t, string(content), "default_debug")
	assert.NotContains(t, string(content), "default_info")
	assert.Contains(t, string(content), "default_error")

	assert.NotContains(t, string(content), "nop_debug")
	assert.NotContains(t, string(content), "nop_info")
	assert.NotContains(t, string(content), "nop_error")

}
