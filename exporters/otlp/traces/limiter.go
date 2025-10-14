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
	"sync/atomic"

	"galiosight.ai/galio-sdk-go/exporters/otlp/traces/tracestate"
	"galiosight.ai/galio-sdk-go/lib/limit"
	"galiosight.ai/galio-sdk-go/model"
)

type limiter struct {
	impl atomic.Value
}

type limiterImpl struct {
	defaults   *limit.TokenBucket
	rootSample map[tracestate.Strategy]*limit.TokenBucket
}

func (l *limiter) UpdateConfig(cfg []*model.TokenBucketConfig) {
	li := &limiterImpl{}
	for _, c := range cfg {
		if c.GetStrategy() == "default" {
			li.defaults = limit.New(c.GetRate(), c.GetBurst())
		} else {
			s := tracestate.ParseStrategy(c.GetStrategy())
			if s == tracestate.StrategyNotExist {
				continue
			}
			if li.rootSample == nil {
				li.rootSample = map[tracestate.Strategy]*limit.TokenBucket{}
			}
			li.rootSample[s] = limit.New(c.GetRate(), c.GetBurst())
		}
	}
	l.impl.Store(li)
}

func (l *limiter) Consume(s tracestate.Strategy) bool {
	a := l.impl.Load()
	if a == nil {
		return true
	}
	li := a.(*limiterImpl)
	if li.rootSample != nil {
		if tb := li.rootSample[s]; tb != nil {
			return tb.Consume(1)
		}
	}
	if li.defaults != nil {
		return li.defaults.Consume(1)
	}
	return true
}
