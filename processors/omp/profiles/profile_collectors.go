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
	"errors"
	"fmt"
	"runtime"
	"runtime/pprof"
	"time"

	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/processors/omp/profiles/delta"
)

// profileCollector model.ProfileType 对应实现
type profileCollector struct {
	// Type profile 类型
	Type model.ProfileType
	// Name profile 类型的名称
	Name string
	// Filename 生成的 profile 文件名
	Filename string
	// Collect 不同类型 profile 数据的收集函数
	Collect func(p *processor) ([]byte, error)
	// DeltaValues 指定当启用增量 profiles 时，哪些 samples 里的值需要修改。
	// 如果 DeltaValues 为空，则表示该 profile type 不支持增量采集。
	DeltaValues []delta.DeltaValueType
}

func getProfileCollector(t model.ProfileType) profileCollector {
	c, ok := profileCollectors[t]
	if ok {
		c.Type = t
		return c
	}
	return profileCollector{
		Type:     t,
		Name:     "unknown",
		Filename: "unknown",
		Collect: func(_ *processor) ([]byte, error) {
			return nil, errors.New("profile type not implemented")
		},
	}
}

// profileTypes 定义每种 profile 类型采集的实现
var profileCollectors = map[model.ProfileType]profileCollector{
	model.CPUProfile: {
		Name:     "cpu",
		Filename: "cpu.pprof",
		Collect: func(p *processor) ([]byte, error) {
			// runtime.StartCPUProfile 中默认 cpu profile rate 为 100 hz
			const hz = 100
			var buf bytes.Buffer
			// 在上一个采集周期结束后，启动 CPU Profiling
			p.interruptibleSleep(time.Duration(p.cfg.Processor.PeriodSeconds-p.cfg.Processor.CpuDurationSeconds) * time.Second)
			// 如果 CpuProfileRate 为默认值（100 Hz），则不需要再设置 cpu profile rate
			if p.cfg.Processor.CpuProfileRate != 0 && p.cfg.Processor.CpuProfileRate != hz {
				// 需在 pprof.StartCPUProfile 前调用才能生效
				runtime.SetCPUProfileRate(int(p.cfg.Processor.CpuProfileRate))
			}

			if err := pprof.StartCPUProfile(&buf); err != nil {
				return nil, err
			}
			p.interruptibleSleep(time.Duration(p.cfg.Processor.CpuDurationSeconds) * time.Second)

			// 通过 pendingProfiles.Wait() 等待其他类型 profiles 完成收集,
			// CPU profiles 最后完成收集, 以保证收集完整的 profiles 数据
			p.pendingProfiles.Wait()
			pprof.StopCPUProfile()
			return buf.Bytes(), nil
		},
	},
	// HeapProfile 包含 4 种 sample:
	// alloc_objects/count, alloc_space/bytes, inuse_objects/count, inuse_space/bytes.
	// alloc_objects/count 和 alloc_space/bytes 是整个进程生命周期内的内存分配, 所以进行增量收集
	// inuse_objects/count, inuse_space/bytes 是当前堆状态的 snapshots
	model.HeapProfile: {
		Name:     "heap",
		Filename: "heap.pprof",
		Collect:  collectGenericProfile("heap", model.HeapProfile),
		DeltaValues: []delta.DeltaValueType{
			{Type: "alloc_objects", Unit: "count"},
			{Type: "alloc_space", Unit: "bytes"},
		},
	},
	model.MutexProfile: {
		Name:     "mutex",
		Filename: "mutex.pprof",
		Collect:  collectGenericProfile("mutex", model.MutexProfile),
		DeltaValues: []delta.DeltaValueType{
			{Type: "contentions", Unit: "count"},
			{Type: "delay", Unit: "nanoseconds"},
		},
	},
	model.BlockProfile: {
		Name:     "block",
		Filename: "block.pprof",
		Collect:  collectGenericProfile("block", model.BlockProfile),
		DeltaValues: []delta.DeltaValueType{
			{Type: "contentions", Unit: "count"},
			{Type: "delay", Unit: "nanoseconds"},
		},
	},
	model.GoroutineProfile: {
		Name:     "goroutine",
		Filename: "goroutine.pprof",
		Collect:  collectGenericProfile("goroutine", model.GoroutineProfile),
	},
}

func collectGenericProfile(name string, pt model.ProfileType) func(p *processor) ([]byte, error) {
	return func(p *processor) ([]byte, error) {
		p.interruptibleSleep(time.Duration(p.cfg.Processor.PeriodSeconds) * time.Second)

		var buf bytes.Buffer
		err := p.lookupProfile(name, &buf, 0)
		data := buf.Bytes()
		dp, ok := p.deltas[pt]
		if !ok || !p.cfg.Processor.EnableDeltaProfiles {
			return data, err
		}
		delta, err := dp.Delta(data)
		if err != nil {
			return nil, fmt.Errorf("delta profile error: %s", err)
		}
		return delta, err
	}
}
