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
	"bytes"
	"testing"
	"time"

	"galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/protocols"
	"github.com/stretchr/testify/assert"
)

type fakeExporter struct {
	data []*model.ProfilesBatch
}

func (f *fakeExporter) Run(_ chan struct{}) {
}

func (f *fakeExporter) Export(batch *model.ProfilesBatch) {
	f.data = append(f.data, batch)
}

func (f *fakeExporter) UpdateConfig(_ *configs.Profiles) {
}

func (f *fakeExporter) Shutdown() {
	f.data = nil
}

func (f *fakeExporter) GetProfileData() []*model.ProfilesBatch {
	return f.data
}

func newFakeExporter() *fakeExporter {
	return &fakeExporter{}
}

func newTestCfg(ptypes []string) *configs.Profiles {
	return &configs.Profiles{
		Log: logs.NopWrapper(),
		Resource: model.Resource{
			Target: "PCG-123.xxx.yyy",
		},
		Processor: model.ProfilesProcessor{
			Protocol:             protocols.OMP,
			ProfileTypes:         ptypes,
			PeriodSeconds:        1,
			CpuDurationSeconds:   1,
			CpuProfileRate:       100,
			MutexProfileFraction: 10,
			BlockProfileRate:     100000000,
			EnableDeltaProfiles:  true,
		},
	}
}

func TestRunCPUProfile(t *testing.T) {
	exporter := newFakeExporter()
	cfg := newTestCfg([]string{"cpu"})
	processor := newProcessor(cfg, exporter)
	processor.run()
	// 等待 2 个 PeriodSeconds，保证至少有一个 profile batch 生成
	time.Sleep(time.Second * 2 * time.Duration(processor.cfg.Processor.PeriodSeconds))
	processor.stop()
	assert.NotEmpty(t, exporter.data)
	batch := exporter.data[0]
	assert.NotEmpty(t, batch)
	profs := batch.Profiles
	assert.Equal(t, "cpu.pprof", profs[0].Name)
	assert.True(t, isGzipData(profs[0].Data))
}

func TestRunHeapProfile(t *testing.T) {
	exporter := newFakeExporter()
	cfg := newTestCfg([]string{"heap"})
	processor := newProcessor(cfg, exporter)
	processor.run()
	// 等待 2 个 PeriodSeconds，保证至少有一个 profile batch 生成
	time.Sleep(time.Second * 2 * time.Duration(processor.cfg.Processor.PeriodSeconds))
	processor.stop()
	assert.NotEmpty(t, exporter.data)
	batch := exporter.data[0]
	assert.NotEmpty(t, batch)
	profs := batch.Profiles
	assert.Equal(t, "delta-heap.pprof", profs[0].Name)
	assert.True(t, isGzipData(profs[0].Data))
}

func TestRunMutexProfile(t *testing.T) {
	exporter := newFakeExporter()
	cfg := newTestCfg([]string{"mutex"})
	processor := newProcessor(cfg, exporter)
	processor.run()
	// 等待 2 个 PeriodSeconds，保证至少有一个 profile batch 生成
	time.Sleep(time.Second * 2 * time.Duration(processor.cfg.Processor.PeriodSeconds))
	processor.stop()
	assert.NotEmpty(t, exporter.data)
	batch := exporter.data[0]
	assert.NotEmpty(t, batch)
	profs := batch.Profiles
	assert.Equal(t, "delta-mutex.pprof", profs[0].Name)
	assert.True(t, isGzipData(profs[0].Data))
}

func TestRunBlockProfile(t *testing.T) {
	exporter := newFakeExporter()
	cfg := newTestCfg([]string{"block"})
	processor := newProcessor(cfg, exporter)
	processor.run()
	// 等待 2 个 PeriodSeconds，保证至少有一个 profile batch 生成
	time.Sleep(time.Second * 2 * time.Duration(processor.cfg.Processor.PeriodSeconds))
	processor.stop()
	assert.NotEmpty(t, exporter.data)
	batch := exporter.data[0]
	assert.NotEmpty(t, batch)
	profs := batch.Profiles
	assert.Equal(t, "delta-block.pprof", profs[0].Name)
	assert.True(t, isGzipData(profs[0].Data))
}

func TestRunGoroutineProfile(t *testing.T) {
	exporter := newFakeExporter()
	cfg := newTestCfg([]string{"goroutine"})
	processor := newProcessor(cfg, exporter)
	processor.run()
	// 等待 2 个 PeriodSeconds，保证至少有一个 profile batch 生成
	time.Sleep(time.Second * 2 * time.Duration(processor.cfg.Processor.PeriodSeconds))
	processor.stop()
	assert.NotEmpty(t, exporter.data)
	batch := exporter.data[0]
	assert.NotEmpty(t, batch)
	profs := batch.Profiles
	assert.Equal(t, "goroutine.pprof", profs[0].Name)
	assert.True(t, isGzipData(profs[0].Data))
}

func isGzipData(data []byte) bool {
	return bytes.HasPrefix(data, []byte{0x1f, 0x8b})
}
