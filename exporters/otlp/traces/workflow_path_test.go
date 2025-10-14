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
package traces

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	lru "github.com/qianbin/directcache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	apitrace "go.opentelemetry.io/otel/trace"

	"galiosight.ai/galio-sdk-go/exporters/otlp/traces/tracestate"
	"galiosight.ai/galio-sdk-go/model"
	omp3 "galiosight.ai/galio-sdk-go/semconv"
	semconv "galiosight.ai/galio-sdk-go/semconv/v1.0.0"
)

type TestSuite struct {
	suite.Suite
}

func (s *TestSuite) SetupSuite() {
}

func (s *TestSuite) TearDownSuite() {
}

func (s *TestSuite) SetupTest() {
}

func (s *TestSuite) TearDownTest() {
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func mkcfg(c, l, c2 int32) *model.WorkflowSamplerConfig {
	return &model.WorkflowSamplerConfig{
		PathMaxCount:      c,
		LifetimeSec:       l,
		MaxCountPerMinute: c2,
	}
}

func manualWorkflowPath(c *model.WorkflowSamplerConfig) *WorkflowPathSampler {
	w := &WorkflowPathSampler{
		clientSpanCache: lru.New(0),
		serverSpanCache: lru.New(0),
	}
	w.UpdateConfig(c)
	return w
}

func (s *TestSuite) TestUpdateConfig() {
	w := manualWorkflowPath(mkcfg(4, 1, 120))
	s.True(w.enable)
	s.NotNil(w.clientSpanCache)
	old := w.clientSpanCache

	l := newLoop(w.loop, w.lifetime)
	l.lifetimeTicker = time.NewTicker(w.lifetime)
	l.resetTicker = time.NewTicker(time.Minute)
	tc := make(chan time.Time, 1)
	l.tc = tc
	round := func() {
		tc <- time.Now()
		w.loop(l)
	}

	w.UpdateConfig(mkcfg(4, 1, 120))
	round()
	s.Equal(w.clientSpanCache, old) // 没有发生重构

	// 更改大小，时间都会在下次窗口发生重构
	w.UpdateConfig(mkcfg(1024*16, 1, 120))
	round()
	s.Equal(1024*16, w.pathMaxCount)
	s.Equal(1024*16, w.clientSpanCache.Capacity()/len(pathbin{}))

	old = w.clientSpanCache
	w.UpdateConfig(mkcfg(8, 2, 120))
	round()
	s.Equal(l.lifetime, w.lifetime)

	w.UpdateConfig(mkcfg(0, 2, 120))
	s.False(w.enable)
}

func genP(kind trace.SpanKind, kv ...string) sdktrace.SamplingParameters {
	sp := sdktrace.SamplingParameters{
		Kind: kind,
	}
	sp.Attributes = append(
		sp.Attributes,
		semconv.TrpcCallerServerKey.String(kv[0]),
		semconv.TrpcCallerMethodKey.String(kv[1]),
		semconv.TrpcCalleeServerKey.String(kv[2]),
		semconv.TrpcCalleeMethodKey.String(kv[3]),
	)
	return sp

}

func genS(path string) tracestate.WorkflowState {
	ts := tracestate.WorkflowState{
		Path:       path,
		ParentPath: path,
	}
	return ts
}

func genS2(path, parent string) tracestate.WorkflowState {
	ts := tracestate.WorkflowState{
		Path:       path,
		ParentPath: parent,
	}
	return ts
}

var (
	ddrop = sdktrace.Drop       // decision drop
	dkeep = sdktrace.RecordOnly // decision keep
	kc    = trace.SpanKindClient
	ks    = trace.SpanKindServer
)

type pathKeys struct {
	Path  string
	Child string
}

func newPathKeys(kv []attribute.KeyValue) pathKeys {
	var k pathKeys
	for i := range kv {
		switch kv[i].Key {
		case omp3.WorkflowPathKey:
			k.Path = kv[i].Value.AsString()
		case omp3.WorkflowChildPathKey:
			k.Child = kv[i].Value.AsString()
		}
	}
	return k
}

// mkHash 随便生成一个 hash 即可
func mkHash(server, method string) string {
	var buf [len(pathbin{})]byte
	to := rootPath(buf[:], server, method)
	return encodeToString(to)
}

// TestServerTracestate 测试 server span 的 tracestate 符合预期
func (s *TestSuite) TestServerTracestate() {
	ih := mkHash("1", "2")
	ih2 := mkHash("3", "4")
	ih3 := mkHash("5", "6")
	cases := []struct {
		sp sdktrace.SamplingParameters
		ts tracestate.WorkflowState
	}{
		{genP(ks, "A", "a", "B", "b"), genS2("", "")},
		{genP(ks, "A", "a", "B", "b"), genS2(ih, "")},
		{genP(ks, "A", "a", "B", "b"), genS2(ih2, ih3)},
	}
	w := NewWorkflowPathSampler(mkcfg(8, 2, 120))
	var buf WorkflowPathBuffer
	for _, cc := range cases {
		s.Run(
			"", func() {
				res := w.ShouldSample(&cc.sp, &cc.ts, &buf)
				key := newPathKeys(res.Attributes)
				s.Equal(key.Child, cc.ts.Path)
				s.Equal(key.Path, cc.ts.ParentPath)
			},
		)
	}
}

func (s *TestSuite) TestShouldSample() {
	ih := mkHash("1", "2")
	cases := []struct {
		sp       sdktrace.SamplingParameters
		ts       tracestate.WorkflowState
		decision sdktrace.SamplingDecision
		state    string
	}{
		{genP(kc, "A", "a", "B", "b"), genS(""), dkeep, "w:4:081c8ff8de2b1285f988e9b5:081c8ff8a9b9ec16a1a563ee"},
		{genP(kc, "A", "a", "B", "b"), genS(""), ddrop, "w:1:081c8ff8de2b1285f988e9b5:081c8ff8a9b9ec16a1a563ee"}, // 第二次遇见不上报
		{genP(kc, "A", "a", "B", "b"), genS(ih), dkeep, "w:4:a3942ff75a0125ee8718866f:a3942ff79d48c1e53edcee12"}, // 边一样，起点不同，需要上报
		{genP(kc, "A", "a", "B", "b"), genS(ih), ddrop, "w:1:a3942ff75a0125ee8718866f:a3942ff79d48c1e53edcee12"},
		{genP(kc, "A", "a", "C", "c"), genS(ih), dkeep, "w:4:a3942ff77654a4d7615058f5:a3942ff79d48c1e53edcee12"}, // 目的不一样，需要上报
		{genP(kc, "A", "a", "C", "c"), genS(ih), ddrop, "w:1:a3942ff77654a4d7615058f5:a3942ff79d48c1e53edcee12"}, // 目的不一样，需要上报
		{genP(ks, "A", "a", "B", "b"), genS(""), dkeep, "w:4:b57ba42161656472d41ec053"},
		{genP(ks, "A", "a", "B", "b"), genS(""), ddrop, "w:1:b57ba42161656472d41ec053"}, // 第二次遇见
		{genP(ks, "A", "a", "B", "b"), genS(ih), dkeep, "w:4:a3942ff79d48c1e53edcee12:a3942ff79d48c1e53edcee12"},
		{genP(ks, "A", "a", "B", "b"), genS(ih), ddrop, "w:1:a3942ff79d48c1e53edcee12:a3942ff79d48c1e53edcee12"}, // 第二次遇见
		{genP(ks, "E", "e", "B", "b"), genS(""), ddrop, "w:1:b57ba42161656472d41ec053"},                          // 根节点，serverSpan 只看被调，和主调无关
		{genP(ks, "E", "e", "F", "f"), genS(ih), ddrop, "w:1:a3942ff79d48c1e53edcee12:a3942ff79d48c1e53edcee12"}, // 子节点，serverSpan 只看 tracestate，和主调被调均无关
	}
	w := NewWorkflowPathSampler(mkcfg(8, 2, 120))
	// w.UpdateConfig()
	var buf WorkflowPathBuffer
	var x [16]byte
	n := len(rootPath(x[:], "A", "B"))
	ts, _ := tracestate.Parse("")

	for _, cc := range cases {
		s.Run(
			"", func() {
				res := w.ShouldSample(&cc.sp, &cc.ts, &buf)
				s.Equal(res.Decision, cc.decision)
				if res.Decision == dkeep {
					key := newPathKeys(res.Attributes)
					var buf [16]byte
					if cc.sp.Kind == kc || (cc.sp.Kind == ks && cc.ts.ParentPath != "") {
						x := decodeToBuf(buf[:], []byte(key.Path))
						s.Len(x, n) // hash 内容没必要固定，长度是符合预期即可
					} else if cc.sp.Kind == ks && cc.ts.ParentPath == "" {
						s.Equal("", key.Path)
					}
					x := decodeToBuf(buf[:], []byte(key.Child))
					s.Len(x, n)
					cc.ts.Result = tracestate.WorkflowPath
				} else {
					cc.ts.Result = tracestate.WorkflowDrop
				}
				ts.Workflow = cc.ts
				s.Equal(cc.state, ts.String())
			},
		)
	}
	w.UpdateConfig(mkcfg(0, 2, 120))
}

func (s *TestSuite) TestSamePrefix() {
	w := manualWorkflowPath(mkcfg(1000, 60, 120))
	sp := genP(kc, "A", "A", "B", "B")
	ts := genS("")
	var buf WorkflowPathBuffer
	res1 := w.ShouldSample(&sp, &ts, &buf)
	sp = genP(kc, "B", "B", "C", "C")
	var buf2 WorkflowPathBuffer
	res2 := w.ShouldSample(&sp, &ts, &buf2)
	s.Equal(res1.Attributes[0].Value.AsString()[:hexHeadLen], res2.Attributes[0].Value.AsString()[:hexHeadLen])
}

func free(w *WorkflowPathSampler) {
	w.UpdateConfig(mkcfg(0, 0, 0))
}

func BenchmarkServerPath(b *testing.B) {
	// 全丢场景
	b.Run(
		"Match", func(b *testing.B) {
			w := manualWorkflowPath(mkcfg(10000, 60, 120))
			sp := genP(ks, "A", "A", "B", "B")
			var buf WorkflowPathBuffer
			h := mkHash("C", "D")
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var tp string
				if i%2 == 0 {
					tp = ""
				} else {
					tp = h
				}
				ts := genS(tp)
				w.ShouldSample(&sp, &ts, &buf) // 大部分情况应该命中缓存
			}
			free(w)
		},
	)
	b.Run(
		"Sequence", func(b *testing.B) {
			w := manualWorkflowPath(mkcfg(int32(b.N/8), 120, int32(b.N)))
			var buf WorkflowPathBuffer
			in := prepare(
				b.N, ks, func(i int) (string, string) {
					return fmt.Sprint(i / 4), fmt.Sprint(i/4 + 1)
				},
			)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				w.ShouldSample(&in[i].sp, &in[i].ts, &buf)
			}
			free(w)
		},
	)
}

type input struct {
	sp sdktrace.SamplingParameters
	ts tracestate.WorkflowState
}

func prepare(n int, sk trace.SpanKind, fn func(i int) (string, string)) []input {
	in := make([]input, n+1)
	for i := 0; i < n+1; i++ {
		caller, callee := fn(i)
		in[i].sp = genP(sk, caller, caller, callee, callee)
		if i%2 == 0 {
			in[i].ts = genS("")
		} else {
			in[i].ts = genS(mkHash(caller, callee))
		}
	}
	return in
}

func TestHash(t *testing.T) {
	var ref WorkflowPathBuffer
	var tmp, expect temporary
	in := []attribute.KeyValue{
		semconv.TrpcCallerServiceKey.String("A"),
		semconv.TrpcCalleeMethodKey.String("C"),
		semconv.TrpcCalleeServiceKey.String("D"),
		semconv.TrpcCallerMethodKey.String("E"),
	}
	pathHex := "e51b254b22d98e1dc7c2ab56"
	_, child, _ := childPath(&ref, &tmp, pathHex, in)

	expPath := decodeToBuf(expect.path[:], []byte(pathHex))
	expChild := hashPath(expPath[headLen:], expect.child[:], "D", "C")
	assert.Equal(t, expChild, child[headLen:])
}

// go test -v -bench=WorkflowPath -run='^$' -benchtime=100000x -benchmem -cpuprofile cpu.out
// go tool pprof -http=0.0.0.0:8000 cpu.out
func BenchmarkWorkflowPath(b *testing.B) {
	b.Run(
		"Match", func(b *testing.B) {
			w := manualWorkflowPath(mkcfg(10000, 60, 120))
			sp := genP(kc, "A", "A", "B", "B")
			var buf WorkflowPathBuffer
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ts := genS("")
				w.ShouldSample(&sp, &ts, &buf) // 大部分情况应该命中缓存
			}
			free(w)
		},
	)
	b.Run(
		"Match2", func(b *testing.B) {
			w := manualWorkflowPath(mkcfg(10000, 60, 120))
			sp := genP(kc, "A", "A", "B", "B")
			var buf WorkflowPathBuffer
			h := mkHash("C", "D")
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ts := genS(h)
				w.ShouldSample(&sp, &ts, &buf) // 大部分情况应该命中缓存
			}
			free(w)
		},
	)

	b.Run(
		"Sequence", func(b *testing.B) {
			w := manualWorkflowPath(mkcfg(10000, 120, int32(b.N)))
			var buf WorkflowPathBuffer
			in := prepare(
				b.N, kc, func(i int) (string, string) {
					return fmt.Sprint(i), fmt.Sprint(i + 1)
				},
			)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				w.ShouldSample(&in[i].sp, &in[i].ts, &buf)
			}
			free(w)
		},
	)
	b.Run(
		"Sequence2", func(b *testing.B) {
			w := manualWorkflowPath(mkcfg(int32(b.N/8), 120, int32(b.N)))
			var buf WorkflowPathBuffer
			in := prepare(
				b.N, kc, func(i int) (string, string) {
					return fmt.Sprint(i / 4), fmt.Sprint(i/4 + 1)
				},
			)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				w.ShouldSample(&in[i].sp, &in[i].ts, &buf)
			}
			free(w)
		},
	)
	b.Run(
		"Parallel", func(b *testing.B) {
			w := manualWorkflowPath(mkcfg(int32(b.N/8), 120, int32(b.N)))
			in := prepare(
				b.N, kc, func(i int) (string, string) {
					return fmt.Sprint(i / 4), fmt.Sprint(i/4 + 1)
				},
			)
			b.ResetTimer()
			var i int32
			b.RunParallel(
				func(pb *testing.PB) {
					for pb.Next() {
						j := atomic.AddInt32(&i, 1)
						var buf WorkflowPathBuffer
						w.ShouldSample(&in[j].sp, &in[j].ts, &buf)
					}
				},
			)
			free(w)
		},
	)
}

type Gen struct {
}

func (g *Gen) param(
	state *apitrace.TraceState,
	flag apitrace.TraceFlags,
) sdktrace.SamplingParameters {
	traceID, _ := apitrace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	sccfg := apitrace.SpanContextConfig{
		TraceFlags: flag,
	}
	if state != nil {
		sccfg.TraceState = *state
	}
	parentCtx := apitrace.ContextWithSpanContext(
		context.Background(),
		apitrace.NewSpanContext(sccfg),
	)
	p := sdktrace.SamplingParameters{
		ParentContext: parentCtx,
		TraceID:       traceID,
		Name:          "parent_is_workflow_span",
	}
	return p
}

func (g *Gen) tracestate(ts string) *apitrace.TraceState {
	var state apitrace.TraceState
	if ts != "" {
		state, _ = state.Insert(galileoVendor, ts)
	}
	return &state
}

var gen Gen
