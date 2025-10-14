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

package traces

import (
	"context"
	"runtime"
	"strings"
	"sync"
	"time"

	"galiosight.ai/galio-sdk-go/debug"
	"galiosight.ai/galio-sdk-go/exporters/otlp/traces/tracetransform"
	"galiosight.ai/galio-sdk-go/lib/file"
	"galiosight.ai/galio-sdk-go/lib/logs"
	selflog "galiosight.ai/galio-sdk-go/self/log"
	"galiosight.ai/galio-sdk-go/self/metric"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/atomic"
)

// Defaults for BatchSpanProcessorOptions.
const (
	// DefaultMaxQueueSize 默认最大 queue size
	DefaultMaxQueueSize = 2048
	// DefaultBatchTimeout 默认批处理定时时间
	DefaultBatchTimeout = 5000 * time.Millisecond
	// DefaultExportTimeout 默认最大上报超时时间
	DefaultExportTimeout = 30000 * time.Millisecond
	// DefaultMaxExportBatchSize 默认最大上传批大小
	DefaultMaxExportBatchSize = 512
	// DefaultMaxBatchedPacketSize 触发大包上传默认值
	DefaultMaxBatchedPacketSize = 2097152
)

// BatchSpanProcessorOption BatchSpanProcessor Option helper
type BatchSpanProcessorOption func(o *BatchSpanProcessorOptions)

// BatchSpanProcessorOptions BatchSpanProcessor 控制项
type BatchSpanProcessorOptions struct {
	// MaxQueueSize is the maximum queue size to buffer spans for delayed processing. If the
	// queue gets full it drops the spans. Use BlockOnQueueFull to change this behavior.
	// The default value of MaxQueueSize is 2048.
	MaxQueueSize int

	// BatchTimeout is the maximum duration for constructing a batch. Processor
	// forcefully sends available spans when timeout is reached.
	// The default value of BatchTimeout is 5000 msec.
	BatchTimeout time.Duration

	// ExportTimeout specifies the maximum duration for exporting spans. If the timeout
	// is reached, the export will be cancelled.
	// The default value of ExportTimeout is 30000 msec.
	ExportTimeout time.Duration

	// MaxExportBatchSize is the maximum number of spans to process in a single batch.
	// If there are more than one batch worth of spans then it processes multiple batches
	// of spans one batch after the other without any delay.
	// The default value of MaxExportBatchSize is 512.
	MaxExportBatchSize int

	// MaxPacketSize is the maximum number of packet size that will forcefully trigger a batch process.
	// The deault value of MaxPacketSize is 2M (in bytes) .
	MaxPacketSize int

	// BlockOnQueueFull blocks onEnd() and onStart() method if the queue is full
	// AND if BlockOnQueueFull is set to true.
	// Blocking option should be used carefully as it can severely affect the performance of an
	// application.
	BlockOnQueueFull bool

	// log 自监控日志
	log *logs.Wrapper
	// exportToFile 是否导出到文件
	exportToFile bool
}

// batchSpanProcessor is a SpanProcessor that batches asynchronously-received
// spans and sends them to a trace.Exporter when complete.
type batchSpanProcessor struct {
	e sdktrace.SpanExporter
	o BatchSpanProcessorOptions

	queue       chan sdktrace.ReadOnlySpan
	dropped     atomic.Int64
	batchedSize int

	batch        []sdktrace.ReadOnlySpan
	batchMutex   sync.Mutex
	timer        *time.Timer
	stopWait     sync.WaitGroup
	stopOnce     sync.Once
	stopCh       chan struct{}
	fileExporter *file.Exporter
	debugger     debug.UTF8Debugger
}

var _ sdktrace.SpanProcessor = (*batchSpanProcessor)(nil)

// NewBatchSpanProcessor creates a new SpanProcessor that will send completed
// span batches to the exporter with the supplied options.
//
// If the exporter is nil, the span processor will preform no action.
func NewBatchSpanProcessor(
	exporter sdktrace.SpanExporter, options ...BatchSpanProcessorOption,
) sdktrace.SpanProcessor {
	o := BatchSpanProcessorOptions{
		BatchTimeout:       DefaultBatchTimeout,
		ExportTimeout:      DefaultExportTimeout,
		MaxQueueSize:       DefaultMaxQueueSize,
		MaxExportBatchSize: DefaultMaxExportBatchSize,
		MaxPacketSize:      DefaultMaxBatchedPacketSize,
	}
	for _, opt := range options {
		opt(&o)
	}
	bsp := &batchSpanProcessor{
		e:            exporter,
		o:            o,
		batch:        make([]sdktrace.ReadOnlySpan, 0, o.MaxExportBatchSize),
		timer:        time.NewTimer(o.BatchTimeout),
		queue:        make(chan sdktrace.ReadOnlySpan, o.MaxQueueSize),
		stopCh:       make(chan struct{}),
		fileExporter: file.NewExporter(o.exportToFile, "galileo/traces", o.log),
		debugger:     debug.NewUTF8Debugger(),
	}

	bsp.stopWait.Add(1)
	go func() {
		defer bsp.stopWait.Done()
		bsp.processQueue()
		bsp.drainQueue()
	}()

	return bsp
}

// OnStart method does nothing.
func (bsp *batchSpanProcessor) OnStart(
	parent context.Context, s sdktrace.ReadWriteSpan,
) {
}

// OnEnd method enqueues a ReadOnlySpan for later processing.
func (bsp *batchSpanProcessor) OnEnd(s sdktrace.ReadOnlySpan) {
	// Do not enqueue spans if we are just going to drop them.
	if bsp.e == nil {
		return
	}
	bsp.enqueue(s)
}

// Shutdown flushes the queue and waits until all spans are processed.
// It only executes once. Subsequent call does nothing.
func (bsp *batchSpanProcessor) Shutdown(ctx context.Context) error {
	var err error
	bsp.stopOnce.Do(
		func() {
			wait := make(chan struct{})
			go func() {
				close(bsp.stopCh)
				bsp.stopWait.Wait()
				if bsp.e != nil {
					if err1 := bsp.e.Shutdown(ctx); err1 != nil {
						otel.Handle(err1)
					}
				}
				close(wait)
			}()
			// Wait until the wait group is done or the context is cancelled
			select {
			case <-wait:
			case <-ctx.Done():
				err = ctx.Err()
			}
		},
	)
	return err
}

type forceFlushSpan struct {
	sdktrace.ReadOnlySpan
	flushed chan struct{}
}

// SpanContext spanContext
func (f forceFlushSpan) SpanContext() trace.SpanContext {
	return trace.NewSpanContext(trace.SpanContextConfig{TraceFlags: trace.FlagsSampled})
}

// ForceFlush exports all ended spans that have not yet been exported.
func (bsp *batchSpanProcessor) ForceFlush(ctx context.Context) error {
	var err error
	if bsp.e != nil {
		flushCh := make(chan struct{})
		if bsp.enqueueBlockOnQueueFull(
			ctx, forceFlushSpan{flushed: flushCh}, true,
		) {
			select {
			case <-flushCh:
				// Processed any items in queue prior to ForceFlush being called
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		wait := make(chan error)
		go func() {
			wait <- bsp.exportSpans(ctx)
			close(wait)
		}()
		// Wait until the export is finished or the context is cancelled/timed out
		select {
		case err = <-wait:
		case <-ctx.Done():
			err = ctx.Err()
		}
	}
	return err
}

// WithMaxQueueSize 设置 MaxQueueSize helper
func WithMaxQueueSize(size int) BatchSpanProcessorOption {
	return func(o *BatchSpanProcessorOptions) {
		if size > 0 {
			o.MaxQueueSize = size
		}
	}
}

// WithMaxExportBatchSize 设置 MaxExportBatchSize helper
func WithMaxExportBatchSize(size int) BatchSpanProcessorOption {
	return func(o *BatchSpanProcessorOptions) {
		if size > 0 {
			o.MaxExportBatchSize = size
		}
	}
}

// WithBatchTimeout 设置 BatchTimeout helper
func WithBatchTimeout(delay time.Duration) BatchSpanProcessorOption {
	return func(o *BatchSpanProcessorOptions) {
		if delay > 0 {
			o.BatchTimeout = delay
		}
	}
}

// WithExportTimeout 设置上报超时时间 helper
func WithExportTimeout(timeout time.Duration) BatchSpanProcessorOption {
	return func(o *BatchSpanProcessorOptions) {
		if timeout > 0 {
			o.ExportTimeout = timeout
		}
	}
}

// WithBlocking 设置 Blocking helper
func WithBlocking() BatchSpanProcessorOption {
	return func(o *BatchSpanProcessorOptions) {
		o.BlockOnQueueFull = true
	}
}

// WithMaxPacketSize 设置 MaxPacketSize helper
func WithMaxPacketSize(size int) BatchSpanProcessorOption {
	return func(o *BatchSpanProcessorOptions) {
		if size > 0 {
			o.MaxPacketSize = size
		}
	}
}

// WithExportToFile 是否导出配置到文件。
func WithExportToFile(b bool) BatchSpanProcessorOption {
	return func(o *BatchSpanProcessorOptions) {
		o.exportToFile = b
	}
}

// WithLog 设置自监控日志对象。
func WithLog(log *logs.Wrapper) BatchSpanProcessorOption {
	return func(o *BatchSpanProcessorOptions) {
		o.log = log
	}
}

// exportSpans is a subroutine of processing and draining the queue.
func (bsp *batchSpanProcessor) exportSpans(ctx context.Context) error {
	bsp.timer.Reset(bsp.o.BatchTimeout)

	bsp.batchMutex.Lock()
	defer bsp.batchMutex.Unlock()

	if bsp.o.ExportTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, bsp.o.ExportTimeout)
		defer cancel()
	}

	if l := len(bsp.batch); l > 0 {
		size := int64(len(bsp.batch))
		batchedSize := int64(bsp.batchedSize)
		err := bsp.exportBatch(ctx)
		bsp.handleError(err)
		// A new batch is always created after exporting, even if the batch failed to be exported.
		//
		// It is up to the exporter to implement any type of retry logic if a batch is failing
		// to be exported, since it is specific to the protocol and backend being sent to.
		bsp.batch = bsp.batch[:0]
		bsp.batchedSize = 0
		tracesStats := &metric.GetSelfMonitor().Stats.TracesStats
		if err != nil {
			tracesStats.FailedExportCounter.Add(size)
			tracesStats.FailedWriteByteSize.Add(batchedSize)
			return err
		}
		tracesStats.SucceededExportCounter.Add(size)
		tracesStats.SucceededWriteByteSize.Add(batchedSize)
	}
	return nil
}

// handleError 处理错误。
// 重点处理 errInvalidUTF8 错误。
// 此错误来源于：google.golang.org/protobuf@v1.33.0/internal/impl/codec_field.go 里面的 errInvalidUTF8
// 当出现 errInvalidUTF8 错误时，我们切换成标准模式，这样后续的上报数据都会进行 UTF-8 转义，避免继续失败。
// 但是只处理 body。如果用户自定义添加的 tag 有非法字符，还是会继续上报失败，需要用户主动修改代码。
func (bsp *batchSpanProcessor) handleError(err error) {
	if err != nil {
		selflog.Errorf("[galileo]exportSpans, err=%v", err)
		if strings.Contains(err.Error(), "string field contains invalid UTF-8") {
			SetSonicFastest(false)
			if bsp.debugger.Enabled() {
				bsp.debugger.DebugSpansInvalidUTF8(err, bsp.batch)
			}
		}
	}
}

func (bsp *batchSpanProcessor) exportBatch(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			selflog.Errorf("Recovered from panic: %v\n", r)
		}
	}()
	err := bsp.e.ExportSpans(ctx, bsp.batch)
	if bsp.o.exportToFile {
		protoSpans := tracetransform.Spans(bsp.batch)
		bsp.fileExporter.Export(protoSpans)
	}
	return err
}

// processQueue removes spans from the `queue` channel until processor
// is shut down. It calls the exporter in batches of up to MaxExportBatchSize
// waiting up to BatchTimeout to form a batch.
func (bsp *batchSpanProcessor) processQueue() {
	defer bsp.timer.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		select {
		case <-bsp.stopCh:
			return
		case <-bsp.timer.C:
			metric.GetSelfMonitor().Stats.TracesStats.BatchByTimerCounter.Inc()
			if err := bsp.exportSpans(ctx); err != nil {
				otel.Handle(err)
			}
		case span := <-bsp.queue:
			if ffs, ok := span.(forceFlushSpan); ok {
				close(ffs.flushed)
				continue
			}
			bsp.batchMutex.Lock()
			bsp.batch = append(bsp.batch, span)
			bsp.batchedSize += calcSpanSize(span)
			shouldExport := bsp.shouldProcessInBatch()
			bsp.batchMutex.Unlock()
			if shouldExport {
				if !bsp.timer.Stop() {
					<-bsp.timer.C
				}
				if err := bsp.exportSpans(ctx); err != nil {
					otel.Handle(err)
				}
			}
		}
	}
}

// drainQueue awaits any caller that had added to bsp.stopWait
// to finish the enqueue, then exports the final batch.
func (bsp *batchSpanProcessor) drainQueue() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		select {
		case sd := <-bsp.queue:
			if sd == nil {
				if err := bsp.exportSpans(ctx); err != nil {
					otel.Handle(err)
				}
				return
			}

			bsp.batchMutex.Lock()
			bsp.batch = append(bsp.batch, sd)
			shouldExport := len(bsp.batch) == bsp.o.MaxExportBatchSize
			bsp.batchMutex.Unlock()

			if shouldExport {
				if err := bsp.exportSpans(ctx); err != nil {
					otel.Handle(err)
				}
			}
		default:
			close(bsp.queue)
		}
	}
}

func (bsp *batchSpanProcessor) enqueue(sd sdktrace.ReadOnlySpan) {
	bsp.enqueueBlockOnQueueFull(context.TODO(), sd, bsp.o.BlockOnQueueFull)
}

func (bsp *batchSpanProcessor) enqueueBlockOnQueueFull(
	ctx context.Context, span sdktrace.ReadOnlySpan, block bool,
) bool {
	metric.GetSelfMonitor().Stats.TracesStats.EnqueueCounter.Inc()

	// This ensures the bsp.queue<- below does not panic as the
	// processor shuts down.
	defer func() {
		x := recover()
		switch err := x.(type) {
		case nil:
			return
		case runtime.Error:
			if err.Error() == "send on closed channel" {
				return
			}
		}
		panic(x)
	}()

	select {
	case <-bsp.stopCh:
		return false
	default:
	}

	if block {
		select {
		case bsp.queue <- span:
			return true
		case <-ctx.Done():
			return false
		}
	}

	select {
	case bsp.queue <- span:
		return true
	default:
		bsp.dropped.Inc()
		metric.GetSelfMonitor().Stats.TracesStats.DropCounter.Inc()
	}
	return false
}

// shouldProcessInBatch determines whether to export in batches
func (bsp *batchSpanProcessor) shouldProcessInBatch() bool {
	if len(bsp.batch) == bsp.o.MaxExportBatchSize {
		metric.GetSelfMonitor().Stats.TracesStats.BatchByCountCounter.Inc()
		return true
	}

	if bsp.batchedSize >= bsp.o.MaxPacketSize {
		metric.GetSelfMonitor().Stats.TracesStats.BatchByPacketSizeCounter.Inc()
		return true
	}

	return false
}

// calcSpanSize calculates the packet size of a Span
func calcSpanSize(sd sdktrace.ReadOnlySpan) int {
	if sd == nil {
		return 0
	}

	size := 0
	// just calculate events size for now.
	for _, event := range sd.Events() {
		for _, kv := range event.Attributes {
			size += len(kv.Key)
			size += len(kv.Value.AsString())
		}
	}
	return size
}
