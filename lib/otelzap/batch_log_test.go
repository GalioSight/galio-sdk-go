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
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	v1 "go.opentelemetry.io/proto/otlp/logs/v1"
	v12 "go.opentelemetry.io/proto/otlp/resource/v1"
)

func TestBatchLog(t *testing.T) {
	b := &batchLog{
		maxRecordCnt: 10,
		maxByteCnt:   1000,
	}

	// 测试 add 方法
	data := []byte(`{"sampled":"true","level":"info","traceID":"12345678901234561234567890123456","spanID":"1234567812345678","msg":"test message","caller":"caller info","ts":1633099123.123}`)
	b.add(data)
	assert.Equal(t, 1, len(b.records))
	assert.Equal(t, len(data), b.recordByteCnt)

	// 测试 hasBatch 方法
	assert.False(t, b.hasBatch())

	// 添加足够的记录以满足 maxRecordCnt 条件
	for i := 0; i < 9; i++ {
		b.add(data)
	}
	assert.True(t, b.hasBatch())

	// 测试 extractAndResetResourceLogs 方法
	resPB := &v12.Resource{}
	resourceLogs, recordCnt, recordBytes := b.extractAndResetResourceLogs(resPB, "")
	assert.Equal(t, 10, len(resourceLogs[0].ScopeLogs[0].LogRecords))
	assert.Equal(t, 10, recordCnt)
	assert.Equal(t, len(data)*10, recordBytes)

	// 确保 batchLog 已重置
	assert.Equal(t, 0, len(b.records))
	assert.Equal(t, 0, b.recordByteCnt)
}

func TestConvertLogRecord(t *testing.T) {
	// 正确的 JSON 数据
	data := []byte(`{"sampled":"true","level":"info","traceID":"12345678901234561234567890123456",
"spanID":"1234567812345678","msg":"test message","caller":"caller info","ts":1633099123.123,"abcde":"efg"}`)
	logRecord, err := convertToLogRecord(data)
	assert.NoError(t, err)
	assert.NotNil(t, logRecord)

	// traceID 不正确的 JSON 数据
	data = []byte(`{"sampled":"true","level":"info","traceID":"1234567890123456123456789012345",
"spanID":"1234567812345678","msg":"test message","caller":"caller info","ts":1633099123.123,"abcde":"efg"}`)
	logRecord, err = convertToLogRecord(data)
	assert.Error(t, err)
	assert.Nil(t, logRecord)

	// 错误的 JSON 数据
	data = []byte(`{"sampled":"true","level":"info","traceID":"123456789012345","spanID":"12345678","msg":"test message","caller":"caller info","ts":1633099123.123}`)
	logRecord, err = convertToLogRecord(data)
	assert.Error(t, err)
	assert.Nil(t, logRecord)
}

func TestConvertToRecordV2(t *testing.T) {
	// 正确的 JSON 数据
	data := []byte(`{"sampled":"true","level":"info","traceID":"12345678901234561234567890123456","spanID":"1234567812345678","msg":"test message","caller":"caller info","ts":1633099123.123}`)
	iter := jsoniter.ConfigFastest.BorrowIterator(data)
	defer jsoniter.ConfigFastest.ReturnIterator(iter)
	logRecord, err := convertToRecordV2(iter)
	assert.NoError(t, err)
	assert.NotNil(t, logRecord)

	// 错误的 JSON 数据
	data = []byte(`\\{"sampled":"true","level":"info","traceID":"123456789012345","spanID":"12345678",
"msg":"test message","caller":"caller info","ts":1633099123.123}`)
	iter = jsoniter.ConfigFastest.BorrowIterator(data)
	defer jsoniter.ConfigFastest.ReturnIterator(iter)
	logRecord, err = convertToRecordV2(iter)
	assert.Error(t, err)
	assert.Equal(t, &v1.LogRecord{}, logRecord)
}

func TestIsValidRecord(t *testing.T) {
	logRecord := &v1.LogRecord{
		TraceId: []byte("1234567890123456"),
		SpanId:  []byte("12345678"),
	}
	assert.True(t, isValidRecord(logRecord))

	logRecord = &v1.LogRecord{
		TraceId: []byte("123456789012345"),
		SpanId:  []byte("12345678"),
	}
	assert.False(t, isValidRecord(logRecord))

	logRecord = &v1.LogRecord{
		TraceId: []byte("1234567890123456"),
		SpanId:  []byte("1234567"),
	}
	assert.False(t, isValidRecord(logRecord))
}
