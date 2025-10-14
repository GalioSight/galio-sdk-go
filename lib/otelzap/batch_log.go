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
	"encoding/hex"
	"errors"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	common "go.opentelemetry.io/proto/otlp/common/v1"
	logs "go.opentelemetry.io/proto/otlp/logs/v1"
	resource "go.opentelemetry.io/proto/otlp/resource/v1"

	"galiosight.ai/galio-sdk-go/self/metric"
)

// batchLog 结构体用于存储日志记录和相关信息
type batchLog struct {
	records       []*logs.LogRecord // 存储日志记录的切片
	recordByteCnt int               // 存储日志记录的总字节数
	mu            sync.Mutex        // 互斥锁，确保线程安全
	maxRecordCnt  int               // 最大日志记录数
	maxByteCnt    int               // 最大字节计数
}

// add 添加一行日志
// 此方法是线程安全的。
func (b *batchLog) add(data []byte) {
	log, err := convertToLogRecord(data)
	if err != nil {
		return
	}
	b.mu.Lock()
	b.records = append(b.records, log)
	b.recordByteCnt += len(data)
	b.mu.Unlock()
}

// hasBatch 判断是否有一批足够的数据用于上报了。
// 此方法是线程安全的。
func (b *batchLog) hasBatch() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.records) >= b.maxRecordCnt {
		metric.GetSelfMonitor().Stats.BatchByCountCounter.Inc()
		return true
	}
	if b.recordByteCnt >= b.maxByteCnt {
		metric.GetSelfMonitor().Stats.BatchByPacketSizeCounter.Inc()
		return true
	}
	return false
}

// extractAndResetResourceLogs 提取 batchLog 中的日志用于上报，并重置 batchLog 的日志。
// 此方法是线程安全的。
// 提取完之后，batchLog 中的日志会清空。
func (b *batchLog) extractAndResetResourceLogs(resPB *resource.Resource, schemaURL string) ([]*logs.ResourceLogs, int, int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	resourceLogs := []*logs.ResourceLogs{
		{
			Resource: resPB,
			ScopeLogs: []*logs.ScopeLogs{
				{
					LogRecords: b.records,
				},
			},
			SchemaUrl: schemaURL,
		},
	}
	recordCnt := len(b.records)
	recordBytes := b.recordByteCnt
	b.records = b.records[:0]
	b.recordByteCnt = 0
	return resourceLogs, recordCnt, recordBytes
}

// Write 实现 Write 接口
func convertToLogRecord(data []byte) (*logs.LogRecord, error) {
	// bytes -> JSON
	iter := jsoniter.ConfigFastest.BorrowIterator(data)
	defer jsoniter.ConfigFastest.ReturnIterator(iter)
	// JSON -> LogRecord
	logRecord, err := convertToRecordV2(iter)
	if err != nil {
		return nil, err
	}
	if ok := isValidRecord(logRecord); !ok {
		return nil, errors.New("invalid traceID spanID")
	}
	return logRecord, nil
}

const (
	fieldSampled = "sampled"
	fieldLevel   = "level"
	fieldTraceID = "traceID"
	fieldSpanID  = "spanID"
	trueString   = "true"
)

func convertToRecordV2(iter *jsoniter.Iterator) (*logs.LogRecord, error) {
	logRecord := &logs.LogRecord{}
	iter.ReadObjectCB(
		func(iterator *jsoniter.Iterator, f string) bool {
			switch f {
			case fieldSampled:
				logRecord.Attributes = append(
					logRecord.Attributes, &common.KeyValue{
						Key: fieldSampled,
						Value: &common.AnyValue{
							Value: &common.AnyValue_BoolValue{
								BoolValue: iter.ReadString() == trueString,
							},
						},
					},
				)
			case fieldLevel:
				logRecord.SeverityText = iter.ReadString()
			case "msg":
				logRecord.Body = &common.AnyValue{
					Value: &common.AnyValue_StringValue{
						StringValue: iter.ReadString(),
					},
				}
			case fieldTraceID:
				logRecord.TraceId, _ = hex.DecodeString(iter.ReadString())
			case fieldSpanID:
				logRecord.SpanId, _ = hex.DecodeString(iter.ReadString())
			case "caller":
				logRecord.Attributes = append(
					logRecord.Attributes, &common.KeyValue{
						Key: "line",
						Value: &common.AnyValue{
							Value: &common.AnyValue_StringValue{
								StringValue: iter.ReadString(),
							},
						},
					},
				)
			case "ts":
				logRecord.TimeUnixNano = uint64(iter.ReadFloat64() * float64(time.Second))
			default:
				// 支持 trpc 0.9.0 任意类型的 log field.
				logRecord.Attributes = append(
					logRecord.Attributes, &common.KeyValue{
						Key: f,
						Value: &common.AnyValue{
							Value: &common.AnyValue_StringValue{
								StringValue: iter.ReadAny().ToString(),
							},
						},
					},
				)
			}
			return true
		},
	)
	return logRecord, iter.Error
}

func isValidRecord(logRecord *logs.LogRecord) bool {
	if len(logRecord.GetTraceId()) != 16 && len(logRecord.GetTraceId()) != 0 {
		return false
	}
	if len(logRecord.GetSpanId()) != 8 && len(logRecord.GetSpanId()) != 0 {
		return false
	}
	return true
}
