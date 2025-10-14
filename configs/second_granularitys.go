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

package configs

import (
	"time"

	"galiosight.ai/galio-sdk-go/lib/times"
	"galiosight.ai/galio-sdk-go/model"
)

// Range 时间 range 配置。
type Range struct {
	Begin  int64         // 开始时间，时间戳方便快速对比。
	End    int64         // 结束时间。
	Window time.Duration // 秒级监控聚合窗口大小，如 1s 5s 10s。
}

// SecondGranularitys 秒级监控配置。
type SecondGranularitys struct {
	GroupToMonitors [model.MaxGroup]map[string]*Range // group -> monitor -> range{begin, end}
}

// NewSecondGranularitys 新增秒级监控配置。
func NewSecondGranularitys() *SecondGranularitys {
	return &SecondGranularitys{}
}

var monitorNameToGroups map[string]model.MetricGroup = map[string]model.MetricGroup{
	model.RPCClient:      model.ClientGroup,
	model.RPCServer:      model.ServerGroup,
	model.NormalProperty: model.NormalGroup,
}

// ConvSecondGranularitys 转换秒级监控配置到 map，方便使用。
func (m *Metrics) ConvSecondGranularitys() {
	m.SecondGranularitys = NewSecondGranularitys()
	secondGranularitys := m.Processor.SecondGranularitys
	for _, s := range secondGranularitys {
		monitorName := s.MonitorName
		group, ok := monitorNameToGroups[monitorName]
		if ok {
			m.SecondGranularitys.Add(
				group, monitorName, s.BeginSecond, s.EndSecond, secondsToDuration(s.WindowSeconds),
			)
		} else {
			m.SecondGranularitys.Add(
				model.CustomGroup, monitorName, s.BeginSecond, s.EndSecond, secondsToDuration(s.WindowSeconds),
			)
		}
	}
}

func secondsToDuration(seconds int32) time.Duration {
	return time.Duration(seconds) * time.Second
}

// Clone clone 一份秒级监控配置。
func (s *SecondGranularitys) Clone() *SecondGranularitys {
	secondGranularitys := NewSecondGranularitys()
	if s == nil {
		return secondGranularitys
	}
	for i := range s.GroupToMonitors {
		for k, v := range s.GroupToMonitors[i] {
			secondGranularitys.Add(model.MetricGroup(i), k, v.Begin, v.End, v.Window)
		}
	}
	return secondGranularitys
}

// Add 增加配置。
func (s *SecondGranularitys) Add(group model.MetricGroup, monitor string, begin, end int64, window time.Duration) {
	if s == nil { // 未初始化。
		return
	}
	g := int(group)
	if g < 0 || g >= len(s.GroupToMonitors) { // 非法组。
		return
	}
	if s.GroupToMonitors[g] == nil {
		s.GroupToMonitors[g] = make(map[string]*Range)
	}
	s.GroupToMonitors[g][monitor] = &Range{Begin: begin, End: end, Window: window}
}

// Enabled 是否开启秒级监控。
func (s *SecondGranularitys) Enabled(group model.MetricGroup, monitor string) (bool, time.Duration) {
	if s == nil { // 未初始化。
		return false, 0
	}
	g := int(group)
	if g < 0 || g >= len(s.GroupToMonitors) { // 非法组。
		return false, 0
	}
	monitors := s.GroupToMonitors[g]
	if len(monitors) == 0 { // 组无秒级监控配置。
		return false, 0
	}
	r, ok := monitors[monitor]
	if !ok || r == nil { // 监控项无秒级监控配置。
		return false, 0
	}
	now := times.SecondPrecisionUnix()
	return now >= r.Begin && now <= r.End, r.Window
}
