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
	"testing"

	"galiosight.ai/galio-sdk-go/model"
	"github.com/stretchr/testify/assert"
)

func Test_convertName(t *testing.T) {
	type args struct {
		c *model.CustomMetrics
	}
	tests := []struct {
		args args
		want *model.CustomMetrics
		name string
	}{
		{
			name: "test测试",
			args: args{
				c: &model.CustomMetrics{
					MonitorName:  "test测试",
					CustomLabels: []model.Label{{Name: "type", Value: "app"}, {Name: "城市", Value: "深圳"}},
					Metrics: []model.Metric{
						{Name: "sum", Value: 100, Aggregation: model.Aggregation_AGGREGATION_SUM},
						{Name: "最新值", Value: 100, Aggregation: model.Aggregation_AGGREGATION_SET},
						{Name: "avg", Value: 100, Aggregation: model.Aggregation_AGGREGATION_AVG},
						{Name: "histogram", Value: 200, Aggregation: model.Aggregation_AGGREGATION_HISTOGRAM},
					},
				},
			},
			want: &model.CustomMetrics{
				Metrics: []model.Metric{
					{
						Name: "custom_counter_base62VeV0XsWzpbuyoD_sum_total", Value: 100, Aggregation: 2,
					}, {
						Name: "custom_gauge_base62VeV0XsWzpbuyoD_base628CY5wap5Ayp5_set", Value: 100,
						Aggregation: 1,
					},
					{Name: "custom_counter_base62VeV0XsWzpbuyoD_avg", Value: 100, Aggregation: 3},
					{
						Name: "custom_histogram_base62VeV0XsWzpbuyoD_histogram", Value: 200, Aggregation: 6,
					},
				}, CustomLabels: []model.Label{
					model.Label{Name: "type", Value: "app"},
					{Name: "base62Cib5OezyB", Value: "深圳"},
				},
				MonitorName: "test测试",
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				convertName(tt.args.c)
				assert.Equalf(t, tt.want, tt.args.c, "convertName(%v)", tt.args.c)
			},
		)
	}
}
