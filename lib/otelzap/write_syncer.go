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

	"go.opentelemetry.io/otel/sdk/resource"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logpb "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	"go.uber.org/zap/zapcore"

	"galiosight.ai/galio-sdk-go/components"
	"galiosight.ai/galio-sdk-go/debug"
	"galiosight.ai/galio-sdk-go/lib/file"
	"galiosight.ai/galio-sdk-go/lib/timer"
	selflog "galiosight.ai/galio-sdk-go/self/log"
	"galiosight.ai/galio-sdk-go/self/metric"
)

var _ zapcore.WriteSyncer = (*writeSyncer)(nil)

type writeSyncer struct {
	exporter     components.LogsExporter
	options      *writeSyncerOptions
	res          *resource.Resource
	resPB        *resourcepb.Resource
	schemaURL    string
	timer        *timer.SafeTimer
	rawData      chan []byte
	batch        batchLog
	fileExporter *file.Exporter
	debugger     debug.UTF8Debugger
	syncRequest  chan struct{}
	syncResponse chan struct{}
}

// NewWriteSyncer  return BatchWriteSyncer
func NewWriteSyncer(
	exporter components.LogsExporter, res *resource.Resource,
	opts ...WriteSyncerOption,
) *writeSyncer {
	o := defaultWriteSyncerOptions()
	for _, opt := range opts {
		opt(o)
	}
	o.log.Infof("[galileo]NewWriteSyncer|res=%+v,opts=%+v\n", res, o)
	w := &writeSyncer{
		exporter:     exporter,
		options:      o,
		res:          res,
		resPB:        nil,
		timer:        timer.NewSafeTimer(o.batchTimeout),
		rawData:      make(chan []byte, o.maxQueueSize),
		batch:        batchLog{maxRecordCnt: o.maxExportBatchSize, maxByteCnt: o.maxPacketSize},
		fileExporter: file.NewExporter(o.exportToFile, "galileo/logs", o.log),
		debugger:     debug.NewUTF8Debugger(),
		syncRequest:  make(chan struct{}, 1),
		syncResponse: make(chan struct{}, 1),
	}
	w.setResource(res)

	go w.processQueue()
	return w
}

func (w *writeSyncer) setResource(res *resource.Resource) {
	if res.Len() == 0 {
		return
	}
	resPB := &resourcepb.Resource{}
	for _, kv := range res.Attributes() {
		resPB.Attributes = append(
			resPB.Attributes, &commonpb.KeyValue{
				Key: string(kv.Key),
				Value: &commonpb.AnyValue{
					Value: &commonpb.AnyValue_StringValue{StringValue: kv.Value.Emit()},
				},
			},
		)
	}
	w.resPB = resPB
	w.schemaURL = res.SchemaURL()
}

// processQueue 消费队列中的日志数据。
// 此方法是线程安全的。
// 当前只运行了一个线程，必要时改成多个线程。
func (w *writeSyncer) processQueue() {
	for {
		select {
		case <-w.timer.C():
			metric.GetSelfMonitor().Stats.BatchByTimerCounter.Inc()
			w.export()
		case data := <-w.rawData:
			w.processRecord(data)
		case <-w.syncRequest:
			w.processAllRecords()
			w.export()
			w.syncResponse <- struct{}{}
		}
	}
}

// processRecord 处理一条日志。
// 此方法是线程安全的。
func (w *writeSyncer) processRecord(data []byte) {
	w.batch.add(data)
	if w.batch.hasBatch() {
		w.export()
	}
}

// export 将日志数据上报到伽利略平台。
// 此方法是线程安全的。
func (w *writeSyncer) export() {
	w.timer.Reset(w.options.batchTimeout)
	resourceLogs, recordCnt, recordBytes := w.batch.extractAndResetResourceLogs(w.resPB, w.schemaURL)
	if recordCnt > 0 {
		err := w.exportBatch(resourceLogs)
		if err != nil {
			if w.debugger.Enabled() {
				w.debugger.DebugLogsInvalidUTF8(err, resourceLogs)
			}
			w.options.log.Errorf("writeSyncer ExportLogs err=%v", err)
			w.options.stats.LogsStats.FailedExportCounter.Add(int64(recordCnt))
			w.options.stats.LogsStats.FailedWriteByteSize.Add(int64(recordBytes))
		} else {
			w.options.stats.LogsStats.SucceededExportCounter.Add(int64(recordCnt))
			w.options.stats.LogsStats.SucceededWriteByteSize.Add(int64(recordBytes))
		}
	}
}

func (w *writeSyncer) exportBatch(resourceLogs []*logpb.ResourceLogs) error {
	defer func() {
		if r := recover(); r != nil {
			selflog.Errorf("Recovered from panic: %v\n", r)
		}
	}()
	if w.options.exportToFile {
		w.fileExporter.Export(resourceLogs)
	}
	err := w.exporter.ExportLogs(context.Background(), resourceLogs)
	return err
}

// Write 实现 Write 接口。
// 由于 data 在 Write 函数退出之后，会被主调方 free 掉，所以需要深拷贝。
// 此处可以将 data 解析成 LogRecord 在放队列，但是解析 JSON 耗时较长，会阻塞用户日志线程，直接复制会快很多。
func (w *writeSyncer) Write(data []byte) (n int, err error) {
	metric.GetSelfMonitor().Stats.LogsStats.RawWriteByteSize.Add(int64(len(data)))
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	if w.options.blockOnQueueFull {
		w.rawData <- dataCopy
		return len(data), nil
	}

	select {
	case w.rawData <- dataCopy:
	default:
		w.options.stats.LogsStats.DropCounter.Inc()
		w.options.log.Infof("[galileo]otelzap writeSyncer Enqueue dropped")
	}
	return len(data), nil
}

// Sync 实现 Sync 接口，将当前队列中的数据全部导出到伽利略平台。
// 此方法是线程安全的。
// 在进程退出的时候，可以调用此方法保证异步日志上报完成。
// 此方法可能会消费较长的时间。
// 在时延要求高的情况下，不能调用此方法。
// 注意 Write 接口是异步的，Sync 只能把当前队列中的日志全部上报。
// 如果此时又有新的异步日志进入队列，新日志可能不会立即上报到伽利略。
// 所以如果要确保某条日志一定上报，需要在这条日志后面调用 Sync 方法。
func (w *writeSyncer) Sync() error {
	w.syncRequest <- struct{}{}
	<-w.syncResponse
	return nil
}

// processAllRecords 处理队列中的所有日志。
// 此方法是线程安全的。
func (w *writeSyncer) processAllRecords() {
	for {
		select {
		case data := <-w.rawData:
			w.processRecord(data)
		default:
			return
		}
	}
}
