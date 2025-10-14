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
	"galiosight.ai/galio-sdk-go/model"
)

// convertMetricName 转换指标名，以符合 prometheus 和 OMP 规范。
func convertName(rec *model.CustomMetrics) *model.CustomMetrics {
	convertLabelName(rec)
	convertMetricName(rec)
	return rec
}

// convertMetricName 转换指标名，以符合 prometheus 和 OMP 规范。
func convertMetricName(rec *model.CustomMetrics) {
	for i := range rec.Metrics {
		m := &rec.Metrics[i]
		aggregation := m.Aggregation
		m.Name = model.CustomName(rec.MonitorName, m.Name, aggregation)
	}
}

// convertLabelName 转换 label 名，以符合 prometheus 和 OMP 规范。
func convertLabelName(rec *model.CustomMetrics) {
	for i := range rec.CustomLabels {
		v := &rec.CustomLabels[i]
		v.Name = model.NameToIdentifier(v.Name)
	}
}
