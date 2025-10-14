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
	"fmt"
	"log"
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"galiosight.ai/galio-sdk-go/configs/ocp"
	"galiosight.ai/galio-sdk-go/exporters/otlp/traces/tracestate"
	ts "galiosight.ai/galio-sdk-go/exporters/otlp/traces/tracestate"
	"galiosight.ai/galio-sdk-go/lib/bloom"
	"galiosight.ai/galio-sdk-go/model"
	semconv "galiosight.ai/galio-sdk-go/semconv/v1.0.0"
)

// go test -count=1 -v -test.run Test_WithFraction
func Test_WithFraction(t *testing.T) {
	defaultAdaptiveOptions := defaultOptions()
	WithFraction(0)(&defaultAdaptiveOptions)
	assert.Equal(t, uint64(0), defaultAdaptiveOptions.traceIDUpperBound)

	WithFraction(1)(&defaultAdaptiveOptions)
	assert.Equal(t, uint64(9223372036854775808), defaultAdaptiveOptions.traceIDUpperBound)
}

func tr(d sdktrace.SamplingDecision) sdktrace.SamplingResult {
	return sdktrace.SamplingResult{Decision: d}
}

func TestMergeSamplingResult(t *testing.T) {
	rs := sdktrace.RecordAndSample
	r := sdktrace.RecordOnly
	d := sdktrace.Drop
	wf := func(m bool) ts.WorkflowState {
		r := tracestate.WorkflowDrop
		if m {
			r = tracestate.WorkflowSample
		}
		return ts.WorkflowState{Result: r}
	}
	s := func(s ts.Strategy) ts.SampleState {
		return ts.SampleState{SampledStrategy: s}
	}
	tests := []struct {
		wf   sdktrace.SamplingDecision
		wfts ts.WorkflowState
		u    sdktrace.SamplingDecision
		uts  ts.SampleState
		res  sdktrace.SamplingDecision
		rts  string // result ts
	}{
		{rs, wf(true), rs, s(ts.StrategyDyeing), rs, "w:2;s:2"},
		{d, wf(false), rs, s(ts.StrategyMatch), rs, "w:1;s"},     // workflow 采样丢弃，以用户采样为准
		{d, wf(false), r, s(ts.StrategyMatch), r, "w:1;s"},       // workflow 采样丢弃，以用户采样为准
		{d, wf(false), d, s(ts.StrategyNotMatch), d, "w:1"},      // workflow 采样丢弃，以用户采样为准
		{rs, wf(true), d, s(ts.StrategyNotMatch), rs, "w:2"},     // 用户采样丢弃，以 workflow 采样为准
		{rs, wf(true), r, s(ts.StrategyMatch), rs, "w:2;s"},      // 用户采样仅记录，要把 workflow ts 状态合入
		{rs, wf(true), rs, s(ts.StrategyNotMatch), rs, "w:2"},    // 用户和 workflow 都采样，用户 ts 空，workflow ts 非空
		{rs, wf(false), rs, s(ts.StrategyMatch), rs, "w:1;s"},    // 用户和 workflow 都采样，用户 ts 非空，workflow ts 空
		{rs, wf(false), rs, s(ts.StrategyDyeing), rs, "w:1;s:2"}, // 用户和 workflow 都采样，用户 ts 非空，workflow ts 空
		{rs, wf(true), rs, s(ts.StrategyMatch), rs, "w:2;s"},     // 用户和 workflow 都采样，用户 ts 非空，workflow ts 非空，需合并 ts
	}

	a := assert.New(t)
	for _, test := range tests {
		t.Run(
			"", func(t *testing.T) {
				a.Equal(test.res, mergeDecision(tr(test.wf), tr(test.u)).Decision)
				ts, _ := ts.Parse("")
				ts.Workflow = test.wfts
				ts.Sample = test.uts
				a.Equal(test.rts, ts.String())
			},
		)
	}
}

func randTraceID() trace.TraceID {
	id := trace.TraceID([16]byte{})
	for i := range id {
		id[i] = byte(rand.Int31n(256))
	}
	return id
}

func Test_adaptiveSampler_ShouldSample_Fraction0(t *testing.T) {
	s := NewAdaptiveSampler(
		WithMinSampleCount(0),
		WithFraction(0),
	)
	p := sdktrace.SamplingParameters{
		Attributes: []attribute.KeyValue{
			semconv.TrpcCallerServiceKey.String("trpc.galileo.apiserver.apiserver"),
			semconv.TrpcCallerMethodKey.String("collectBusiness"),
			semconv.TrpcCalleeServiceKey.String("trpc.galileo.apiserver.MetricData"),
			semconv.TrpcCalleeMethodKey.String("getBusinessOperData"),
		},
	}
	result := s.ShouldSample(p)
	assert.Equal(t, sdktrace.Drop, result.Decision)
}

// Test_adaptiveSampler_ShouldSample_onlyWf 测试自适应只有 wf 采样器的逻辑
func Test_adaptiveSampler_ShouldSample_onlyWf(t *testing.T) {
	c := &model.WorkflowSamplerConfig{
		SampleCountPerMinute: 10,
		MaxCountPerMinute:    100,
	}
	t.Run(
		"", func(t *testing.T) {
			s := NewAdaptiveSampler(WithWorkflow(c))
			p := sdktrace.SamplingParameters{
				Attributes: []attribute.KeyValue{
					semconv.TrpcCallerServiceKey.String("trpc.galileo.apiserver.apiserver"),
					semconv.TrpcCallerMethodKey.String("collectBusiness"),
					semconv.TrpcCalleeServiceKey.String("trpc.galileo.apiserver.MetricData"),
					semconv.TrpcCalleeMethodKey.String("getBusinessOperData"),
				},
			}
			result := s.ShouldSample(p)
			assert.Equal(t, sdktrace.RecordAndSample, result.Decision)
			assert.Equal(t, "w:1;s:3;r:3", result.Tracestate.Get(galileoVendor))
		},
	)
}

// Test_adaptiveSampler_ShouldSample_only_Deying 测试自适应只有染色采样器的逻辑
func Test_adaptiveSampler_ShouldSample_only_Deying(t *testing.T) {
	t.Run(
		"", func(t *testing.T) {
			s := NewAdaptiveSampler(
				WithDeferredSample(true),
			)
			rule := []model.Dyeing{
				{
					Key: "uin",
					Values: []string{
						"1",
					},
				},
			}
			sampleConf := ocp.DefaultConfig("").TracesConfig.Processor
			sampleConf.Sampler.Dyeing = rule
			s.UpdateConfig(updateSamplerOption(&sampleConf)...)
			p := sdktrace.SamplingParameters{
				Attributes: []attribute.KeyValue{
					semconv.TrpcCallerServiceKey.String("trpc.galileo.apiserver.apiserver"),
					semconv.TrpcCallerMethodKey.String("collectBusiness"),
					semconv.TrpcCalleeServiceKey.String("trpc.galileo.apiserver.MetricData"),
					semconv.TrpcCalleeMethodKey.String("getBusinessOperData"),
					attribute.Key("uin").String("1"),
				},
			}
			result := s.ShouldSample(p)
			assert.Equal(t, sdktrace.RecordAndSample, result.Decision)
		},
	)
}

func Test_adaptiveSampler_ShouldSample(t *testing.T) {
	type fields struct {
		enableMinSample   bool    // 是否开启最小采样
		minSampleCount    int32   // 最小采集数
		fraction          float64 // 采样率
		enableBloomDyeing bool    // 是否开启布隆过滤器染色
		enableWorkflow    bool    // 是否开启 Workflow 采样器
	}

	tests := []struct {
		name   string
		fields fields
		want   sdktrace.SamplingResult
	}{
		{
			name: "enableMinSample",
			fields: fields{
				enableMinSample: true,
				minSampleCount:  2,
				fraction:        0,
			},
			want: recordAndSample,
		},
		{
			name: "enableMinSample",
			fields: fields{
				enableMinSample: true,
				minSampleCount:  0,
				fraction:        0,
			},
			want: drop,
		},
		{
			name: "enableMinSample",
			fields: fields{
				enableMinSample: false,
				minSampleCount:  2,
				fraction:        0,
			},
			want: drop,
		},
		{
			name: "enableMinSample",
			fields: fields{
				enableMinSample: false,
				minSampleCount:  2,
				fraction:        -1,
			},
			want: drop,
		},
		{
			name: "enableMinSample",
			fields: fields{
				enableMinSample: false,
				minSampleCount:  2,
				fraction:        1,
			},
			want: recordAndSample,
		},
		{
			name: "enableMinSample",
			fields: fields{
				enableMinSample: false,
				minSampleCount:  2,
				fraction:        2,
			},
			want: recordAndSample,
		},
		{
			name: "enableBloomDyeing",
			fields: fields{
				enableMinSample:   false,
				minSampleCount:    2,
				fraction:          2,
				enableBloomDyeing: true,
			},
			want: recordAndSample,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name+fmt.Sprintf("_%v", tt.fields), func(t *testing.T) {
				s := NewAdaptiveSampler(
					WithEnableMinSample(tt.fields.enableMinSample),
					WithMinSampleCount(tt.fields.minSampleCount),
					WithFraction(tt.fields.fraction),
				)
				s.opts.enableBloomDyeing = tt.fields.enableBloomDyeing
				p := sdktrace.SamplingParameters{
					Attributes: []attribute.KeyValue{
						semconv.TrpcCallerServiceKey.String("trpc.galileo.apiserver.apiserver"),
						semconv.TrpcCallerMethodKey.String("collectBusiness"),
						semconv.TrpcCalleeServiceKey.String("trpc.galileo.apiserver.MetricData"),
						semconv.TrpcCalleeMethodKey.String("getBusinessOperData"),
					},
				}
				result := s.ShouldSample(p)
				assert.Equal(t, tt.want.Decision, result.Decision)
			},
		)
	}
}

func Test_adaptiveSampler_UpdateConfig(t *testing.T) {
	type fields struct {
		opts        adaptiveOptions
		dyeing      *dyeingSampler
		bloomDyeing *bloomDyeingSampler
	}
	type args struct {
		s *model.SamplerConfig
		w model.WorkflowSamplerConfig
	}
	tests := []struct {
		name                  string
		fields                fields
		args                  args
		wantDyeing            map[string]map[string]bool
		wantEnableBloomDyeing bool
		wantBloomDyeing       map[string]*bloom.BloomFilter
	}{
		{
			name: "开启布隆过滤器染色",
			fields: fields{
				opts:        defaultOptions(),
				dyeing:      new(dyeingSampler),
				bloomDyeing: new(bloomDyeingSampler),
			},
			args: args{
				s: &model.SamplerConfig{
					EnableBloomDyeing: true,
					BloomDyeing: []model.BloomDyeing{
						{
							Key:        "uin",
							BitSize:    64,
							HashNumber: 14,
							Bitmap:     []int64{111},
						},
					},
				},
				w: model.WorkflowSamplerConfig{},
			},
			wantDyeing:            map[string]map[string]bool{},
			wantEnableBloomDyeing: true,
			wantBloomDyeing: map[string]*bloom.BloomFilter{
				"uin": bloom.From(64, 14, []int64{111}),
			},
		},
		{
			name: "未开启布隆过滤器染色",
			fields: fields{
				opts:        defaultOptions(),
				dyeing:      new(dyeingSampler),
				bloomDyeing: new(bloomDyeingSampler),
			},
			args: args{
				s: &model.SamplerConfig{
					Dyeing: []model.Dyeing{
						{
							Key:    "uin",
							Values: []string{"aaa", "bbb"},
						},
					},
					EnableDyeing:      true,
					EnableBloomDyeing: false,
				},
				w: model.WorkflowSamplerConfig{},
			},
			wantDyeing: map[string]map[string]bool{
				"uin": {
					"aaa": true,
					"bbb": true,
				},
			},
			wantEnableBloomDyeing: false,
			wantBloomDyeing:       map[string]*bloom.BloomFilter{},
		},
		{
			name: "未开启布隆过滤器染色，测试 workflow 更新采样配置",
			fields: fields{
				opts:        defaultOptions(),
				dyeing:      new(dyeingSampler),
				bloomDyeing: new(bloomDyeingSampler),
			},
			args: args{
				s: &model.SamplerConfig{
					Dyeing: []model.Dyeing{
						{
							Key:    "uin",
							Values: []string{"aaa", "bbb"},
						},
					},
					EnableDyeing:      true,
					EnableBloomDyeing: false,
				},
				w: model.WorkflowSamplerConfig{
					SampleCountPerMinute: 1,
					MaxCountPerMinute:    10,
				},
			},
			wantDyeing: map[string]map[string]bool{
				"uin": {
					"aaa": true,
					"bbb": true,
				},
			},
			wantEnableBloomDyeing: false,
			wantBloomDyeing:       map[string]*bloom.BloomFilter{},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				a := NewAdaptiveSampler()
				a.opts = tt.fields.opts
				a.dyeing = tt.fields.dyeing
				a.bloomDyeing = tt.fields.bloomDyeing
				a.UpdateConfig(
					updateSamplerOption(
						&model.TracesProcessor{
							Sampler: *tt.args.s, WorkflowSampler: tt.args.w,
						},
					)...,
				)
				assert.Equal(t, tt.wantDyeing, a.dyeing.dyeingRules.Load())
				assert.Equal(t, tt.wantEnableBloomDyeing, a.opts.enableBloomDyeing)
				assert.Equal(t, tt.wantBloomDyeing, a.bloomDyeing.bloomDyeingRules.Load())
			},
		)
	}
}

func TestAdaptiveWorkflowPath(t *testing.T) {
	// 主要单测在 TestSuite.TestShouldSample，这里只是为了覆盖
	as := NewAdaptiveSampler()
	as.path = NewWorkflowPathSampler(mkcfg(8, 2, 120))
	p := genP(kc, "A", "a", "B", "b")
	s := genS("")

	a := assert.New(t)
	a.Equal(sdktrace.RecordOnly, as.workflowSample(&p, &s).Decision)
	a.Equal(tracestate.WorkflowPath, s.Result)
}

func TestFollowParent(t *testing.T) {
	as := NewAdaptiveSampler(WithEnableMinSample(false))
	p := gen.param(nil, trace.FlagsSampled)
	a := assert.New(t)
	ts, err := tracestate.Parse("") // 上游非新 SDK
	a.NoError(err)
	a.Equal(recordAndSample, as.userSample(&p, &ts.Sample)) // 旧逻辑
	a.Equal(tracestate.StrategyFollow, ts.Sample.SampledStrategy)

	ts, _ = tracestate.Parse("w:4") // 上游是新 SDK，并且未采样，即使 FlagSampled，user 采样也不命中 FollowParent
	a.Equal(drop, as.userSample(&p, &ts.Sample))
	a.Equal(tracestate.StrategyNotMatch, ts.Sample.SampledStrategy)
}

func Test_randomSampler(t *testing.T) {
	randomConf := randomSampler{
		fraction: 0.5,
	}
	a := &adaptiveSampler{
		opts: adaptiveOptions{
			randomSampleConf: randomConf,
		},
	}
	fraction := 0.5
	var cnt float64 = 10000.0
	var sampleCnt float64 = 0
	for i := 0.0; i < cnt; i++ {
		id := randTraceID()
		p := &sdktrace.SamplingParameters{
			TraceID: id,
		}
		sampler := a.opts.randomSampleConf.randomSampler(p)
		if sampler {
			sampleCnt++
		}
	}
	assert.True(t, math.Abs(sampleCnt/cnt-fraction) < 0.05)
}

func TestWithServerAndClient(t *testing.T) {
	c := model.RpcSamplingConfig{
		Fraction: 0.5,
		Rpc: []model.RpcConfig{
			{
				Name:     "name1",
				Fraction: 0.3,
			},
		},
	}
	defaultAdaptiveOptions := defaultOptions()
	WithServer(c)(&defaultAdaptiveOptions)
	serverConf := rpcSamplingConfig{
		fraction:       0.5,
		methodFraction: map[string]float64{"name1": 0.3},
	}
	assert.Equal(t, serverConf, defaultAdaptiveOptions.randomSampleConf.server)

	clientConf1 := rpcSamplingConfig{
		fraction:       0.5,
		methodFraction: map[string]float64{"name1": 0.3},
	}
	WithClient(c)(&defaultAdaptiveOptions)
	assert.Equal(t, clientConf1, defaultAdaptiveOptions.randomSampleConf.client)

	defaultAdaptiveOptions = defaultOptions()
	realClient := defaultAdaptiveOptions.randomSampleConf.client
	if realClient.fraction != -1.0 || len(realClient.methodFraction) != 0 {
		t.Errorf("wrong when without using WithClient")
	}
}

func Test_randomSampler_getFraction(t *testing.T) {
	r := randomSampler{
		fraction: 0.5,
		server: rpcSamplingConfig{
			methodFraction: map[string]float64{"name1": 0.1},
			fraction:       -1.0,
		},
		client: rpcSamplingConfig{
			methodFraction: map[string]float64{"name2": 0.2},
			fraction:       -1.0,
		},
	}
	if r.getFraction(&sdktrace.SamplingParameters{Name: "name1", Kind: trace.SpanKindServer}) != 0.1 {
		t.Errorf("server interface fraction incorrect")
	}
	if r.getFraction(&sdktrace.SamplingParameters{Name: "nonexist", Kind: trace.SpanKindServer}) != 0.5 {
		t.Errorf("global fraction incorrect")
	}
	if r.getFraction(&sdktrace.SamplingParameters{Name: "name2", Kind: trace.SpanKindClient}) != 0.2 {
		t.Errorf("client interface fraction incorrect")
	}
	if r.getFraction(&sdktrace.SamplingParameters{Name: "nonexist", Kind: trace.SpanKindClient}) != 0.5 {
		t.Errorf("global fraction incorrect")
	}

}

func Test_rpcSamplingConfig_getFraction(t *testing.T) {
	r := rpcSamplingConfig{
		fraction:       0.5,
		methodFraction: map[string]float64{"name1": 0.2},
	}
	frac1 := r.getFraction(&sdktrace.SamplingParameters{Name: "name1"}, 0.3)
	log.Printf("fraction1: %+v", frac1)
	if frac1 != 0.2 {
		t.Errorf("interface fraction incorrect, want: %+v, now: %+v", 0.2, frac1)
	}
	frac2 := r.getFraction(&sdktrace.SamplingParameters{Name: "name2"}, 0.3)
	log.Printf("fraction2: %+v", frac2)
	if frac2 != 0.5 {
		t.Errorf("fraction incorrect, want: %+v, now: %+v", 0.5, frac2)
	}
	r = rpcSamplingConfig{
		fraction:       -1.0,
		methodFraction: map[string]float64{"name1": 0.2},
	}
	frac3 := r.getFraction(&sdktrace.SamplingParameters{Name: "name2"}, 0.3)
	log.Printf("fraction2: %+v", frac3)
	if frac3 != 0.3 {
		t.Errorf("fraction incorrect, want: %+v, now: %+v", 0.3, frac3)
	}
	r = rpcSamplingConfig{
		fraction: 0.5,
	}
	frac4 := r.getFraction(&sdktrace.SamplingParameters{Name: "name2"}, 0.3)
	log.Printf("fraction2: %+v", frac4)
	if frac4 != 0.5 {
		t.Errorf("fraction incorrect, want: %+v, now: %+v", 0.5, frac4)
	}
}

func Test_randomSampler_randomSampler(t *testing.T) {
	r := randomSampler{
		fraction: 0.05,
	}
	cnt := 0
	for i := 1; i <= 1000000; i++ {
		id := randTraceID()
		p := sdktrace.SamplingParameters{TraceID: id}
		if r.randomSampler(&p) {
			cnt++
		}
	}
	if math.Abs(float64(cnt)/1000000.0-0.05) > 0.001 {
		t.Errorf("fraction incorrect, real fraction: %+v", float64(cnt)/100000.0)
	}
}
