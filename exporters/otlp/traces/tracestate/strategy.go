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

type Strategy int

const (
	StrategyNotExist Strategy = -1 // 不存在 user sample 标记，说明上游不是新版 SDK
	// StrategyNotMatch 每种采样是固定 id，用 itoa 容易错位
	StrategyNotMatch Strategy = 0 // 未命中采样
	StrategyMatch    Strategy = 1 // 命中某个采样
	StrategyDyeing   Strategy = 2 // 命中染色采样
	StrategyMinCount Strategy = 3 // 命中最小采样
	StrategyRandom   Strategy = 4 // 命中随机采样
	StrategyFollow   Strategy = 5 // 继承 parent 采样结果
	StrategyError    Strategy = 6 // 错误采样（后置采样）
	StrategySlow     Strategy = 7 // 慢采样（后置采样）
	StrategyUser     Strategy = 8 // 用户自定义采样
)

var strategyNameList = []string{
	"not_match", "match", "dyeing", "min_count", "random", "follow",
	"error", "slow", "user",
}

// String convert Strategy to String
func (s Strategy) String() string {
	if s == StrategyNotExist {
		return "not_exist"
	}
	if int(s) < len(strategyNameList) {
		return strategyNameList[s]
	}
	return ""
}

var strategyValueMap = map[string]Strategy{
	"dyeing":    StrategyDyeing,
	"min_count": StrategyMinCount,
	"random":    StrategyRandom,
	"follow":    StrategyFollow,
	"error":     StrategyError,
	"slow":      StrategySlow,
	"user":      StrategyUser,
}

// ParseStrategy 从字符串解析 Strategy
func ParseStrategy(name string) Strategy {
	ret, ok := strategyValueMap[name]
	if !ok {
		return StrategyNotExist
	}
	return ret
}
