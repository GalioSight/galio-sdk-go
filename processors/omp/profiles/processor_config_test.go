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
	"runtime"
	"testing"

	"galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/configs/ocp"
	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/protocols"
	"github.com/stretchr/testify/assert"
)

func Test_profilesProcessor_UpdateConfig(t *testing.T) {
	oldCfg := &configs.Profiles{
		Log: logs.NopWrapper(),
		Resource: model.Resource{
			Target: "PCG-123.xxx.yyy",
		},
		Processor: model.ProfilesProcessor{
			Protocol:             protocols.OMP,
			ProfileTypes:         []string{"cpu"},
			PeriodSeconds:        1,
			CpuDurationSeconds:   1,
			CpuProfileRate:       100,
			MutexProfileFraction: 10,
			BlockProfileRate:     100000000,
			EnableDeltaProfiles:  true,
		},
	}
	newCfg := &configs.Profiles{
		Log: logs.NopWrapper(),
		Resource: model.Resource{
			Target: "PCG-123.xxx.yyy",
		},
		Processor: model.ProfilesProcessor{
			Protocol:             protocols.OMP,
			ProfileTypes:         []string{"mutex"},
			PeriodSeconds:        2,
			CpuDurationSeconds:   10,
			CpuProfileRate:       500,
			MutexProfileFraction: 10,
			BlockProfileRate:     100000000,
			EnableDeltaProfiles:  true,
		},
	}
	// 关闭 mutex profile
	start := runtime.SetMutexProfileFraction(0)
	defer runtime.SetMutexProfileFraction(start)
	exporter := newFakeExporter()
	processor := newProcessor(oldCfg, exporter)
	processor.run()
	processor.Watch(
		&ocp.GalileoConfig{
			Config: model.GetConfigResponse{
				ProfilesConfig: model.ProfilesConfig{
					Enable:    newCfg.Enable,
					Processor: newCfg.Processor,
					Exporter:  newCfg.Exporter,
				},
			},
			Resource: newCfg.Resource,
		},
	)
	processor.UpdateConfig(newCfg)
	assert.Equal(t, processor.cfg.Processor.ProfileTypes, []string{"mutex"})
	assert.Equal(t, len(processor.enabledProfiles), 1)
	assert.Equal(t, len(processor.deltas), 1)
	assert.Equal(t, processor.enabledProfiles["mutex"], true)
	assert.Equal(t, int64(2), processor.cfg.Processor.PeriodSeconds)
	assert.Equal(t, int64(2), processor.cfg.Processor.CpuDurationSeconds)
	assert.Equal(t, int32(500), processor.cfg.Processor.CpuProfileRate)
	assert.Equal(t, 10, runtime.SetMutexProfileFraction(-1))
	processor.stop()
}
