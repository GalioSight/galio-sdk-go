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
	"math"
	"sync/atomic"
	"time"

	"galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/processors/omp/metrics/point"
)

// fixConfig 修正配置错误的情况。
func fixConfig(cfg *configs.Metrics) {
	if cfg.Stats == nil {
		cfg.Stats = &model.SelfMonitorStats{}
	}
	if cfg.Processor.WindowSeconds <= 0 {
		cfg.Processor.WindowSeconds = 15 // 默认 15 秒聚合窗口。
	}
	if cfg.Processor.ClearSeconds <= 0 {
		cfg.Processor.ClearSeconds = 60 * 10 // 默认 10 分钟清理一次线。
	}
	if cfg.Processor.ExpiresSeconds <= 0 {
		cfg.Processor.ExpiresSeconds = 60 * 60 // 默认 1 小时线过期。
	}
	if cfg.Processor.PointLimit <= 0 {
		cfg.Processor.PointLimit = math.MaxInt64 // 默认不限制。
	}
	if cfg.Processor.ProcessMetricsSeconds <= 0 {
		cfg.Processor.ProcessMetricsSeconds = 15 // 默认 15 秒上报一次 go runtime metrics。
	}
	if cfg.Log == nil {
		cfg.Log = logs.DefaultWrapper()
	}
	cfg.ConvBuckets()
	cfg.ConvIgnoreLabels()
	cfg.ConvSecondGranularitys()
}

// UpdateConfig 更新配置。
func (p *processor) UpdateConfig(cfg *configs.Metrics) {
	cfg.Log.Infof("[galileo]processor.UpdateConfig|cfg=%+v\n", cfg)
	fixConfig(cfg)
	cfg.Log.Infof("[galileo]processor.UpdateConfig|fix.cfg=%+v\n", cfg)
	p.exporter.UpdateConfig(cfg)                    // 导出器更新配置。
	p.setWindowSeconds(cfg.Processor.WindowSeconds) // 处理器更新聚合窗口。
	p.setBuckets(cfg.HistogramBuckets)              // 处理器更新分桶配置。
	p.setSecondLevels(cfg.SecondGranularitys)       // 秒级监控配置更新。
	p.setIgnoreLabels(cfg.Processor.LabelIgnores)   // 屏蔽配置热更新。
	p.sampler.updateConfigs(cfg.Processor.SampleMonitors)
}

func (p *processor) setSecondLevels(cfg *configs.SecondGranularitys) {
	c := cfg.Clone()
	p.cfg.SecondGranularitys = c
}

func (p *processor) setWindowSeconds(windowSeconds int32) {
	atomic.StoreInt32(&p.cfg.Processor.WindowSeconds, windowSeconds)
}

func (a *aggregator) resetTicker(windowUse time.Duration, ticker *time.Ticker) time.Duration {
	windowCfg := a.windowFunc()
	if windowUse != windowCfg { // 使用值 ≠ 最新配置值。
		windowUse = windowCfg
		ticker.Reset(windowUse)
	}
	return windowUse
}

func (p *processor) setIgnoreLabels(labelIgnores []model.LabelIgnore) {
	p.cfg.Processor.LabelIgnores = labelIgnores
	p.cfg.ConvIgnoreLabels()
}

func (p *processor) setBuckets(buckets map[string]*configs.Bucket) {
	if buckets == nil {
		return
	}
	for name, bucket := range buckets {
		p.setBucket(name, bucket)
	}
}

func (p *processor) setBucket(name string, bucket *configs.Bucket) {
	p.cfg.Mu.Lock()
	defer p.cfg.Mu.Unlock()
	oldBucket, ok := p.cfg.HistogramBuckets[name]
	if !ok {
		p.cfg.HistogramBuckets[name] = bucket
		p.cfg.Log.Infof(
			"[galileo]HistogramBuckets.change|name=%s|old=nil|new=%v\n",
			name, bucket.Values,
		)
		return
	}
	if oldBucket.Key != bucket.Key {
		p.cfg.HistogramBuckets[name] = bucket
		p.cfg.Log.Infof(
			"[galileo]HistogramBuckets.change|name=%s|old=%v|new=%v\n",
			name, oldBucket.Values, bucket.Values,
		)
		return
	}
}

// defaultBucket 默认直方图分桶。
var defaultBucket = configs.NewBucket([]float64{0.0, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10})

func (p *processor) getBucket(name string) point.BucketFunc {
	return func() *configs.Bucket {
		if p.cfg.HistogramBuckets == nil {
			return defaultBucket
		}
		p.cfg.Mu.RLock()
		defer p.cfg.Mu.RUnlock()
		bucket, ok := p.cfg.HistogramBuckets[name]
		if !ok {
			return defaultBucket
		}
		return bucket
	}
}
