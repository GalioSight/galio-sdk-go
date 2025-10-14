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
	"sync"

	"galiosight.ai/galio-sdk-go/lib/times"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/processors/omp/metrics/point"
)

// multi 多值点。
type multi struct {
	monitorName  string
	toOTPFunc    multiToOTPFunc
	rpcLabels    *model.RPCLabels
	customLabels []*model.Label
	points       []*point.Point
	pk           PK
	updateUnix   int64
	mu           sync.Mutex // 保护 points 并发读写。
}

var (
	rpcMultiPool    sync.Pool
	customMultiPool sync.Pool
	otherMultiPool  sync.Pool
)

func getMulti(p *PK) *multi {
	var m *multi
	if p.IsRPC() {
		m, _ = rpcMultiPool.Get().(*multi)
	} else if p.IsCustom() {
		m, _ = customMultiPool.Get().(*multi)
	} else {
		m, _ = otherMultiPool.Get().(*multi)
	}
	if m != nil {
		return m
	}
	return &multi{}
}

func putMulti(m *multi) {
	g := m.pk.group
	m.reset()
	put(g, m)
}

func put(g model.MetricGroup, m *multi) {
	switch g {
	case model.ClientGroup, model.ServerGroup:
		rpcMultiPool.Put(m)
	case model.CustomGroup:
		customMultiPool.Put(m)
	default:
		otherMultiPool.Put(m)
	}
}

func (m *multi) reset() {
	m.resetRPCLabels()
	m.resetCustomLabels()
	m.toOTPFunc = nil
	m.resetPoints()
	m.pk.reset()
	m.updateUnix = 0
	m.monitorName = ""
}

func (m *multi) resetRPCLabels() {
	if m.rpcLabels == nil {
		return
	}
	for i := range m.rpcLabels.Fields {
		m.rpcLabels.Fields[i].Name = 0
		m.rpcLabels.Fields[i].Value = ""
	}
	m.rpcLabels.Fields = m.rpcLabels.Fields[:0]
}

func (m *multi) resetCustomLabels() {
	if len(m.customLabels) == 0 {
		return
	}
	for i := range m.customLabels {
		if m.customLabels[i] == nil {
			continue
		}
		m.customLabels[i].Name = ""
		m.customLabels[i].Value = ""
	}
	m.customLabels = m.customLabels[:0]
}

func (m *multi) resetPoints() {
	for i := range m.points {
		point.Put(m.points[i])
		m.points[i] = nil
	}
	m.points = m.points[:0]
}

func (m *multi) setPoints(extractor model.OMPMetric, getBucket func(name string) point.BucketFunc) {
	count := extractor.PointCount()
	if cap(m.points) < count {
		m.points = make([]*point.Point, 0, count)
	}
	m.points = m.points[:count]
	for i := range m.points {
		name := extractor.PointName(i)
		aggregation := extractor.PointAggregation(i)
		m.points[i] = point.Get(aggregation, name)
		if aggregation == model.Aggregation_AGGREGATION_HISTOGRAM {
			m.points[i].SetBucket(getBucket(name))
		}
	}
}

func (m *multi) setRPCLabels(labels *model.RPCLabels) {
	if labels == nil || len(labels.Fields) == 0 {
		return
	}
	if m.rpcLabels == nil {
		m.rpcLabels = &model.RPCLabels{}
	}
	fields := labels.Fields
	if cap(m.rpcLabels.Fields) < len(fields) {
		m.rpcLabels.Fields = make([]model.RPCLabels_Field, 0, len(fields))
	}
	m.rpcLabels.Fields = m.rpcLabels.Fields[:len(fields)]
	for i := range fields {
		m.rpcLabels.Fields[i].Name = fields[i].Name
		m.rpcLabels.Fields[i].Value = fields[i].Value
	}
}

func (m *multi) setCustomLabels(labels []model.Label) {
	if len(labels) == 0 {
		return
	}
	if cap(m.customLabels) < len(labels) {
		m.customLabels = make([]*model.Label, 0, len(labels))
	}
	m.customLabels = m.customLabels[:len(labels)]
	for i := range labels {
		if m.customLabels[i] == nil {
			m.customLabels[i] = &model.Label{}
		}
		m.customLabels[i].Name = labels[i].Name
		m.customLabels[i].Value = labels[i].Value
	}
}

func (m *multi) update(extractor model.OMPMetric) {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := extractor.PointCount()
	if count != len(m.points) { // 点数不一致，不更新。
		return
	}
	for i := range m.points {
		if m.points[i] == nil {
			return
		}
		if m.points[i].Aggregation() != extractor.PointAggregation(i) { // 点类型不一致，不更新。
			return
		}
	}
	for i := range m.points {
		m.points[i].Update(extractor.PointValue(i))
	}
	m.updateUnix = times.SecondPrecisionUnix() // 15 ns/op
}

func (m *multi) change(factor float64) {
	for i := range m.points {
		if err := m.points[i].Change(factor); err != nil {
			return
		}
	}
}
