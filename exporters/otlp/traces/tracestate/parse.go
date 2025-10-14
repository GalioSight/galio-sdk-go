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

// Package tracestate ...
package tracestate

import (
	"strconv"
	"strings"
)

// State 表示 galileo trace state
// tracestate: g=w;s:a1:a2:a3
// g 表示 galileo 的 vendor，虽然
// https://opentelemetry.io/docs/reference/specification/trace/tracestate-handling/ 文档中说，
// otel 生态的都用 ot，但这里还是单独使用 g，保持独立
//
// w 表示 workflow 采样，即默认兜底采样，目前没有参数
// s 表示 sample 用户定义的采样，目前没有参数
//
// 使用 tracestate.Parse(trace.ParseTraceState("g=...").Get("g")) 来解析
type State struct {
	Workflow WorkflowState
	Sample   SampleState
	ori      string // 引用原始字符串
}

// WorkflowState workflow 采样
type WorkflowState struct {
	Result     WorkflowResult
	Path       string
	ParentPath string // 上游的 caller
}

// Sampled ws 命中 workflow 采样，允许 ws 为 nil
func (ws *WorkflowState) Sampled() bool {
	return ws != nil && ws.Result >= WorkflowSample // WorkflowSample 之后的表示具体命中的采样，所以是大于等于
}

// SampleState 用户采样
type SampleState struct {
	SampledStrategy Strategy
	RootStrategy    Strategy
}

type parentContext interface {
	IsSampled() bool
}

// Sampled 使用 Sampled(p.ParentContext())
func (ss *SampleState) Sampled(parent parentContext) bool {
	if ss == nil {
		return false // impossible
	}
	if ss.SampledStrategy >= StrategyMatch {
		return true
	}
	if ss.SampledStrategy == StrategyNotExist && parent != nil && parent.IsSampled() {
		// 如果不存在，相信上游的结果
		return true
	}
	return false
}

// WorkflowResult 为了判断根节点，无论是否采样都需要透传到下游
type WorkflowResult int

const (
	WorkflowNotExist WorkflowResult = 0 // 不存在 workflow 标记
	WorkflowDrop     WorkflowResult = 1 // workflow 标记不采样
	WorkflowSample   WorkflowResult = 2 // workflow 标记命中某个采样 作为默认值
	WorkflowRandom   WorkflowResult = 3 // workflow 标记命中随机采样（废弃）
	WorkflowPath     WorkflowResult = 4 // workflow 标记命中 path 采样
)

const (
	listDelimiter   = ","
	memberDelimiter = "="
	argDelimiter    = ":"
	subDelimiter    = ";"
	workflowSubKey  = "w"
	sampleSubKey    = "s"
	rootSubKey      = "r"
)

func strcut(s, sep string) (before, after string, found bool) {
	return strings.Cut(s, sep)
}

// TakeTraceState 用于替代 trace.ParseTraceState，不进行校验
// 后者确实太慢了，等效于 trace.ParseTraceState(ts).Get(vendor)
func TakeTraceState(ts string, vendor string) string {
	for ts != "" {
		var member string
		member, ts, _ = strcut(ts, listDelimiter)
		key, val, found := strcut(member, memberDelimiter)
		// 按照 otel 的实现=不能省
		if found && key == vendor {
			return val
		}
	}
	return ""
}

func readNumber(val string, def int) int {
	num, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return num
}

// Parse w;s:arg1:arg2;r:arg3...
// 有 s 存在即表示 match，具体的 int 表示 match 的是哪个策略
// see opentelemetry-go-contrib/samplers/probability/consistent/tracestate.go
func Parse(ts string) (State, error) {
	s := State{ori: ts}
	// SDK 代码里面不要放 map，直接手工解析，保持高效
	if ts == "" {
		// 只有 ts 完全为空，用户采样才初始化为不存在，表明上游用的是旧 SDK，这里不
		// 能设置 StrategyMatch 或 StrategyNotMatch，因为上游命中或不命中 ts 都是空的。
		//
		// 即 NotExist 时需要结合 IsSampled() 才能判断，其它时候 Sampled 可以自解释。这也是 Sampled() 函数的行为
		//
		// 为什么不设计成 Parse(ts,isSampled)，因为额外增加参数会导致使用不方便，并且官
		// 方 ts 也只是 parse string 本身，没有借助额外参数。
		//
		// 最后输出到下游时 SampledStrategy 必然会设置成具体的策略，所以 NotExist
		// 不具有传染性，只在本服务内存在，不会传播给下游。
		s.Sample.SampledStrategy = StrategyNotExist
		s.Sample.RootStrategy = StrategyNotExist
	}
	if s.Sample.RootStrategy <= StrategyNotMatch && s.Sample.SampledStrategy >= StrategyMatch {
		// 数据不一样的情况：只能是上游用的是旧 SDK
		if s.Sample.SampledStrategy == StrategyFollow {
			// 上游表示是跟随，说明最原始的类型已经不可考据了。root 设置成 StrategyMatch，表示命中了采样，但不知道是啥采样。
			s.Sample.RootStrategy = StrategyMatch
		} else {
			// 上游表示具体的采样，说明上游一定是根，可以还原。
			s.Sample.RootStrategy = s.Sample.SampledStrategy
		}
	}
	for ts != "" {
		var arg string
		arg, ts, _ = strcut(ts, subDelimiter)
		key, val, _ := strcut(arg, argDelimiter)
		switch key {
		case workflowSubKey:
			res, val, _ := strcut(val, argDelimiter)
			s.Workflow.Result = WorkflowResult(readNumber(res, int(WorkflowSample)))
			s.Workflow.Path, val, _ = strcut(val, argDelimiter)
			s.Workflow.ParentPath, _, _ = strcut(val, argDelimiter)
		case sampleSubKey:
			val, _, _ = strcut(val, argDelimiter)
			s.Sample.SampledStrategy = Strategy(readNumber(val, int(StrategyMatch)))
		case rootSubKey:
			val, _, _ = strcut(val, argDelimiter)
			s.Sample.RootStrategy = Strategy(readNumber(val, int(StrategyMatch)))
		default:
		}
	}
	return s, nil
}

type builder struct {
	strings.Builder
}

func (sb *builder) semi() {
	if sb.Len() != 0 {
		sb.WriteString(subDelimiter)
	}
}

func (sb *builder) arg(x string) {
	if x != "" {
		sb.WriteString(argDelimiter)
		sb.WriteString(x)
	}
}

// omitArgument support 1 argument and can be omit
func omitArgument(b *builder, sarg, key string, num, min int) {
	if num > 0 {
		b.semi()
		b.WriteString(key)
		_, left, _ := strcut(sarg, argDelimiter)
		if left != "" || num > min {
			b.WriteString(argDelimiter)
			b.WriteString(strconv.Itoa(num))
			b.arg(left) // 剩下参数不识别的继续透传
		}
	}
}

// write argument support more argument
func writeArgument(b *builder, sarg, key string, field ...func()) {
	b.semi()
	b.WriteString(key)
	for i := 0; i < len(field); i++ {
		field[i]()
		_, sarg, _ = strcut(sarg, argDelimiter)
	}
	b.arg(sarg)
}

func (s *State) String() string {
	var b builder
	b.Grow(len(s.ori)) // 大部分修改可能不会修改长度，预先申请和之前一样的空间
	var wfarg, sarg, rarg string

	ts := s.ori
	for ts != "" {
		var arg string
		arg, ts, _ = strcut(ts, subDelimiter)
		key, val, _ := strcut(arg, argDelimiter)
		switch key {
		case workflowSubKey:
			wfarg = val
		case sampleSubKey:
			sarg = val
		case rootSubKey:
			rarg = val
		default:
			b.semi()
			b.WriteString(arg)
		}
	}

	if s.Workflow.Result != WorkflowNotExist {
		writeArgument(
			&b, wfarg, workflowSubKey, func() {
				b.WriteString(argDelimiter)
				b.WriteString(strconv.Itoa(int(s.Workflow.Result)))
			}, func() {
				if s.Workflow.Path != "" {
					b.WriteString(argDelimiter)
					b.WriteString(s.Workflow.Path)
				}
			}, func() {
				if s.Workflow.Path != "" && s.Workflow.ParentPath != "" { // 只有 parent_path，没有 path 认为无效
					b.WriteString(argDelimiter)
					b.WriteString(s.Workflow.ParentPath)
				}
			},
		)
	}
	omitArgument(&b, sarg, sampleSubKey, int(s.Sample.SampledStrategy), int(StrategyMatch))
	omitArgument(&b, rarg, rootSubKey, int(s.Sample.RootStrategy), int(StrategyMatch))
	return b.String()
}
