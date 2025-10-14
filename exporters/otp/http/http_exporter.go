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

package http

import (
	"net/http"
	"time"

	httputil "galiosight.ai/galio-sdk-go/lib/http"
	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"go.uber.org/atomic"
)

// HTTPExporter 监控导出器。
type HTTPExporter interface {
	// Export 导出监控指标。buffer 由外面传进来，进行对象重用，减少内存分配。
	Export(message proto.Message, obj *ReuseObject) error
	// UpdateConfig 更新配置，用于运行时配置热更新。
	UpdateConfig(maxRetryCount int32, collector model.Collector)
}

// HTTPGeneralExporter 使用 HTTP post 方式导出数据。
type HTTPGeneralExporter struct {
	Log           *logs.Wrapper
	HTTPClient    *http.Client
	CollectorAddr *CollectorAddr
	MaxRetryCount atomic.Int32
	Headers       map[string]string
}

// UpdateConfig 更新配置，用于运行时配置热更新。
func (h *HTTPGeneralExporter) UpdateConfig(
	maxRetryCount int32, collector model.Collector,
) {
	h.MaxRetryCount.Store(maxRetryCount)
	h.CollectorAddr.UpdateIPPorts(collector.Addr, collector.DirectIpPort)
	h.Log.Debugf(
		"[galileo]MaxRetryCount=%v,cfg=%v,h.collectorAddr=%v",
		maxRetryCount, collector, h.CollectorAddr,
	)
}

// NewHTTPGeneralExporter 构造 HTTP otp metrics 导出器。
func NewHTTPGeneralExporter(
	timeoutMs int, collectorAddr string, log *logs.Wrapper, options ...option,
) *HTTPGeneralExporter {
	opt := opts{}
	for _, o := range options {
		o(&opt)
	}
	return &HTTPGeneralExporter{
		Log: log,
		HTTPClient: &http.Client{
			Timeout:   time.Duration(timeoutMs) * time.Millisecond,
			Transport: &http.Transport{DisableKeepAlives: true},
		},
		CollectorAddr: NewCollectorAddr(collectorAddr, opt.directIPPorts),
		MaxRetryCount: *atomic.NewInt32(opt.maxRetryCount),
		Headers:       opt.headers,
	}
}

type opts struct {
	directIPPorts []string
	headers       map[string]string
	maxRetryCount int32
}

type option func(*opts)

// WithDirectIPPorts 设置直连 ip。
func WithDirectIPPorts(directIPPorts []string) option {
	return func(o *opts) {
		o.directIPPorts = directIPPorts
	}
}

// WithHeaders 设置请求头
func WithHeaders(headers map[string]string) option {
	return func(o *opts) {
		o.headers = headers
	}
}

// WithMaxRetryCount 设置最大重试次数
func WithMaxRetryCount(maxRetryCount int32) option {
	return func(o *opts) {
		o.maxRetryCount = maxRetryCount
	}
}

// ReuseObject 保存可以重用的对象，减少内存分配。
type ReuseObject struct {
	PbBuf     *proto.Buffer
	SnappyBuf []byte
}

// NewReuseObject 将 worker 里面用到的 buffer 进行对象重用，减少内存分配。
func NewReuseObject() *ReuseObject {
	const bufSize = 1024 * 10
	r := &ReuseObject{
		PbBuf:     proto.NewBuffer(make([]byte, bufSize)),
		SnappyBuf: make([]byte, bufSize),
	}
	return r
}

// Reset 将 ReuseObject 对象置空以复用
func (r *ReuseObject) Reset() {
	// 重置 buffer
	r.PbBuf.Reset()
	// 恢复最大容量
	r.SnappyBuf = r.SnappyBuf[:cap(r.SnappyBuf)]
}

// Export 使用 HTTP 方式导出监控指标数据。真正执行数据上报。
// 此函数是并发安全的。
// 用到的 buffer 由外面通过 reuseObject 传进来，进行对象重用，减少内存分配。
// 实现 HTTPExporter 接口，方便进行单测。
// 由于 HTTP 长连接有时候会出现 EOF 错误，所以进行最多 h.maxRetryCount 次重试发送。
func (h *HTTPGeneralExporter) Export(message proto.Message, obj *ReuseObject) error {
	obj.Reset()
	err := obj.PbBuf.Marshal(message)
	if err != nil {
		return err
	}
	obj.SnappyBuf = snappy.Encode(obj.SnappyBuf, obj.PbBuf.Bytes())
	err = h.tryExport(message, obj)
	if err != nil {
		h.Log.Errorf("[galileo]HTTPGeneralExporter.Export|err=%v", err)
		return err
	}
	return nil
}

// tryExport 尝试发送请求，失败会进行重试。
// 重试参数 h.maxRetryCount 可通过配置文件控制。
// 实际执行次数等于 h.maxRetryCount+1。
// 只有连接失败等 net.OpError 错误才重试，HTTP 超时等情况不进行重试，避免服务端收到重复的包，导致数据翻倍。
func (h *HTTPGeneralExporter) tryExport(message proto.Message, obj *ReuseObject) error {
	var err error
	for i := 0; i < int(h.MaxRetryCount.Load())+1; i++ {
		addr := h.CollectorAddr.GetAddr(i)
		var rsp []byte
		rsp, err = httputil.Post(h.HTTPClient, addr, obj.SnappyBuf, httputil.WithHeaders(h.Headers))
		h.Log.Debugf(
			"[galileo]HTTPGeneralExporter.Export|addr=%v,message=%+v,rsp=%v,err=%v",
			addr, message.String(), string(rsp), err,
		)
		if err == nil {
			break
		}
		if !httputil.IsNetOpError(err) {
			break
		}
	}
	return err
}
