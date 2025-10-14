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

// Package profiles ...
package profiles

import (
	"sync"

	"galiosight.ai/galio-sdk-go/components"
	"galiosight.ai/galio-sdk-go/configs"
	otphttp "galiosight.ai/galio-sdk-go/exporters/otp/http"
	"galiosight.ai/galio-sdk-go/lib/file"
	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
)

type profilesExporter struct {
	httpExporter otphttp.HTTPExporter
	cfg          *configs.Profiles
	log          *logs.Wrapper
	queue        chan *model.ProfilesBatch
	stats        *model.SelfMonitorStats
	fileExporter *file.Exporter
	// 确保 Shutdown 只执行一次
	stopOnce sync.Once
	// 退出时，控制 exporter 退出
	stopCh chan struct{}
}

// NewExporter 根据配置创建导出器。
func NewExporter(cfg *configs.Profiles) (components.ProfilesExporter, error) {
	cfg = mergeDefaultCfg(cfg)
	exporter := &profilesExporter{
		cfg: cfg,
		httpExporter: otphttp.NewHTTPGeneralExporter(
			int(cfg.Exporter.TimeoutMs), cfg.Exporter.Collector.Addr, cfg.Log,
			otphttp.WithDirectIPPorts(cfg.Exporter.Collector.DirectIpPort),
			otphttp.WithHeaders(
				map[string]string{
					model.TenantHeaderKey: cfg.Resource.TenantId,
					model.TargetHeaderKey: cfg.Resource.Target,
					model.APIKeyHeaderKey: cfg.APIKey,
				},
			),
			otphttp.WithMaxRetryCount(cfg.Exporter.MaxRetryCount),
		),
		queue:        make(chan *model.ProfilesBatch, cfg.Exporter.BufferSize),
		log:          cfg.Log,
		stats:        cfg.Stats,
		fileExporter: file.NewExporter(cfg.Exporter.ExportToFile, "galileo/profiles", cfg.Log),
	}
	go exporter.processQueue()
	return exporter, nil
}

// Shutdown
func (p *profilesExporter) Shutdown() {
	p.stopOnce.Do(
		func() {
			p.stop()
		},
	)
}

func (p *profilesExporter) stop() {
	close(p.stopCh)
	p.drainQueue()
}

// UpdateConfig 更新配置。
func (p *profilesExporter) UpdateConfig(cfg *configs.Profiles) {
	p.httpExporter.UpdateConfig(
		cfg.Exporter.MaxRetryCount,
		cfg.Exporter.Collector,
	)
}

// mergeDefaultCfg 如果 cfg 参数未配置，使用默认配置填充。
func mergeDefaultCfg(cfg *configs.Profiles) *configs.Profiles {
	if cfg.Stats == nil {
		cfg.Stats = &model.SelfMonitorStats{}
	}
	if cfg.Exporter.BufferSize <= 0 {
		cfg.Exporter.BufferSize = 5
	}
	if cfg.Exporter.TimeoutMs <= 0 {
		cfg.Exporter.TimeoutMs = 1000
	}
	return cfg
}

// Export 将数据放到 chan 中，然后通过多个 worker 并发进行上报。
func (p *profilesExporter) Export(batch *model.ProfilesBatch) {
	p.enqueue(batch)
}

func (p *profilesExporter) enqueue(batch *model.ProfilesBatch) {
	for {
		select {
		case p.queue <- batch:
			p.stats.ProfilesStats.EnqueueCounter.Add(int64(len(batch.Profiles)))
			return
		default:
			select {
			// queue 满了，drop 掉最老的 profile batch
			case drop := <-p.queue:
				p.stats.ProfilesStats.DropCounter.Add(int64(len(drop.Profiles)))
				p.log.Debugf("[galileo]profilesExporter.queue is full, evict the oldest profile batch from the queue")
			default:
			}
		}
	}
}

func (p *profilesExporter) processQueue() {
	for {
		select {
		case <-p.stopCh:
			return
		case batch := <-p.queue:
			p.log.Infof("[galileo]profilesExporter begin Export seq %d", batch.Sequence)
			if err := p.uploadProfiles(batch); err != nil {
				p.log.Errorf("[galileo]profilesExporter|err=%v\n", err)
			}
		}
	}
}

func (p *profilesExporter) drainQueue() {
	for {
		select {
		case batch := <-p.queue:
			if err := p.uploadProfiles(batch); err != nil {
				p.log.Errorf("[galileo]profilesExporter.worker|err=%v\n", err)
			}
		default:
			close(p.queue)
		}
	}
}

func (p *profilesExporter) uploadProfiles(batch *model.ProfilesBatch) error {
	r := otphttp.NewReuseObject()
	size := int64(len(batch.Profiles))
	batchSize := int64(calBatchSize(batch))
	err := p.httpExporter.Export(batch, r)
	if err != nil {
		p.stats.ProfilesStats.FailedExportCounter.Add(size)
		p.stats.ProfilesStats.FailedWriteByteSize.Add(batchSize)
		return err
	}
	p.stats.ProfilesStats.SucceededExportCounter.Add(size)
	p.stats.ProfilesStats.SucceededWriteByteSize.Add(batchSize)
	if p.cfg.Exporter.ExportToFile {
		p.fileExporter.ExportProfilesBatch(batch)
	}
	return nil
}

func calBatchSize(batch *model.ProfilesBatch) int {
	return batch.XXX_Size()
}
