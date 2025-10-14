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

// Package logs 日志导出器
package logs

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	collectorlogpb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	logpb "go.opentelemetry.io/proto/otlp/logs/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"galiosight.ai/galio-sdk-go/components"
	"galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/errs"
	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"

	// 启用 gzip 压缩，需要 import gzip 包
	_ "google.golang.org/grpc/encoding/gzip"
)

// exporter 日志导出器实现。
type exporter struct {
	grpcLogServiceClient       collectorlogpb.LogsServiceClient
	httpLogClient              *logClient
	grpcMetaDatas              metadata.MD
	grpcClientConn             *grpc.ClientConn
	disconnectedCh             chan bool
	backgroundConnectionDoneCh chan bool
	lastConnectErr             unsafe.Pointer
	stopCh                     chan bool
	grpcOptions                grpcOptions
	mu                         sync.RWMutex
	startOnce                  sync.Once
	exportMu                   sync.Mutex
	started                    bool
	log                        *logs.Wrapper
	stats                      *model.SelfMonitorStats
}

// NewExporter 构造 otlp logs 导出器。
func NewExporter(baseCfg *configs.Logs) (components.LogsExporter, error) {
	baseCfg.Log.Infof("[galileo]new otlp logs exporter|baseCfg=%+v", baseCfg)
	e, err := newExporter(
		withInsecure(),
		withAddress(baseCfg.Exporter.Collector.Addr),
		withCompressor("gzip"),
		withGRPCHeaders(
			map[string]string{
				model.TenantHeaderKey: baseCfg.Resource.TenantId,
				model.TargetHeaderKey: baseCfg.Resource.Target,
				model.APIKeyHeaderKey: baseCfg.APIKey,
			},
		),
		withHTTPEnabled(true), // 默认使用 http 协议，因为 ias 域名不支持 grpc
	)
	e.log = baseCfg.Log
	e.stats = baseCfg.Stats
	if err != nil {
		e.log.Errorf(
			"[galileo]new otlp logs exporter|err=%v, baseCfg=%+v", err, baseCfg,
		)
		e.stats.LogsStats.InitErrorTotal.Inc()
	}
	return e, err
}

// newExporter constructs a new Exporter and starts it.
func newExporter(opts ...grpcOption) (*exporter, error) {
	e := newUnstartedExporter(opts...)
	if err := e.start(); err != nil {
		return nil, err
	}
	return e, nil
}

// newUnstartedExporter constructs a new Exporter and does not start it.
func newUnstartedExporter(opts ...grpcOption) *exporter {
	e := &exporter{}
	e.grpcOptions = newGRPCOptions(opts...)
	if len(e.grpcOptions.headers) > 0 {
		e.grpcMetaDatas = metadata.New(e.grpcOptions.headers)
	}
	e.httpLogClient = newLogClient(e.grpcOptions.addr, e.grpcOptions.headers)
	return e
}

// start 导出器启动。
func (e *exporter) start() error {
	var err = errs.ErrOTLPLogsExporterAlreadyStarted
	e.startOnce.Do(
		func() {
			e.mu.Lock()
			e.started = true
			e.disconnectedCh = make(chan bool, 1)
			e.stopCh = make(chan bool)
			e.backgroundConnectionDoneCh = make(chan bool)
			e.mu.Unlock()
			// An optimistic first connection attempt to ensure that
			// applications under heavy load can immediately process
			// data. See https://github.com/census-ecosystem/opencensus-go-exporter-ocagent/pull/63
			if err = e.connect(); err == nil { // 启动连接一次。
				e.setStateConnected() // 设置连接成功状态。
			} else {
				e.setStateDisconnected(err) // 设置连接失败状态。
			}
			go e.indefiniteBackgroundConnection() // 后台保活重连。
			err = nil
		},
	)
	return err
}

// connect grpc 连接。
func (e *exporter) connect() error {
	cc, err := e.dialToCollector()
	if err != nil {
		return err
	}
	return e.createLogServiceConnection(cc)
}

// dialToCollector grpc 连接 collector。
func (e *exporter) dialToCollector() (*grpc.ClientConn, error) {
	addr := e.prepareCollectorAddress()
	dialOpts := []grpc.DialOption{}
	if e.grpcOptions.serviceConfig != "" {
		dialOpts = append(
			dialOpts,
			grpc.WithDefaultServiceConfig(e.grpcOptions.serviceConfig),
		)
	}
	if e.grpcOptions.clientCredentials != nil {
		dialOpts = append(
			dialOpts,
			grpc.WithTransportCredentials(e.grpcOptions.clientCredentials),
		)
	} else if e.grpcOptions.dialInsecure {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}
	if e.grpcOptions.compressor != "" {
		dialOpts = append(
			dialOpts,
			grpc.WithDefaultCallOptions(grpc.UseCompressor(e.grpcOptions.compressor)),
		)
	}
	if len(e.grpcOptions.dialOptions) != 0 {
		dialOpts = append(dialOpts, e.grpcOptions.dialOptions...)
	}
	ctx := e.contextWithMetadata(context.Background())
	return grpc.DialContext(ctx, addr, dialOpts...)
}

func (e *exporter) contextWithMetadata(ctx context.Context) context.Context {
	if e.grpcMetaDatas.Len() > 0 {
		return metadata.NewOutgoingContext(ctx, e.grpcMetaDatas)
	}
	return ctx
}

func (e *exporter) createLogServiceConnection(cc *grpc.ClientConn) error {
	e.mu.RLock()
	started := e.started
	e.mu.RUnlock()
	if !started {
		return errs.ErrOTLPLogsExporterNotStarted
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	// If previous clientConn is same as the current then just return.
	// This doesn't happen right now as this func is only called with new ClientConn.
	// It is more about future-proofing.
	if e.grpcClientConn == cc {
		return nil
	}
	// If the previous clientConn was non-nil, close it
	if e.grpcClientConn != nil {
		_ = e.grpcClientConn.Close()
	}
	e.grpcClientConn = cc
	e.grpcLogServiceClient = collectorlogpb.NewLogsServiceClient(cc)
	return nil
}

const (
	// defaultCollectorPort is the port the Exporter will attempt to connect to
	// if no collector port is provided.
	defaultCollectorPort uint16 = 55680
	// defaultCollectorHost is the host address the Exporter will attempt
	// connect to if no collector address is provided.
	defaultCollectorHost string = "localhost"
)

func (e *exporter) prepareCollectorAddress() string {
	if e.grpcOptions.addr != "" {
		return e.grpcOptions.addr
	}
	return fmt.Sprintf("%s:%d", defaultCollectorHost, defaultCollectorPort)
}

// ExportLogs 导出日志。
func (e *exporter) ExportLogs(
	parent context.Context, logs []*logpb.ResourceLogs,
) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	go func(ctx context.Context, cancel context.CancelFunc) {
		select {
		case <-ctx.Done():
		case <-e.stopCh:
			cancel()
		}
	}(ctx, cancel)

	if len(logs) == 0 {
		return nil
	}

	if !e.connected() {
		return errs.ErrOTLPLogsExporterDisconnected
	}

	select {
	case <-e.stopCh:
		return errs.ErrOTLPLogsExporterStopped
	case <-ctx.Done():
		return errs.ErrOTLPLogsExporterContextCanceled
	default:
		e.exportMu.Lock()
		resp, err := e.exportLogs(logs, ctx)
		e.exportMu.Unlock()
		e.log.Debugf("ExportLogs err=%v,len(logs)=%v,rsp=%v\n", err, len(logs), resp)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *exporter) exportLogs(logs []*logpb.ResourceLogs, ctx context.Context) (
	*collectorlogpb.ExportLogsServiceResponse, error,
) {
	request := &collectorlogpb.ExportLogsServiceRequest{
		ResourceLogs: logs,
	}
	if e.grpcOptions.httpEnabled {
		return e.httpLogClient.Export(ctx, request)
	}
	response, err := e.grpcLogServiceClient.Export(e.contextWithMetadata(ctx), request)
	if err != nil {
		e.setStateDisconnected(err)
	}
	return response, err
}

func (e *exporter) connected() bool {
	return e.lastConnectError() == nil
}

func (e *exporter) lastConnectError() error {
	errPtr := (*error)(atomic.LoadPointer(&e.lastConnectErr))
	if errPtr == nil {
		return nil
	}
	return *errPtr
}

// Shutdown 关闭导出器。
func (e *exporter) Shutdown(ctx context.Context) error {
	e.mu.RLock()
	cc := e.grpcClientConn
	started := e.started
	e.mu.RUnlock()

	if !started {
		return nil
	}

	var err error
	if cc != nil {
		// Clean things up before checking this error.
		err = cc.Close()
	}

	// At this point we can change the state variable started
	e.mu.Lock()
	e.started = false
	e.mu.Unlock()
	closeStopCh(e.stopCh)
	// Ensure that the backgroundConnector returns
	select {
	case <-e.backgroundConnectionDoneCh:
	case <-ctx.Done():
		return ctx.Err()
	}
	return err
}

// closeStopCh is used to wrap the exporters stopCh channel closing for testing.
var closeStopCh = func(stopCh chan bool) {
	close(stopCh)
}

func (e *exporter) setStateConnected() {
	e.saveLastConnectError(nil)
}

func (e *exporter) setStateDisconnected(err error) {
	e.saveLastConnectError(err)
	select {
	case e.disconnectedCh <- true:
	default:
	}
}

func (e *exporter) saveLastConnectError(err error) {
	var errPtr *error
	if err != nil {
		errPtr = &err
	}
	atomic.StorePointer(&e.lastConnectErr, unsafe.Pointer(errPtr))
}

const defaultConnReattemptPeriod = 10 * time.Second

func (e *exporter) indefiniteBackgroundConnection() {
	defer func() {
		e.backgroundConnectionDoneCh <- true
	}()
	connReattemptPeriod := e.grpcOptions.reconnectionPeriod
	if connReattemptPeriod <= 0 {
		connReattemptPeriod = defaultConnReattemptPeriod
	}
	// No strong seeding required, nano time can
	// already help with pseudo uniqueness.
	rng := rand.New(rand.NewSource(time.Now().UnixNano() + rand.Int63n(1024)))
	// maxJitterNanos: 70% of the connectionReattemptPeriod
	maxJitterNanos := int64(0.7 * float64(connReattemptPeriod))
	for {
		// Otherwise these will be the normal scenarios to enable
		// reconnection if we trip out.
		// 1. If we've stopped, return entirely
		// 2. Otherwise, block until we are disconnected, and
		//    then retry connecting
		select {
		case <-e.stopCh:
			return
		case <-e.disconnectedCh:
			// Normal scenario that we'll wait for
		}
		if err := e.connect(); err == nil {
			e.setStateConnected()
		} else {
			e.setStateDisconnected(err)
		}
		// Apply some jitter to avoid lockstep retrials of other
		// collector-exporters. Lockstep retrials could result in an
		// innocent DDOS, by clogging the machine's resources and network.
		jitter := time.Duration(rng.Int63n(maxJitterNanos))
		select {
		case <-e.stopCh:
			return
		case <-time.After(connReattemptPeriod + jitter):
		}
	}
}
