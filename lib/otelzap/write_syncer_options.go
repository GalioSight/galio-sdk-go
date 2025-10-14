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
	"time"

	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
)

// writeSyncerOptions defines the configuration for the various elements of BatchSyncer
type writeSyncerOptions struct {
	// maxQueueSize is the maximum queue size to buffer spans for delayed processing. If the
	// queue gets full it drops the spans. Use BlockOnQueueFull to change this behavior.
	// The default value of maxQueueSize is 2048.
	maxQueueSize int
	// batchTimeout is the maximum duration for constructing a batch. Processor
	// forcefully sends available spans when timeout is reached.
	// The default value of batchTimeout is 5000 msec.
	batchTimeout time.Duration
	// maxExportBatchSize is the maximum number of spans to process in a single batch.
	// If there are more than one batch worth of spans then it processes multiple batches
	// of spans one batch after the other without any delay.
	// The default value of maxExportBatchSize is 512.
	maxExportBatchSize int
	// blockOnQueueFull blocks onEnd() and onStart() method if the queue is full
	// AND if blockOnQueueFull is set to true.
	// Blocking option should be used carefully as it can severely affect the performance of an
	// application.
	blockOnQueueFull bool
	// writeLevel write_syncer 的日志级别。
	writeLevel model.WriteLevel
	// maxPacketSize is the maximum number of packet size that will forcefully trigger a batch process.
	// The deault value of maxPacketSize is 2M (in bytes) .
	maxPacketSize int
	// stats 自监控
	stats *model.SelfMonitorStats
	// log 自监控日志
	log *logs.Wrapper
	// exportToFile 是否导出到文件
	exportToFile bool
	// originalWriteLevel 如果设置采样必须上报，则 writer 会降级到 debug . 故此处记录原始的写入级别
	originalLevel string
}

func defaultWriteSyncerOptions() *writeSyncerOptions {
	o := &writeSyncerOptions{
		maxQueueSize:       1024 * 10,
		batchTimeout:       5000 * time.Millisecond,
		maxExportBatchSize: 512,
		blockOnQueueFull:   false,
		maxPacketSize:      1024 * 1024 * 2, // 当累积的日志大小到达 2MB 时，也要上传
		writeLevel:         model.WriteAll,
		stats:              &model.SelfMonitorStats{},
	}
	return o
}

// WriteSyncerOption apply changes to internalOptions.
type WriteSyncerOption func(o *writeSyncerOptions)

// WithMaxPacketSize WithMaxPacketSize
func WithMaxPacketSize(size int) WriteSyncerOption {
	return func(o *writeSyncerOptions) {
		if size > 0 {
			o.maxPacketSize = size
		}
	}
}

// WithStats 设置自监控对象。
func WithStats(stats *model.SelfMonitorStats) WriteSyncerOption {
	return func(o *writeSyncerOptions) {
		o.stats = stats
	}
}

// WithOriginalLevel 记录日志的原始写入级别 如 debug info error
func WithOriginalLevel(level string) WriteSyncerOption {
	return func(o *writeSyncerOptions) {
		o.originalLevel = level
	}
}

// WithMaxQueueSize return BatchSyncerOption which to set MaxQueueSize
func WithMaxQueueSize(size int) WriteSyncerOption {
	return func(o *writeSyncerOptions) {
		if size > 0 {
			o.maxQueueSize = size
		}
	}
}

// WithMaxExportBatchSize return BatchSyncerOption which to set  MaxExportBatchSize
func WithMaxExportBatchSize(size int) WriteSyncerOption {
	return func(o *writeSyncerOptions) {
		if size > 0 {
			o.maxExportBatchSize = size
		}
	}
}

// WithBatchTimeout return BatchSyncerOption which to set BatchTimeout
func WithBatchTimeout(delay time.Duration) WriteSyncerOption {
	return func(o *writeSyncerOptions) {
		if delay > 0 {
			o.batchTimeout = delay
		}
	}
}

// WithExportToFile 是否导出 log 数据到文件。
func WithExportToFile(b bool) WriteSyncerOption {
	return func(o *writeSyncerOptions) {
		o.exportToFile = b
	}
}

// WithLog 设置自监控日志对象。
func WithLog(log *logs.Wrapper) WriteSyncerOption {
	return func(o *writeSyncerOptions) {
		o.log = log
	}
}

// WithBlocking return BatchSyncerOption which to set BlockOnQueueFull
func WithBlocking() WriteSyncerOption {
	return func(o *writeSyncerOptions) {
		o.blockOnQueueFull = true
	}
}

// WithWriteLevel 设置 write syncer 写等级。
func WithWriteLevel(level model.WriteLevel) WriteSyncerOption {
	return func(o *writeSyncerOptions) {
		o.writeLevel = level
	}
}
