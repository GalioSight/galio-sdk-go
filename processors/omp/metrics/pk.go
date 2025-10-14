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

	"galiosight.ai/galio-sdk-go/model"
)

// PK 监控项（主调、被调、属性、自定义）唯一主键。
// 如：
// ip1 请求 ip2 A 接口的主调，主键是 pk1。
// ip3 请求 ip2 A 接口的主调，主键是 pk2。
type PK struct {
	group          model.MetricGroup // 0：主调 1：被调 2：属性 3：自定义。
	labelsHashCode uint64            // 所有 label value 的 hash。
	labelsBytes    int               // 所有 label value 的总字节数。
	pointCount     int               // 数据点数（单值点 1 个，多值点大于 1）。
	namesHashCode  uint64            // 所有数据点 name 的 hash。
	namesBytes     int               // 所有数据点 name 的总字节数。
}

func newPK() *PK {
	return &PK{}
}

func (p *PK) reset() {
	p.group = 0
	p.labelsHashCode = 0
	p.labelsBytes = 0
	p.pointCount = 0
	p.namesHashCode = 0
	p.namesBytes = 0
}

func (p *PK) copyTo(to *PK) {
	to.group = p.group
	to.labelsHashCode = p.labelsHashCode
	to.labelsBytes = p.labelsBytes
	to.pointCount = p.pointCount
	to.namesHashCode = p.namesHashCode
	to.namesBytes = p.namesBytes
}

func (p *PK) set(extractor model.OMPMetric) {
	p.group = extractor.Group()
	p.pointCount = extractor.PointCount()
	p.labelsHashCode, p.labelsBytes = hashLabels(extractor)
	if p.group == model.ClientGroup || p.group == model.ServerGroup { // 主调、被调 name 都是固定的，不需要计算。
		p.namesBytes = 0
		p.namesHashCode = 0
		return
	}
	p.namesHashCode, p.namesBytes = hashNames(extractor)
}

var pkPool sync.Pool

func getPK() *PK {
	if p, ok := pkPool.Get().(*PK); ok {
		return p
	}
	return newPK()
}

func putPK(p *PK) {
	p.reset()
	pkPool.Put(p)
}

// IsRPC 是否模调。
func (p *PK) IsRPC() bool {
	return p.group == model.ClientGroup || p.group == model.ServerGroup
}

// IsCustom 是否自定义。
func (p *PK) IsCustom() bool {
	return p.group == model.CustomGroup
}
