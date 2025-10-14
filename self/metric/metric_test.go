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

package metric

import (
	"testing"
	"time"

	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	type args struct {
		resource *model.Resource
		monitor  model.SelfMonitor
		log      *logs.Wrapper
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"t1",
			args{
				resource: &model.Resource{
					Target:        "PCG-123.example.greeter",
					Namespace:     "Development",
					EnvName:       "test",
					Region:        "",
					Instance:      "127.0.0.1",
					Node:          "",
					ContainerName: "sz1",
					Version:       "v0.0.1",
					Platform:      "PCG-123",
					ObjectName:    "example.greeter",
					App:           "example",
					Server:        "greeter",
					SetName:       "",
					FrameCode:     "trpc",
					ServiceName:   "trpc.example.greeter.Unknown",
					TenantId:      "galileo",
				},
				monitor: model.SelfMonitor{
					Protocol: "otp", Collector: model.Collector{
						Addr: "", TelemetryData: 1,
						DataProtocol: 5, DataTransmission: 1, Version: 0,
					},
					ReportSeconds: 1,
				},
				log: logs.DefaultWrapper(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				Init(tt.args.resource, tt.args.monitor, tt.args.log)
				time.Sleep(1500 * time.Millisecond)
				stats := GetSelfMonitor().Stats
				assert.Equal(t, int64(1), stats.SelfMonitorCount.Load())
				assert.Equal(t, int64(0), stats.SelfMonitorError.Load())
			},
		)
	}
}
