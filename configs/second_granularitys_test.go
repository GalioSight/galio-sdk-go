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
	"testing"
	"time"

	"galiosight.ai/galio-sdk-go/model"
	"github.com/stretchr/testify/require"
)

func TestSecondGranularitys_Enabled(t *testing.T) {
	now := time.Now().Unix()
	tests := []struct {
		cfgFunc func() *Metrics
		group   model.MetricGroup
		monitor string
		want    bool
		window  time.Duration
	}{
		{
			cfgFunc: func() *Metrics { return &Metrics{} },
			group:   0,
			monitor: "",
			want:    false,
			window:  0,
		},
		{
			cfgFunc: func() *Metrics {
				cfg := &Metrics{
					Processor: model.MetricsProcessor{
						SecondGranularitys: []model.SecondGranularity{
							{
								MonitorName: "自定义秒级监控项 1", BeginSecond: now - 60, EndSecond: now + 60,
								WindowSeconds: 10,
							},
						},
					},
				}
				return cfg
			},
			group:   model.CustomGroup,
			monitor: "自定义秒级监控项 1",
			want:    true,
			window:  time.Second * 10,
		},
		{
			cfgFunc: func() *Metrics {
				cfg := &Metrics{
					Processor: model.MetricsProcessor{
						SecondGranularitys: []model.SecondGranularity{
							{MonitorName: "自定义秒级监控项 2", BeginSecond: now + 60, EndSecond: now + 120},
						},
					},
				}
				return cfg
			},
			group:   model.CustomGroup,
			monitor: "自定义秒级监控项 2",
			want:    false,
			window:  0,
		},
		{
			cfgFunc: func() *Metrics {
				cfg := &Metrics{
					Processor: model.MetricsProcessor{
						SecondGranularitys: []model.SecondGranularity{
							{MonitorName: "自定义秒级监控项 3", BeginSecond: now - 60, EndSecond: now + 60},
						},
					},
				}
				return cfg
			},
			group:   model.CustomGroup,
			monitor: "自定义秒级监控项 33",
			want:    false,
			window:  0,
		},
		{
			cfgFunc: func() *Metrics {
				cfg := &Metrics{
					Processor: model.MetricsProcessor{
						SecondGranularitys: []model.SecondGranularity{
							{
								MonitorName:   model.RPCClient,
								BeginSecond:   now - 60,
								EndSecond:     now + 60,
								WindowSeconds: 1,
							},
						},
					},
				}
				return cfg
			},
			group:   model.ClientGroup,
			monitor: model.RPCClient,
			want:    true,
			window:  time.Second,
		},
	}
	for i, tt := range tests {
		cfg := tt.cfgFunc()
		cfg.ConvSecondGranularitys()
		ok, window := cfg.SecondGranularitys.Enabled(tt.group, tt.monitor)
		require.Equalf(t, tt.want, ok, "Case %d", i)
		require.Equalf(t, tt.window, window, "Case %d", i)
	}
}
