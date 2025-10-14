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
	"sync/atomic"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"galiosight.ai/galio-sdk-go/lib/bloom"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/semconv"
)

type dyeingSampler struct {
	// dyeingRules 存储的是染色 key:value，类型 map[string]map[string]bool
	dyeingRules atomic.Value
}

// ShouldSample 是否采样。
func (d *dyeingSampler) ShouldSample(p *sdktrace.SamplingParameters) bool {
	rules, ok := d.dyeingRules.Load().(map[string]map[string]bool)
	if !ok {
		return false
	}
	if len(rules) == 0 {
		return false
	}
	for _, attr := range p.Attributes {
		key := string(attr.Key)
		value := attr.Value.AsString()
		if attr.Key == semconv.TraceForceSampleKey && value != "" {
			// 当染色规则为空时，不会执行到这行代码。
			// 所以用户必须要在伽利略平台上配置 trace.force.sample = 1 的染色规则，才能强制染色。
			// TODO 未来移除
			return true
		}
		if rules[key][value] {
			return true
		}
	}
	return false
}

// UpdateConfig 更新配置。
func (d *dyeingSampler) UpdateConfig(enable bool, data []model.Dyeing) {
	rules := map[string]map[string]bool{}
	if !enable {
		d.dyeingRules.Store(rules)
		return
	}
	for _, k := range data {
		m, ok := rules[k.Key]
		if !ok {
			m = map[string]bool{}
			rules[k.Key] = m
		}
		for _, v := range k.Values {
			m[v] = true
		}
	}
	d.dyeingRules.Store(rules)
}

type bloomDyeingSampler struct {
	// dyeingRules 存储的是染色 key:value，类型 map[string]map[string]bool
	bloomDyeingRules atomic.Value
}

// ShouldSample 是否采样。
func (b *bloomDyeingSampler) ShouldSample(p *sdktrace.SamplingParameters) bool {
	rules, ok := b.bloomDyeingRules.Load().(map[string]*bloom.BloomFilter)
	if !ok {
		return false
	}
	if len(rules) == 0 {
		return false
	}
	for _, attr := range p.Attributes {
		key := string(attr.Key)
		if attr.Key == semconv.TraceForceSampleKey && attr.Value.AsString() != "" {
			return true
		}
		bloomFilter, exist := rules[key]
		if !exist {
			continue
		}
		if bloomFilter.Test(attr.Value.AsString()) {
			return true
		}
	}
	return false
}

// UpdateConfig 更新配置。
func (b *bloomDyeingSampler) UpdateConfig(enable bool, data []model.BloomDyeing) {
	rules := map[string]*bloom.BloomFilter{}
	if !enable {
		b.bloomDyeingRules.Store(rules)
		return
	}
	for _, dyeing := range data {
		rules[dyeing.Key] = bloom.From(dyeing.BitSize, dyeing.HashNumber, dyeing.Bitmap)
	}
	b.bloomDyeingRules.Store(rules)
}
