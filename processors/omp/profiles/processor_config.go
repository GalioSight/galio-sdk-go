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

package profiles

import (
	"sync"

	"galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/configs/ocp"
	"galiosight.ai/galio-sdk-go/configs/profiles"
	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
	yaml "gopkg.in/yaml.v3"
)

var (
	configMutex sync.Mutex
)

// fixConfig 修正配置错误的情况。
func fixConfig(cfg *configs.Profiles) {
	if cfg.Stats == nil {
		cfg.Stats = &model.SelfMonitorStats{}
	}
	if cfg.Log == nil {
		cfg.Log = logs.DefaultWrapper()
	}
	if cfg.Processor.CpuDurationSeconds > cfg.Processor.PeriodSeconds {
		cfg.Processor.CpuDurationSeconds = cfg.Processor.PeriodSeconds
	}
	if cfg.Processor.MutexProfileFraction < 0 {
		cfg.Processor.MutexProfileFraction = 0
	}
}

// Watch 观察配置变化
func (p *processor) Watch(readOnlyConfig *ocp.GalileoConfig) {
	p.UpdateConfig(
		profiles.NewConfig(
			&p.cfg.Resource,
			profiles.WithProcessor(&readOnlyConfig.Config.ProfilesConfig.Processor),
			profiles.WithExporter(&readOnlyConfig.Config.ProfilesConfig.Exporter),
		),
	)
}

// UpdateConfig 更新配置。
func (p *processor) UpdateConfig(cfg *configs.Profiles) {
	configMutex.Lock()
	defer configMutex.Unlock()
	if sameAsYaml(p.cfg.Processor, cfg.Processor) && sameAsYaml(p.cfg.Exporter, cfg.Exporter) {
		return
	}
	p.stop()
	cfg.Log.Infof("[galileo]profiles processor.UpdateConfig|cfg=%+v\n", cfg)
	fixConfig(cfg)
	cfg.Log.Infof("[galileo]profiles processor.UpdateConfig|fix.cfg=%+v\n", cfg)
	p.exporter.UpdateConfig(cfg)
	p.reloadConfig(cfg)
	p.run()
}

func (p *processor) reloadConfig(cfg *configs.Profiles) {
	p.cfg = cfg
	p.stats = cfg.Stats
	p.stopCh = make(chan struct{})
	p.resetProfileTypes(cfg)

}

func (p *processor) resetProfileTypes(cfg *configs.Profiles) {
	for k := range p.enabledProfiles {
		delete(p.enabledProfiles, k)
	}
	for k := range p.deltas {
		delete(p.deltas, k)
	}
	p.addProfileTypes(cfg.Processor.ProfileTypes)
}

func sameAsYaml(old, new interface{}) bool {
	return toYaml(old) == toYaml(new)
}

func toYaml(in interface{}) string {
	out, err := yaml.Marshal(in)
	if err != nil {
		return ""
	}
	return string(out)
}
