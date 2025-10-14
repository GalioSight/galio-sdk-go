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

	"galiosight.ai/galio-sdk-go/model"
	"github.com/stretchr/testify/assert"
)

func TestMetrics_convLabels(t *testing.T) {
	type args struct {
		monitorName string
		labelName   string
		isIgnore    bool
	}
	tests := []struct {
		name         string
		labelIgnores []model.LabelIgnore
		want         *IgnoreLabels
		args         []args
	}{
		{
			"empty",
			nil,
			&IgnoreLabels{},
			[]args{
				{
					"abc",
					"a",
					false,
				},
			},
		},
		{
			"rpc_client",
			[]model.LabelIgnore{
				{
					MonitorName: "rpc_client",
					LabelNames: []string{
						"caller_ip",
					},
				},
			},
			&IgnoreLabels{
				monitorNameToRPCLabels: map[string]map[model.RPCLabels_FieldName]bool{
					"rpc_client": {
						model.RPCLabels_caller_ip: true,
					},
				},
			},
			[]args{
				{
					"rpc_client",
					"a",
					false,
				},
				{
					"rpc_client",
					"caller_ip",
					true,
				},
			},
		},
		{
			"rpc_server",
			[]model.LabelIgnore{
				{
					MonitorName: "rpc_server",
					LabelNames: []string{
						"caller_container",
					},
				},
			},
			&IgnoreLabels{
				monitorNameToRPCLabels: map[string]map[model.RPCLabels_FieldName]bool{
					"rpc_server": {
						model.RPCLabels_caller_container: true,
					},
				},
			},
			[]args{
				{
					"rpc_server",
					"a",
					false,
				},
				{
					"rpc_server",
					"caller_container",
					true,
				},
			},
		},
		{
			"custom",
			[]model.LabelIgnore{
				{
					MonitorName: "custom",
					LabelNames: []string{
						"demo",
					},
				},
				{
					MonitorName: "中文监控项",
					LabelNames: []string{
						"城市",
					},
				},
			},
			&IgnoreLabels{
				monitorNameToCustomLabels: map[string]map[string]bool{
					"custom": {
						"demo": true,
					},
					"中文监控项": {
						"城市": true,
					},
				},
			},
			[]args{
				{
					"custom",
					"demo",
					true,
				},
				{
					"custom",
					"abc",
					false,
				},
				{
					"中文监控项",
					"城市",
					true,
				},
			},
		},
	}
	for _, tt := range tests {
		for _, v := range tt.args {
			t.Run(
				tt.name+"/"+v.monitorName+"/"+v.labelName, func(t *testing.T) {
					m := &Metrics{
						Processor: model.MetricsProcessor{
							LabelIgnores: tt.labelIgnores,
						},
					}
					m.ConvIgnoreLabels()
					assert.Equal(t, tt.want, m.GetIgnoreLabels())

					switch v.monitorName {
					case model.RPCClient, model.RPCServer:
						label, err := convRPCLabel(v.labelName)
						if err != nil {
							return
						}
						isIgnore := m.GetIgnoreLabels().IgnoreRPCLabel(v.monitorName, label)
						assert.Equal(t, v.isIgnore, isIgnore)
					case model.GoRuntime, model.Process:
						// 无维度，不需要处理
					default:
						// 其他均是自定义监控
						isIgnore := m.GetIgnoreLabels().IgnoreCustomLabel(v.monitorName, v.labelName)
						assert.Equal(t, v.isIgnore, isIgnore)
					}
				},
			)
		}
	}
}

func TestIgnoreLabels_IgnoreCustomLabels(t *testing.T) {
	type args struct {
		ignoreLabels *IgnoreLabels
		input        *model.CustomMetrics
		want         *model.CustomMetrics
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"custom",
			args{
				&IgnoreLabels{
					monitorNameToCustomLabels: map[string]map[string]bool{
						"custom": {
							"demo": true,
						},
					},
				},
				&model.CustomMetrics{
					CustomLabels: []model.Label{
						{
							Name:  "demo",
							Value: "test",
						},
						{
							Name:  "demo2",
							Value: "test",
						},
					},
					MonitorName: "custom",
				},
				&model.CustomMetrics{
					CustomLabels: []model.Label{
						{
							Name:  "demo",
							Value: "ignore",
						},
						{
							Name:  "demo2",
							Value: "test",
						},
					},
					MonitorName: "custom",
				},
			},
		},
		{
			"custom 中文",
			args{
				&IgnoreLabels{
					monitorNameToCustomLabels: map[string]map[string]bool{
						"中文监控项": {
							"城市": true,
						},
					},
				},
				&model.CustomMetrics{
					CustomLabels: []model.Label{
						{
							Name:  "城市",
							Value: "test",
						},
						{
							Name:  "demo2",
							Value: "test",
						},
					},
					MonitorName: "中文监控项",
				},
				&model.CustomMetrics{
					CustomLabels: []model.Label{
						{
							Name:  "城市",
							Value: "ignore",
						},
						{
							Name:  "demo2",
							Value: "test",
						},
					},
					MonitorName: "中文监控项",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.args.ignoreLabels.IgnoreCustomLabels(tt.args.input)
				assert.Equal(t, tt.args.want, tt.args.input)
			},
		)
	}
}

func TestIgnoreLabels_IgnoreClientLabels(t *testing.T) {
	type args struct {
		ignoreLabels *IgnoreLabels
		input        *model.ClientMetrics
		want         *model.ClientMetrics
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"rpc_client",
			args{
				&IgnoreLabels{
					monitorNameToRPCLabels: map[string]map[model.RPCLabels_FieldName]bool{
						"rpc_client": {
							model.RPCLabels_callee_ip: true,
						},
					},
				},
				&model.ClientMetrics{
					RpcLabels: model.RPCLabels{
						Fields: []model.RPCLabels_Field{
							{
								Name:  model.RPCLabels_callee_ip,
								Value: "test",
							},
						},
					},
				}, &model.ClientMetrics{
					RpcLabels: model.RPCLabels{
						Fields: []model.RPCLabels_Field{
							{
								Name:  model.RPCLabels_callee_ip,
								Value: "ignore",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.args.ignoreLabels.IgnoreClientLabels(tt.args.input)
				assert.Equal(t, tt.args.want, tt.args.input)
			},
		)
	}
}

func TestIgnoreLabels_IgnoreServerLabels(t *testing.T) {
	type args struct {
		ignoreLabels *IgnoreLabels
		input        *model.ServerMetrics
		want         *model.ServerMetrics
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"rpc_server",
			args{
				&IgnoreLabels{
					monitorNameToRPCLabels: map[string]map[model.RPCLabels_FieldName]bool{
						"rpc_server": {
							model.RPCLabels_callee_ip: true,
						},
					},
				},
				&model.ServerMetrics{
					RpcLabels: model.RPCLabels{
						Fields: []model.RPCLabels_Field{
							{
								Name:  model.RPCLabels_callee_ip,
								Value: "test",
							},
						},
					},
				}, &model.ServerMetrics{
					RpcLabels: model.RPCLabels{
						Fields: []model.RPCLabels_Field{
							{
								Name:  model.RPCLabels_callee_ip,
								Value: "ignore",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.args.ignoreLabels.IgnoreServerLabels(tt.args.input)
				assert.Equal(t, tt.args.want, tt.args.input)
			},
		)
	}
}

func TestMetrics_GetIgnoreLabels(t *testing.T) {
	m := Metrics{}
	assert.EqualValues(t, &IgnoreLabels{}, m.GetIgnoreLabels())
	m.ignoreLabels.Store(&IgnoreLabels{
		monitorNameToRPCLabels:    map[string]map[model.RPCLabels_FieldName]bool{},
		monitorNameToCustomLabels: map[string]map[string]bool{},
	})
	assert.EqualValues(t, &IgnoreLabels{
		monitorNameToRPCLabels:    map[string]map[model.RPCLabels_FieldName]bool{},
		monitorNameToCustomLabels: map[string]map[string]bool{},
	}, m.GetIgnoreLabels())
}
