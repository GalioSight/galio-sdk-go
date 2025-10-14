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
	"runtime/debug"
	"testing"
	"time"

	ts "galiosight.ai/galio-sdk-go/exporters/otlp/traces/tracestate"
	"galiosight.ai/galio-sdk-go/lib/limit"
	"galiosight.ai/galio-sdk-go/model"
	"github.com/stretchr/testify/assert"
)

func TestLimiter(t *testing.T) {
	l := limiter{}
	a := assert.New(t)
	a.True(l.Consume(ts.StrategyRandom))

	limit.Tick = time.Millisecond * 200
	cfg := []*model.TokenBucketConfig{{Strategy: "default", Rate: 2, Burst: 4}}
	type tester struct {
		pass bool
		seq  []ts.Strategy
	}

	exec := func(l *limiter, test []tester) {
		for _, t := range test {
			for _, s := range t.seq {
				a.Equal(t.pass, l.Consume(s), "%s", string(debug.Stack()))
			}
		}
	}

	{
		l := &limiter{}
		l.UpdateConfig(cfg)
		test := []tester{
			{true, []ts.Strategy{ts.StrategyRandom, ts.StrategyDyeing, ts.StrategyMinCount, ts.StrategyMatch}},
			{false, []ts.Strategy{ts.StrategyDyeing}},
		}
		exec(l, test)
		limit.HighResolutionWait(limit.Tick)
		test = []tester{
			{true, []ts.Strategy{ts.StrategyDyeing, ts.StrategyRandom}},
			{false, []ts.Strategy{ts.StrategyMinCount}},
		}
		exec(l, test)
	}
	cfg = append(cfg, &model.TokenBucketConfig{Strategy: "dyeing", Rate: 1, Burst: 2})
	{
		l := &limiter{}
		l.UpdateConfig(cfg)
		test := []tester{
			{
				true, []ts.Strategy{
					ts.StrategyRandom, ts.StrategyDyeing, ts.StrategyMinCount, ts.StrategyMatch, ts.StrategyDyeing,
				},
			},
			{false, []ts.Strategy{ts.StrategyDyeing}},
			{true, []ts.Strategy{ts.StrategyRandom}},
			{false, []ts.Strategy{ts.StrategyRandom}},
		}
		exec(l, test)
		limit.HighResolutionWait(limit.Tick)
		test = []tester{
			{true, []ts.Strategy{ts.StrategyDyeing, ts.StrategyRandom, ts.StrategyMinCount}},
			{false, []ts.Strategy{ts.StrategyDyeing, ts.StrategyRandom}},
		}
		exec(l, test)
	}
}
