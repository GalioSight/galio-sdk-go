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

package model

import (
	"sync"
)

var serverMetricsPool sync.Pool

// GetServerMetrics 从对象池获取 *ServerMetrics，标签长度 labelCount。
func GetServerMetrics(labelCount int) *ServerMetrics {
	if s, ok := serverMetricsPool.Get().(*ServerMetrics); ok {
		s.RpcLabels.grow(labelCount)
		return s
	}
	return NewServerMetrics(labelCount)
}

// PutServerMetrics 把 *ServerMetrics 放回对象池。
func PutServerMetrics(s *ServerMetrics) {
	for i := range s.RpcLabels.Fields {
		s.RpcLabels.Fields[i].Name = RPCLabels_max_field
		s.RpcLabels.Fields[i].Value = ""
	}
	s.RpcLabels.Fields = s.RpcLabels.Fields[:0]
	for i := range s.Metrics {
		s.Metrics[i].Value = 0
	}
	serverMetricsPool.Put(s)
}
