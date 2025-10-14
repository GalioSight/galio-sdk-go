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

package metrics

import (
	"math/rand"
	"sync/atomic"

	"galiosight.ai/galio-sdk-go/model"
)

// hashCodeUpperBound hashCode 上界。
// 只使用 hashCode 63bit 用于计算采样。
// 因为使用 64bit 会出现数学计算溢出的情况
const hashCodeUpperBound = 1 << 63

// sampleConfig 采样配置。
type sampleConfig struct {
	sampleType model.MetricsSampleType // 采样类型
	fraction   float64                 // 采样率。
	upperBound uint64                  // hashCode 上界。
	factor     float64                 // 放大系数。
}

// sampler 采样器。
type sampler struct {
	monitorNameToConfigs atomic.Value // map[string]*sampleConfig key：采样的监控项名称，value：采样配置。
}

func newSampler(sampleMonitors []model.SampleMonitor) *sampler {
	s := &sampler{}
	s.updateConfigs(sampleMonitors)
	return s
}

func (s *sampler) updateConfigs(sampleMonitors []model.SampleMonitor) {
	monitorNameToConfigs := make(map[string]*sampleConfig)
	for i := range sampleMonitors {
		monitorNameToConfigs[sampleMonitors[i].GetMonitorName()] =
			newSampleConfig(sampleMonitors[i].GetSampleType(), sampleMonitors[i].GetFraction())
	}
	s.monitorNameToConfigs.Store(monitorNameToConfigs)
}

func newSampleConfig(sampleType model.MetricsSampleType, fraction float64) *sampleConfig {
	if fraction == 0 {
		return &sampleConfig{sampleType: sampleType}
	}
	return &sampleConfig{
		sampleType: sampleType,
		fraction:   fraction,
		upperBound: upperBound(fraction),
		factor:     1 / fraction,
	}
}

// upperBound 根据采样率，计算出采样的 hashCode 的上界。
func upperBound(fraction float64) uint64 {
	return uint64(fraction * hashCodeUpperBound)
}

// sample 判断是否需要采样，返回放大系数和是否采样标识。
// 入参：监控项名称，所有 label hash 值（不含属性标签）。
// 默认采样，系数为 1。
// 行采样逻辑：所有 label hash 值为 hashCode。
// 通过 hashCode 上限和采样率 fraction 计算出采样 hashCode 划线值，小于划线值的 hashCode 则进行采样。
// hashCode 为 64 位，使用 64bit 会出现数学计算溢出的情况，此处仅使用 hashCode 63bit 进行计算，则上限为 1 << 63。
// 通过对入参的 hashCode 右移一位，取 hashCode 高 63 位进行计算。
// hashCode 均匀分布在 0～maxHashCode(原为 64bit 最大值，防止溢出改为 63bit) 之间。
// 通过计算得出 upperBound，0-upperBound 的 hashCode 均采样。。
// ┌──┐ maxHashCode
// │  │
// │  │ 不采样
// │  │
// │──┼── upperBound 采样划线值
// │  │
// │  │	采样
// │  │
// └──┘ 0
func (s *sampler) sample(monitorName string, labelsHashCode uint64) (float64, bool) {
	monitorNameToConfigs, ok := s.monitorNameToConfigs.Load().(map[string]*sampleConfig)
	if !ok {
		return 1, true
	}
	if len(monitorNameToConfigs) == 0 {
		return 1, true
	}
	cfg, ok := monitorNameToConfigs[monitorName]
	if !ok {
		return 1, true
	}
	switch cfg.sampleType {
	case model.MetricsSampleType_METRICS_SAMPLE_TYPE_RAND:
		return cfg.factor, rand.Float64() < cfg.fraction
	case model.MetricsSampleType_METRICS_SAMPLE_TYPE_ROWS:
		return cfg.factor, labelsHashCode>>1 < cfg.upperBound
	default:
	}
	return 1, true
}
