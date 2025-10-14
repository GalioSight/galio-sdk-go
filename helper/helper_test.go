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

package helper

import (
	"testing"

	"galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/model"
)

func TestGetMetricsProcessor(t *testing.T) {
	type args struct {
		cfg *configs.Metrics
	}
	tests := []struct {
		args    args
		name    string
		wantErr bool
	}{
		{
			name: "获取监控处理器成功",
			args: args{
				cfg: &configs.Metrics{
					Processor: model.MetricsProcessor{
						Protocol:       "omp",
						WindowSeconds:  10,
						ClearSeconds:   100,
						ExpiresSeconds: 100,
					},
					Exporter: model.MetricsExporter{
						Protocol: "otp",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "获取监控处理器失败 (导出器失败)",
			args: args{
				cfg: &configs.Metrics{
					Processor: model.MetricsProcessor{
						Protocol:       "omp",
						WindowSeconds:  10,
						ClearSeconds:   100,
						ExpiresSeconds: 100,
					},
					Exporter: model.MetricsExporter{
						Protocol: "otp1",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "获取监控处理器失败 (处理器失败)",
			args: args{
				cfg: &configs.Metrics{
					Processor: model.MetricsProcessor{
						Protocol: "omp1",
					},
					Exporter: model.MetricsExporter{
						Protocol: "otp",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				_, err := GetMetricsProcessor(tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetMetricsProcessor() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			},
		)
	}
}
