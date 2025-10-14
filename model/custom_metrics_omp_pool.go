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

var customMetricsPool sync.Pool

// GetCustomMetrics 从对象池获取 *CustomMetrics，标签长度 labelCount。
func GetCustomMetrics(labelCount, pointCount int) *CustomMetrics {
	if c, ok := customMetricsPool.Get().(*CustomMetrics); ok {
		if cap(c.Metrics) < pointCount {
			c.Metrics = make([]Metric, pointCount)
		}
		if cap(c.CustomLabels) < labelCount {
			c.CustomLabels = make([]Label, labelCount)
		}
		c.Metrics = c.Metrics[:pointCount]
		c.CustomLabels = c.CustomLabels[:labelCount]
		return c
	}
	return NewCustomMetrics(labelCount, pointCount)
}

// PutCustomMetrics 把 *CustomMetrics 放回对象池。
func PutCustomMetrics(c *CustomMetrics) {
	customMetricsPool.Put(c)
}
