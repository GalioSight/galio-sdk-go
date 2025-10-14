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

// Package traces ...
// 详见 https://doc.weixin.qq.com/doc/w3_AI4AQAbdAFwJ052ylM4TIyhAj9qoS
// 「Workflow 二期 SDK 100% 完图率」
// ShouldSample:
//
//  1. 使用 childPath 计算下游的 hash
//     1.1. takeKey 获取计算用的 attribute 字段
//     1.2. hashPath 计算 hash 值
//  2. 检查是否在 cache 中
//  3. 返回 attribute
package traces

import (
	"encoding/hex"
	"sync/atomic"
	"time"

	"github.com/cespare/xxhash/v2"
	lru "github.com/qianbin/directcache"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	attrutil "galiosight.ai/galio-sdk-go/exporters/otlp/traces/attribute"
	"galiosight.ai/galio-sdk-go/exporters/otlp/traces/tracestate"
	"galiosight.ai/galio-sdk-go/lib/strings"
	"galiosight.ai/galio-sdk-go/model"
	slog "galiosight.ai/galio-sdk-go/self/log"
	"galiosight.ai/galio-sdk-go/semconv"
)

// WorkflowPathSampler path 采样，对没见过的路径 hash 必命中一次，是 min count 采样
// 的升级版，能实现接近 100% 的调用图覆盖，方案文档
// https://doc.weixin.qq.com/doc/w3_AI4AQAbdAFwJ052ylM4TIyhAj9qoS
type WorkflowPathSampler struct {
	enable bool
	// 根据 https://github.com/vmihailenco/go-clientSpanCache-benchmark 选型
	// s4lru 虽然快但易用性不好
	// hashicorp 的 lru, 并发性不好，并且接口导致额外的 convTstring 的调用
	// directcache 和 fastcache 性能相近 (代码也差不多)，但前者的依赖更简单一些
	clientSpanCache *lru.Cache
	serverSpanCache *lru.Cache // 为了上报 serverSpan，不得不新增一个 cache
	pathMaxCount    int
	lifetime        time.Duration
	maximumCount    int32 // max count per minute
	currentCount    int32 // current count
}

// WorkflowPathBuffer 外部 buffer，优化内存
type WorkflowPathBuffer struct {
	encodedPath [hexHeadLen + hexHashLen]byte
	attr        [2]attribute.KeyValue
}

// NewWorkflowPathSampler 创建 WorkflowPathSampler 并且启动更新线程
func NewWorkflowPathSampler(c *model.WorkflowSamplerConfig) *WorkflowPathSampler {
	w := &WorkflowPathSampler{lifetime: time.Hour}
	w.UpdateConfig(c)
	w.clientSpanCache = lru.New(w.pathMaxCount * len(pathbin{}))
	w.serverSpanCache = lru.New(w.pathMaxCount * len(pathbin{}))
	l := newLoop(w.loop, w.lifetime)
	go l.run()
	return w
}

// UpdateConfig 重载配置
func (w *WorkflowPathSampler) UpdateConfig(c *model.WorkflowSamplerConfig) {
	if c == nil {
		w.enable = false
		return
	}
	w.enable = c.PathMaxCount > 0
	if !w.enable {
		return
	}
	w.pathMaxCount = int(c.PathMaxCount)
	lifetime := time.Second * time.Duration(c.LifetimeSec)
	if lifetime == 0 {
		lifetime = time.Hour // 默认 1 小时
	}
	w.lifetime = lifetime
	w.maximumCount = c.MaxCountPerMinute
}

// ShouldSample 是否应该采样
func (w *WorkflowPathSampler) ShouldSample(
	p *sdktrace.SamplingParameters,
	state *tracestate.WorkflowState,
	ref *WorkflowPathBuffer,
) (ret sdktrace.SamplingResult) {

	if !w.enable {
		return drop
	}
	switch p.Kind {
	case trace.SpanKindServer:
		return w.serverShouldSample(p, state, ref)
	case trace.SpanKindClient:
		return w.clientShouldSample(p, state, ref)
	default:
		return drop
	}
}

var separator = []byte("\xff")

func hashPath(path, to []byte, server, method string) []byte {
	h := xxhash.New()
	if path != nil {
		h.Write(path)
		h.Write(separator)
	}
	h.Write(strings.NoAllocBytes(server))
	h.Write(separator)
	h.Write(strings.NoAllocBytes(method))
	res := h.Sum(to[:0])
	return res
}

func decodeToBuf(dst, src []byte) []byte {
	n, err := hex.Decode(dst, src)
	if err != nil {
		slog.Errorf("%v", err)
	}
	return dst[:n]
}

// 不能使用 var encodeToString = hex.EncodeToString，无法内联
func encodeToString(src []byte) string {
	return hex.EncodeToString(src)
}

func encodeToBuf(dst, src []byte) int {
	return hex.Encode(dst, src)
}

func encodeLen(n int) int {
	return hex.EncodedLen(n)
}

func setAttribute(ret *sdktrace.SamplingResult, workflowPathBuffer *WorkflowPathBuffer, child, path string) {
	ret.Attributes = workflowPathBuffer.attr[:]
	ret.Attributes[1] = semconv.WorkflowChildPathKey.String(child)
	ret.Attributes[0] = semconv.WorkflowPathKey.String(path)
}

func (w *WorkflowPathSampler) serverShouldSample(
	p *sdktrace.SamplingParameters,
	state *tracestate.WorkflowState,
	ref *WorkflowPathBuffer,
) (ret sdktrace.SamplingResult) {

	var child []byte
	var childhex string
	var mid temporary
	if state.Path != "" {
		child = decodeToBuf(mid.child[:], strings.NoAllocBytes(state.Path))
		childhex = state.Path
	} else {
		keys := attrutil.NewRPCKeys(p.Attributes)
		child = rootPath(mid.child[:], keys.CalleeService, keys.CalleeMethod)
		childhex = encodeToString(child)
		state.Path = childhex
	}

	if has := w.serverSpanCache.Has(child); has {
		return drop
	}
	if w.currentCount > w.maximumCount {
		return drop
	}
	atomic.AddInt32(&w.currentCount, 1)

	// 前置采样阶段只设置成 readonly，避免将 sampled flag 传递给下游，在后置采样阶段对所有命中的 workflow 都更改为采样
	ret = recordOnly
	// server 只需要直接上报即可
	setAttribute(&ret, ref, childhex, state.ParentPath)
	w.serverSpanCache.Set(child, nil)
	return ret
}

func xor(dst, high, low []byte) {
	dst[0] = high[0] ^ low[0]
	dst[1] = high[1] ^ low[1]
	dst[2] = high[2] ^ low[2]
	dst[3] = high[3] ^ low[3]
}

const (
	headLen    = 4
	hashLen    = 8
	padLen     = 4
	hexHeadLen = headLen * 2
	hexHashLen = hashLen * 2
)

// Path binary format |head:32bit|hash:64bit|pad:32bit| .
// 其中 同一个 root 的所有下游的 head 相同，这样可以快速提取出来所有的 path，
// hash 是历史路径 hash 结果，即 child = hash(path,server,method)
// pad 是 struct padding, 对齐内存，同时兼容未来余量
// 这里并不能把 pathbin 定义成 struct {head int32, hash int64}，因为 cache.Has([]byte) 这里需要使用 []byte
type pathbin [headLen + hashLen + padLen]byte

// 中间临时结构
type temporary struct {
	path  pathbin
	child pathbin
}

func rootPath(dst []byte, server, method string) []byte {
	hash := hashPath(nil, dst[headLen:], server, method)
	const half = hashLen / 2
	xor(dst[:headLen], hash[:half], hash[half:hashLen]) // head = hash[low] xor hash[high]
	return dst[:headLen+hashLen]
}

// childPath = hash(pathstr, t(a))
func childPath(
	ref *WorkflowPathBuffer,
	tempStruct *temporary, pathstr string,
	a []attribute.KeyValue,
) (path, child []byte, childhex string) {
	keys := attrutil.NewRPCKeys(a)
	if pathstr != "" {
		path = decodeToBuf(tempStruct.path[:], strings.NoAllocBytes(pathstr))
	} else {
		// workflow 根节点
		path = rootPath(tempStruct.path[:], keys.CallerService, keys.CallerMethod)
	}

	copy(tempStruct.child[:headLen], tempStruct.path[:headLen]) // 前面 4B 共享前缀
	hashPath(path[headLen:], tempStruct.child[headLen:], keys.CalleeService, keys.CalleeMethod)
	child = tempStruct.child[:headLen+hashLen]
	encodeToBuf(ref.encodedPath[:], child)
	return path, child, strings.NoAllocString(ref.encodedPath[:hexHeadLen+hexHashLen])
}

func (w *WorkflowPathSampler) clientShouldSample(
	p *sdktrace.SamplingParameters,
	state *tracestate.WorkflowState,
	ref *WorkflowPathBuffer,
) (samplingResult sdktrace.SamplingResult) {

	var mid temporary // 兼容未来 128bit, 放 childPath 会逃逸
	old := state.Path
	path, child, childhex := childPath(ref, &mid, state.Path, p.Attributes)
	state.Path = childhex
	if old == "" {
		// 延迟计算，大部分不上报，大部分都不需要额外的 encode
		// 这里会返回到非常外面，不能再用栈内存了
		old = encodeToString(path)
	}
	state.ParentPath = old

	if has := w.clientSpanCache.Has(child); has {
		return drop
	}
	if w.currentCount > w.maximumCount {
		return drop
	}
	atomic.AddInt32(&w.currentCount, 1)

	// 前置采样阶段只设置成 readonly，避免将 sampled flag 传递给下游，在后置采样阶段对所有命中的 workflow 都更改为采样
	samplingResult = recordOnly
	setAttribute(&samplingResult, ref, state.Path, old)
	w.clientSpanCache.Set(child, nil)
	return samplingResult
}

type loop struct {
	loopFunc       func(*loop)
	lifetimeTicker *time.Ticker
	resetTicker    *time.Ticker
	tc             <-chan time.Time // 允许单测覆盖

	closed   chan bool
	lifetime time.Duration
}

func newLoop(loopFunc func(*loop), lifetime time.Duration) *loop {
	return &loop{loopFunc: loopFunc, lifetime: lifetime, closed: make(chan bool)}
}

func (l *loop) run() {
	l.lifetimeTicker = time.NewTicker(l.lifetime)
	l.tc = l.lifetimeTicker.C
	l.resetTicker = time.NewTicker(time.Minute)
	defer l.lifetimeTicker.Stop()
	defer l.resetTicker.Stop()
	for l.closed != nil {
		l.loopFunc(l)
	}
}

func (w *WorkflowPathSampler) loop(l *loop) {
	select {
	case <-l.tc:
		if !w.enable {
			return
		}
		w.clientSpanCache.Reset(w.pathMaxCount * len(pathbin{}))
		w.serverSpanCache.Reset(w.pathMaxCount * len(pathbin{}))
		if l.lifetime != w.lifetime {
			l.lifetime = w.lifetime
			l.lifetimeTicker.Reset(w.lifetime)
		}
	case <-l.resetTicker.C:
		atomic.StoreInt32(&w.currentCount, 0)
	case <-l.closed:
		l.closed = nil
		return
	}
}
