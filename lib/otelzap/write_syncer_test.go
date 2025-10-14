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
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	logpb "go.opentelemetry.io/proto/otlp/logs/v1"
)

type mockLogsExporter struct {
	data []*logpb.ResourceLogs
	mu   sync.Mutex
}

func (e *mockLogsExporter) ExportLogs(ctx context.Context, data []*logpb.ResourceLogs) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.data = append(e.data, data...)
	return nil
}

func (e *mockLogsExporter) Shutdown(ctx context.Context) error {
	return nil
}

func TestWriteSyncer(t *testing.T) {
	exporter := &mockLogsExporter{}
	w := NewWriteSyncer(exporter, nil, WithMaxExportBatchSize(10))
	data := []byte(`{"sampled":"true","level":"info","traceID":"12345678901234561234567890123456",
"spanID":"1234567812345678","msg":"test message","caller":"caller info","ts":1633099123.123,"abcde":"efg"}`)
	w.Write(data)

	// 调用 Sync 方法处理 rawData 通道中的数据
	err := w.Sync()
	if err != nil {
		t.Errorf("writeSyncer.Sync() returned error: %v", err)
	}

	assert.Equal(t, 1, len(exporter.data))
}

func TestWriteSyncerSync(t *testing.T) {
	const numRoutines = 10
	const numIterations = 100

	exporter := &mockLogsExporter{}
	batch := 10
	w := NewWriteSyncer(exporter, nil, WithMaxExportBatchSize(batch))
	data := []byte(`{"sampled":"true","level":"info","traceID":"12345678901234561234567890123456",
"spanID":"1234567812345678","msg":"test message","caller":"caller info","ts":1633099123.123,"abcde":"efg"}`)

	wg := sync.WaitGroup{}

	// 启动多个 goroutine 并发地向 rawData 通道发送数据
	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				w.Write(data)
			}
		}()
	}

	// 等待所有 goroutine 完成
	wg.Wait()

	// 调用 Sync 方法处理 rawData 通道中的数据
	err := w.Sync()
	assert.NoError(t, err)

	assert.Equal(t, numRoutines*numIterations/batch, len(exporter.data))
	assert.Equal(t, numRoutines*numIterations, logRecordCnt(exporter))
}

func logRecordCnt(exporter *mockLogsExporter) int {
	cnt := 0
	for _, log := range exporter.data {
		for _, slog := range log.ScopeLogs {
			cnt += len(slog.LogRecords)
		}
	}
	return cnt
}

func TestWriteSyncerPanic(t *testing.T) {
	w := NewWriteSyncer(nil, nil, WithMaxExportBatchSize(10))

	data := []byte(`{"sampled":"true","level":"info","traceID":"12345678901234561234567890123456",
"spanID":"1234567812345678","msg":"test message","caller":"caller info","ts":1633099123.123,"abcde":"efg"}`)
	w.Write(data)

	// 调用 Sync 方法处理 rawData 通道中的数据
	err := w.Sync()
	if err != nil {
		t.Errorf("writeSyncer.Sync() returned error: %v", err)
	}
}

func TestWriteSyncerFile(t *testing.T) {
	exporter := &mockLogsExporter{}
	w := NewWriteSyncer(exporter, nil, WithMaxExportBatchSize(10))
	w.options.exportToFile = true
	data := []byte(`{"sampled":"true","level":"info","traceID":"12345678901234561234567890123456",
"spanID":"1234567812345678","msg":"test message","caller":"caller info","ts":1633099123.123,"abcde":"efg"}`)
	w.Write(data)

	// 调用 Sync 方法处理 rawData 通道中的数据
	err := w.Sync()
	if err != nil {
		t.Errorf("writeSyncer.Sync() returned error: %v", err)
	}

	assert.Equal(t, 1, len(exporter.data))
}
