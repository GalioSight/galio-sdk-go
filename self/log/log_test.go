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

package log

import (
	"os"
	"testing"

	"galiosight.ai/galio-sdk-go/lib/logs"
	"github.com/stretchr/testify/assert"
)

func TestDebugf(t *testing.T) {
	Debugf("test Debugf")
}

func TestInfof(t *testing.T) {
	Infof("test Infof")
}

func TestErrorf(t *testing.T) {
	Errorf("test Errorf")
}

func TestSetLogger(t *testing.T) {
	SetLogger(logs.NopWrapper())
	logPath := "./a/b.log"
	_ = os.Remove(logPath)
	SetLogPath(logPath)
	SetLogLevel("debug")
	Debugf("test_debug")
	assert.Equal(t, "debug", logs.LevelStrings[logs.DefaultWrapper().GetLevel()])
	assert.True(t, Enable(logs.LevelDebug))
	assert.FileExists(t, logPath)
	// 读取日志文件内容
	content, err := os.ReadFile(logPath)
	assert.Nil(t, err)
	assert.Contains(t, string(content), "test_debug")
}
