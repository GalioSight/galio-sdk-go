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
	"fmt"

	"galiosight.ai/galio-sdk-go/model"
)

const IgnoredValue = "ignore"

// IgnoreLabels 屏蔽的标签列表。
type IgnoreLabels struct {
	// monitorNameToRPCLabels key：屏蔽的监控项名称，value：屏蔽的主被调标签。
	monitorNameToRPCLabels map[string]map[model.RPCLabels_FieldName]bool
	// monitorNameToCustomLabels key：屏蔽的监控项名称，value：屏蔽的自定义标签。
	monitorNameToCustomLabels map[string]map[string]bool
}

// ConvIgnoreLabels 转换屏蔽配置到 map，方便使用。
func (m *Metrics) ConvIgnoreLabels() {
	il := &IgnoreLabels{}
	il.convLabels(m.Processor.LabelIgnores)
	m.ignoreLabels.Store(il)
}

// GetIgnoreLabels 从原子变量中获取 *IgnoreLabels，确保并发安全。
func (m *Metrics) GetIgnoreLabels() *IgnoreLabels {
	il, ok := m.ignoreLabels.Load().(*IgnoreLabels)
	if !ok || il == nil { // 理论上不会进入该分支。
		il = &IgnoreLabels{}
		m.ignoreLabels.Store(il)
		return il
	}
	return il
}

// IgnoreCustomLabels 屏蔽自定义监控项的维度
func (l *IgnoreLabels) IgnoreCustomLabels(c *model.CustomMetrics) {
	if l.Enabled(c.MonitorName) {
		for i := range c.CustomLabels {
			if l.IgnoreCustomLabel(c.MonitorName, c.CustomLabels[i].Name) {
				c.CustomLabels[i].Value = IgnoredValue
			}
		}
	}
}

// IgnoreClientLabels 屏蔽主调监控的维度
func (l *IgnoreLabels) IgnoreClientLabels(c *model.ClientMetrics) {
	if l.Enabled(model.RPCClient) {
		for i := range c.RpcLabels.Fields {
			if l.IgnoreRPCLabel(model.RPCClient, c.RpcLabels.Fields[i].Name) {
				c.RpcLabels.Fields[i].Value = IgnoredValue
			}
		}
	}
}

// IgnoreServerLabels 屏蔽被调监控的维度
func (l *IgnoreLabels) IgnoreServerLabels(c *model.ServerMetrics) {
	if l.Enabled(model.RPCServer) {
		for i := range c.RpcLabels.Fields {
			if l.IgnoreRPCLabel(model.RPCServer, c.RpcLabels.Fields[i].Name) {
				c.RpcLabels.Fields[i].Value = IgnoredValue
			}
		}
	}
}

// IgnoreCustomLabel 自定义指标的对应标签是否要忽略
func (l *IgnoreLabels) IgnoreCustomLabel(monitorName, labelName string) bool {
	return l.monitorNameToCustomLabels[monitorName][labelName]
}

// Enabled 判断是否启用了维度屏蔽功能。
func (l *IgnoreLabels) Enabled(monitorName string) bool {
	switch monitorName {
	case model.RPCClient, model.RPCServer:
		return len(l.monitorNameToRPCLabels[monitorName]) > 0
	case model.GoRuntime, model.Process:
		return false
	default:
		return len(l.monitorNameToCustomLabels[monitorName]) > 0
	}
}

// IgnoreRPCLabel 判断 RPC 指标的对应标签是否要忽略
func (l *IgnoreLabels) IgnoreRPCLabel(monitorName string, labelName model.RPCLabels_FieldName) bool {
	return l.monitorNameToRPCLabels[monitorName][labelName]
}

func (l *IgnoreLabels) setRPCLabel(monitorName string, rpcLabel model.RPCLabels_FieldName) {
	if l.monitorNameToRPCLabels == nil {
		l.monitorNameToRPCLabels = make(map[string]map[model.RPCLabels_FieldName]bool)
	}
	if l.monitorNameToRPCLabels[monitorName] == nil {
		l.monitorNameToRPCLabels[monitorName] = make(map[model.RPCLabels_FieldName]bool)
	}
	l.monitorNameToRPCLabels[monitorName][rpcLabel] = true
}

func (l *IgnoreLabels) setCustomLabel(monitorName string, labelName string) {
	if l.monitorNameToCustomLabels == nil {
		l.monitorNameToCustomLabels = make(map[string]map[string]bool)
	}
	if l.monitorNameToCustomLabels[monitorName] == nil {
		l.monitorNameToCustomLabels[monitorName] = make(map[string]bool)
	}
	l.monitorNameToCustomLabels[monitorName][labelName] = true
}

func (l *IgnoreLabels) convLabels(labelIgnores []model.LabelIgnore) {
	for i := range labelIgnores {
		monitorName := labelIgnores[i].MonitorName
		for _, labelName := range labelIgnores[i].LabelNames {
			l.convLabel(monitorName, labelName)
		}
	}
}

func (l *IgnoreLabels) convLabel(monitorName string, labelName string) {
	switch monitorName {
	case model.RPCClient, model.RPCServer:
		if rpcLabel, ok := convRPCLabel(labelName); ok == nil {
			l.setRPCLabel(monitorName, rpcLabel)
		}
	case model.GoRuntime, model.Process:
		// 无维度，不需要处理
	default:
		// 其他均是自定义监控，屏蔽自定义维度
		l.setCustomLabel(monitorName, labelName)
	}
}

// convRPCLabel 主被调标签从字符串转枚举。
func convRPCLabel(labelName string) (model.RPCLabels_FieldName, error) {
	if labelField, ok := model.RPCLabels_FieldName_value[labelName]; ok {
		return model.RPCLabels_FieldName(labelField), nil
	}
	return 0, fmt.Errorf("not found rpc label")
}
