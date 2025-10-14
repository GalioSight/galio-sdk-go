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

package delta

import (
	"bytes"
	"fmt"

	pprofile "github.com/google/pprof/profile"
)

// NewSimpleProfiler 创建新的 SimpleProfiler
func NewSimpleProfiler(sampleTypes []DeltaValueType) Profiler {
	return &SimpleProfiler{
		sampleTypes: sampleTypes,
	}
}

// SimpleProfiler 简单的 delta profiler
// 主要使用 go 的 profile Merge 方法进行 delta 计算
type SimpleProfiler struct {
	prevProf    *pprofile.Profile
	sampleTypes []DeltaValueType
}

// Delta 实现 profile 数据的增量计算
func (p *SimpleProfiler) Delta(data []byte) ([]byte, error) {
	curProf, err := pprofile.ParseData(data)
	if err != nil {
		return nil, fmt.Errorf("delta prof parse: %v", err)
	}
	prevProf := p.prevProf
	var deltaData []byte
	if prevProf == nil {
		deltaData = data
	} else {
		deltaProf, err := p.CalculateDeltas(prevProf, curProf)
		if err != nil {
			return nil, fmt.Errorf("delta prof merge: %v", err)
		}
		deltaProf.TimeNanos = curProf.TimeNanos
		deltaProf.DurationNanos = curProf.TimeNanos - prevProf.TimeNanos
		deltaBuf := &bytes.Buffer{}
		if err := deltaProf.Write(deltaBuf); err != nil {
			return nil, fmt.Errorf("delta prof write: %v", err)
		}
		deltaData = deltaBuf.Bytes()
	}
	p.prevProf = curProf
	return deltaData, nil
}

// CalculateDeltas 计算 new - old 的 delta profile，逻辑参考
// https://github.com/golang/go/blob/release-branch.go1.17/src/net/http/pprof/pprof.go#L276-L318
func (p *SimpleProfiler) CalculateDeltas(old, new *pprofile.Profile) (*pprofile.Profile, error) {
	ratios := make([]float64, len(old.SampleType))
	// 我们只计算 profileTypes 中定义的 DeltaValues，所以将对应位置的 ration 设置为 -1,
	// 其他不需要计算的 ration 为 0（float64 默认值）。
	for _, dst := range p.sampleTypes {
		var found bool
		for i, st := range old.SampleType {
			if dst.Type == st.Type && dst.Unit == st.Unit {
				ratios[i] = -1
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("sample type %s not found in the profile", dst.Type)
		}
	}
	if err := old.ScaleN(ratios); err != nil {
		return nil, err
	}
	delta, err := pprofile.Merge([]*pprofile.Profile{old, new})
	if err != nil {
		return nil, err
	}
	return delta, delta.CheckValid()
}
