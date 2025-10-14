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
	"encoding/binary"
	"time"

	"github.com/alphadose/haxmap"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"galiosight.ai/galio-sdk-go/exporters/otlp/traces/internal"
	"galiosight.ai/galio-sdk-go/exporters/otlp/traces/tracestate"
	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
	selflog "galiosight.ai/galio-sdk-go/self/log"
	"galiosight.ai/galio-sdk-go/self/metric"
	"galiosight.ai/galio-sdk-go/semconv"
)

var (
	recordAndSample = sdktrace.SamplingResult{Decision: sdktrace.RecordAndSample}
	recordOnly      = sdktrace.SamplingResult{Decision: sdktrace.RecordOnly}
	drop            = sdktrace.SamplingResult{Decision: sdktrace.Drop}
)

type rpcSamplingConfig struct {
	fraction       float64
	methodFraction map[string]float64
}

// adaptiveOptions 自适应采样器的配置参数
type adaptiveOptions struct {
	traceIDUpperBound uint64        // 采样率生成的中间值，确定采样的 traceID 的上界
	enableMinSample   bool          // 是否开启最小采样
	minSampleCount    int32         // 最小采样数
	windowInterval    time.Duration // 最小采样数统计时间窗
	windowMaxKeyCount int32         // 最大采样 key 的保存数，超过要走默认的采样逻辑，避免使用过多内存
	enableDyeing      bool
	dyeing            []model.Dyeing      // 染色元数据
	deferredSample    bool                // 是否开启延迟采样
	enableBloomDyeing bool                // 是否开启布隆过滤器染色
	bloomDyeing       []model.BloomDyeing // 布隆过滤器染色元数据
	workflow          *model.WorkflowSamplerConfig
	limit             []*model.TokenBucketConfig
	randomSampleConf  randomSampler
}

// AdaptiveSamplerOption 选项应用函数。
type AdaptiveSamplerOption func(*adaptiveOptions)

// WithFraction 设置采样率
func WithFraction(fraction float64) AdaptiveSamplerOption {
	return func(options *adaptiveOptions) {
		if fraction > 1 {
			fraction = 1
		}
		if fraction < 0 {
			fraction = 0
		}
		options.randomSampleConf.fraction = fraction
		options.traceIDUpperBound = upperBound(fraction)
	}
}

// upperBound 根据采样率，计算出采样的 traceID 的上界。
func upperBound(fraction float64) uint64 {
	return uint64(fraction * (1 << 63))
}

// WithEnableMinSample 设置是否最小采样
func WithEnableMinSample(enableMinSample bool) AdaptiveSamplerOption {
	return func(options *adaptiveOptions) {
		options.enableMinSample = enableMinSample
	}
}

// WithMinSampleCount 设置每个维度在采样窗口内的最小采样数
func WithMinSampleCount(minSampleCount int32) AdaptiveSamplerOption {
	return func(options *adaptiveOptions) {
		options.minSampleCount = minSampleCount
	}
}

// WithDeferredSample 开启延迟采样
func WithDeferredSample(enable bool) AdaptiveSamplerOption {
	return func(options *adaptiveOptions) {
		options.deferredSample = enable
	}
}

// WithWindowInterval 设置采样窗口时长
func WithWindowInterval(windowInterval time.Duration) AdaptiveSamplerOption {
	return func(options *adaptiveOptions) {
		options.windowInterval = windowInterval
	}
}

// WithDyeing 设置染色数据
func WithDyeing(enable bool, d []model.Dyeing) AdaptiveSamplerOption {
	return func(options *adaptiveOptions) {
		options.enableDyeing = enable
		options.dyeing = d
	}
}

// WithWorkflow 设置开启构建 Workflow 拓扑的采样器
func WithWorkflow(c *model.WorkflowSamplerConfig) AdaptiveSamplerOption {
	return func(o *adaptiveOptions) {
		o.workflow = c
	}
}

// WithLimiter 设置过载保护配置
func WithLimiter(c []*model.TokenBucketConfig) AdaptiveSamplerOption {
	return func(o *adaptiveOptions) {
		o.limit = c
	}
}

// WithBloomDyeing 设置布隆过滤器染色
func WithBloomDyeing(enable bool, c []model.BloomDyeing) AdaptiveSamplerOption {
	return func(o *adaptiveOptions) {
		o.enableBloomDyeing = enable
		o.bloomDyeing = c
	}
}

func createRpcSamplingConfig(c model.RpcSamplingConfig) rpcSamplingConfig {
	methodFraction := make(map[string]float64)
	for _, conf := range c.Rpc {
		methodFraction[conf.Name] = conf.Fraction
	}
	return rpcSamplingConfig{
		fraction:       c.Fraction,
		methodFraction: methodFraction,
	}
}

// WithServer 设置被调 trace 采样率
func WithServer(c model.RpcSamplingConfig) AdaptiveSamplerOption {
	return func(o *adaptiveOptions) {
		o.randomSampleConf.server = createRpcSamplingConfig(c)
	}
}

// WithClient 设置主调 trace 采样率
func WithClient(c model.RpcSamplingConfig) AdaptiveSamplerOption {
	return func(o *adaptiveOptions) {
		o.randomSampleConf.client = createRpcSamplingConfig(c)
	}
}

// NewAdaptiveSampler 构建自适应采样器
func NewAdaptiveSampler(opts ...AdaptiveSamplerOption) *adaptiveSampler {
	as := &adaptiveSampler{
		opts:        defaultOptions(),
		dyeing:      new(dyeingSampler),
		bloomDyeing: new(bloomDyeingSampler),
		path:        NewWorkflowPathSampler(&model.WorkflowSamplerConfig{}),
	}
	as.user = UserSampleFunc(as.defaultSample)
	as.UpdateConfig(opts...)
	for i := 0; i < int(WindowSize); i++ {
		as.windows[i] = &samplerWindow{maxKeyCount: as.opts.windowMaxKeyCount, keys: haxmap.New[string, int32]()}
	}
	as.start()
	return as
}

func defaultOptions() adaptiveOptions {
	options := adaptiveOptions{
		randomSampleConf: randomSampler{
			fraction: 0.001, // 默认采样率千分之一
			server: rpcSamplingConfig{
				fraction:       -1.0, // 默认采样率为 -1，即默认使用全局采样率
				methodFraction: make(map[string]float64),
			},
			client: rpcSamplingConfig{
				fraction:       -1.0, // 默认采样率为 -1，即默认使用全局采样率
				methodFraction: make(map[string]float64),
			},
		},
		enableMinSample:   true,
		minSampleCount:    1, // 默认最小采样数 1
		windowInterval:    time.Minute,
		windowMaxKeyCount: 20000, // 时间窗内保存的最大维度组合数
	}
	WithFraction(options.randomSampleConf.fraction)(&options)
	return options
}

type adaptiveSampler struct {
	opts        adaptiveOptions
	ticker      *time.Ticker
	closed      chan bool
	dyeing      *dyeingSampler
	bloomDyeing *bloomDyeingSampler
	path        *WorkflowPathSampler
	limit       limiter
	user        UserSampler
	customUser  bool // 用户是否自定义了采样器

	minCountWindow
}

const galileoVendor = "g" // 伽利略 SDK traceState 的 vendor Key

// ShouldSample 是否采样。包含用户采样和 workflow 采样。
func (a *adaptiveSampler) ShouldSample(p sdktrace.SamplingParameters) sdktrace.SamplingResult {
	psc := trace.SpanContextFromContext(p.ParentContext)
	state := psc.TraceState()
	parsed, _ := tracestate.Parse(state.Get(galileoVendor))

	res := mergeDecision(a.workflowSample(&p, &parsed.Workflow), a.userSample(&p, &parsed.Sample))
	if parsed.Sample.RootStrategy <= tracestate.StrategyNotMatch && parsed.Sample.Sampled(psc) {
		// 修复 root，如果上游是旧 SDK。通常 root 可以从 sampled 继承，除了 Follow 一种情况，
		// Follow 表示实际的上游不可知，因此赋值为 kMatch 表示命中某种采样策略
		parsed.Sample.RootStrategy = parsed.Sample.SampledStrategy
		if parsed.Sample.SampledStrategy == tracestate.StrategyFollow {
			parsed.Sample.RootStrategy = tracestate.StrategyMatch
		}
	}
	ts := internal.Convert(&state).Insert(galileoVendor, parsed.String())
	res.Tracestate = *ts.Convert()
	res.Attributes = append(res.Attributes, semconv.GalileoStateKey.String(ts.String()))
	return res
}

func mergeDecision(w, u sdktrace.SamplingResult) sdktrace.SamplingResult {
	r := w.Decision | u.Decision // decision 是 int，不是 flag，or 完需要修正下
	if r == (sdktrace.RecordAndSample | sdktrace.RecordOnly) {
		r = sdktrace.RecordAndSample
	}
	w.Decision = r
	w.Attributes = append(w.Attributes, u.Attributes...)
	return w
}

func (a *adaptiveSampler) workflowSample(
	p *sdktrace.SamplingParameters, state *tracestate.WorkflowState,
) (ret sdktrace.SamplingResult) {
	var buf WorkflowPathBuffer
	if res := a.path.ShouldSample(p, state, &buf); res.Decision != sdktrace.Drop { // record only 和 recorad sample 都支持
		state.Result = tracestate.WorkflowPath
		metric.GetSelfMonitor().Stats.TracesStats.WorkflowPathCounter.Inc()
		return res
	}

	state.Result = tracestate.WorkflowDrop
	return drop
}

// UserSampler 是用户自定义采样器，用户可以通过 API 覆盖伽利略的采样器
type UserSampler interface {
	// ShouldSample p 是传入的 trace 的字段，用户采样器中能使用的输入都在这里，state 是状态标记，通常用户不需要做处理，直接透传即可。
	// 返回值，返回 RecordAndSample 表示命中采样，Drop 表示丢弃采样
	ShouldSample(p *sdktrace.SamplingParameters, state *tracestate.SampleState) sdktrace.SamplingResult
}

type UserSampleFunc func(p *sdktrace.SamplingParameters, state *tracestate.SampleState) sdktrace.SamplingResult

// ShouldSample 是用户自定义采样器，用户可以通过 API 覆盖伽利略的采样器
func (f UserSampleFunc) ShouldSample(p *sdktrace.SamplingParameters, state *tracestate.SampleState) sdktrace.SamplingResult {
	return f(p, state)
}

const samplingStage = "galileo-internal-user-sampling-stage"

func isUserSamplingStage(attr []attribute.KeyValue) bool {
	for _, k := range attr {
		if k.Key == samplingStage {
			return true
		}
	}
	return false
}

func clearUserSamplingStage(attr []attribute.KeyValue) []attribute.KeyValue {
	i := 0
	for _, kv := range attr {
		if kv.Key != samplingStage {
			attr[i] = kv
			i++
		}
	}
	return attr[:i]
}

func (a *adaptiveSampler) userSample(p *sdktrace.SamplingParameters,
	state *tracestate.SampleState,
) (ret sdktrace.SamplingResult) {
	if !a.customUser {
		// 其实可以合并到下面的逻辑，分开实现只是担心在原来链路上增加的逻辑会有 bug
		// 是否是 customUser 并不影响最终采样策略
		return a.user.ShouldSample(p, state)
	}
	p.Attributes = append(p.Attributes, attribute.Bool(samplingStage, true))
	user := a.user.ShouldSample(p, state)
	if user.Decision == recordAndSample.Decision && isUserSamplingStage(p.Attributes) {
		// 我们的函数如果是命中采样，每个分支都会正确的设置。但如果没有设置，可以推断是用户覆盖了采样器，并且没有设置 state
		// 因此设置采样策略为用户自定义采样
		state.SampledStrategy = tracestate.StrategyUser
	}
	return user
}

// defaultSample 用户采样策略，由于后续会 merge 采样 tracestate，所以非 Drop 的不返回全局对象
func (a *adaptiveSampler) defaultSample(
	p *sdktrace.SamplingParameters,
	state *tracestate.SampleState,
) (ret sdktrace.SamplingResult) {
	// 之后的阶段不是用户采样，清除标记
	p.Attributes = clearUserSamplingStage(p.Attributes)
	// 继承上游采样结果
	if psc := trace.SpanContextFromContext(p.ParentContext); psc.IsSampled() && state.Sampled(psc) {
		if !a.limit.Consume(state.RootStrategy) {
			state.SampledStrategy = tracestate.StrategyNotMatch
			// 从此处断开了，后续流量可以重新作为根，因此清空 root
			state.RootStrategy = tracestate.StrategyNotMatch
			metric.GetSelfMonitor().Stats.TracesStats.LimitDropCounter.Inc()
			return drop
		}
		state.SampledStrategy = tracestate.StrategyFollow
		return recordAndSample
	}
	// 是否命中染色
	if a.matchDyeing(p) {
		state.SampledStrategy = tracestate.StrategyDyeing
		return recordAndSample
	}
	// 最小采样逻辑，确保每个接口组合都能被采集到至少 minSampleCount 个 trace
	if a.minCountSampler(p) {
		// 命中最低采样策略，采样
		state.SampledStrategy = tracestate.StrategyMinCount
		return recordAndSample
	}
	// 按采样率采样
	if a.opts.randomSampleConf.randomSampler(p) {
		// 命中，采样
		state.SampledStrategy = tracestate.StrategyRandom
		return recordAndSample
	}
	// 开启了延迟采样，先记录，等后面命中延迟采样规则，再进行采样
	if a.opts.deferredSample {
		return recordOnly
	}
	// 默认丢弃，不采样
	state.SampledStrategy = tracestate.StrategyNotMatch
	return drop
}

func (a *adaptiveSampler) matchDyeing(p *sdktrace.SamplingParameters) bool {
	if a.opts.enableBloomDyeing {
		matchDyeing := a.bloomDyeing != nil && a.bloomDyeing.ShouldSample(p)
		if selflog.Enable(logs.LevelDebug) {
			selflog.Debugf(
				"[galileo]matchDyeing=%v,a.dyeingRule=%v,p=%v", matchDyeing, a.bloomDyeing,
				toLogMsg(p),
			)
		}
		return matchDyeing
	}
	if a.opts.enableDyeing {
		matchDyeing := a.dyeing != nil && a.dyeing.ShouldSample(p)
		if selflog.Enable(logs.LevelDebug) {
			selflog.Debugf("[galileo]matchDyeing=%v,a.dyeingRule=%v,p=%v", matchDyeing, a.dyeing, toLogMsg(p))
		}
		return matchDyeing
	}
	return false
}

type samplingParametersLogMsg struct {
	TraceID    string
	Name       string
	Attributes map[string]string
}

func toLogMsg(p *sdktrace.SamplingParameters) samplingParametersLogMsg {
	t := samplingParametersLogMsg{
		TraceID:    p.TraceID.String(),
		Name:       p.Name,
		Attributes: map[string]string{},
	}
	for _, v := range p.Attributes {
		t.Attributes[string(v.Key)] = v.Value.AsString()
	}
	return t
}

func (a *adaptiveSampler) Description() string {
	return "AdaptiveSampler"
}

// UpdateConfig 配置热更新
func (a *adaptiveSampler) UpdateConfig(opts ...AdaptiveSamplerOption) {
	for _, o := range opts {
		o(&a.opts)
	}
	a.dyeing.UpdateConfig(a.opts.enableDyeing, a.opts.dyeing)
	a.bloomDyeing.UpdateConfig(a.opts.enableBloomDyeing, a.opts.bloomDyeing)
	a.path.UpdateConfig(a.opts.workflow)
	a.limit.UpdateConfig(a.opts.limit)
}

type randomSampler struct {
	server   rpcSamplingConfig
	client   rpcSamplingConfig
	fraction float64
}

func (r *rpcSamplingConfig) getFraction(p *sdktrace.SamplingParameters, globalFraction float64) float64 {
	if fraction, ok := r.methodFraction[p.Name]; ok && fraction >= 0 {
		return fraction
	}
	if r.fraction >= 0 {
		return r.fraction
	}
	return globalFraction
}

func needSample(traceID trace.TraceID, fraction float64) bool {
	x := binary.BigEndian.Uint64(traceID[0:8]) >> 1
	return x < upperBound(fraction)
}

func (r *randomSampler) getFraction(p *sdktrace.SamplingParameters) float64 {
	if p.Kind == trace.SpanKindServer || p.Kind == trace.SpanKindConsumer {
		return r.server.getFraction(p, r.fraction)
	} else if p.Kind == trace.SpanKindClient || p.Kind == trace.SpanKindProducer {
		return r.client.getFraction(p, r.fraction)
	}
	return r.fraction
}

// randomSampler 按比例随机采样
func (r *randomSampler) randomSampler(p *sdktrace.SamplingParameters) bool {
	return needSample(p.TraceID, r.getFraction(p))
}

// start 启动 adaptiveSampler，切换到下一个时间窗并清空上一个时间窗的数据
func (a *adaptiveSampler) start() {
	a.ticker = time.NewTicker(a.opts.windowInterval)
	a.closed = make(chan bool)
	go func() {
		defer a.ticker.Stop()
		for {
			select {
			case <-a.ticker.C:
				a.shiftWindow()
			case <-a.closed:
				return
			}
		}
	}()
}

// close 关闭定时器
func (a *adaptiveSampler) close() {
	a.closed <- true
}
