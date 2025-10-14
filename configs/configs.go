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

// Package configs 伽利略 SDK 配置
package configs

import (
	"sync"
	"sync/atomic"

	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
)

// Metrics 监控配置。
type Metrics struct {
	Log                *logs.Wrapper          `yaml:"-"`            // 日志 wrapper。
	HistogramBuckets   map[string]*Bucket     `yaml:"-"`            // 间接配置，从 pb 中的配置构造而成，方便使用。
	ignoreLabels       atomic.Value           `yaml:"-"`            // *IgnoreLabels，从 pb 中的配置构造而成，方便使用。
	SecondGranularitys *SecondGranularitys    `yaml:"-"`            // 秒级监控配置，间接配置，从 pb 中的配置构造而成，方便使用。
	Resource           model.Resource         `yaml:"resource"`     // 资源信息。
	Processor          model.MetricsProcessor `yaml:"processor"`    // 处理器配置。
	SelfMonitor        model.SelfMonitor      `yaml:"self_monitor"` // 自监控配置。
	Exporter           model.MetricsExporter  `yaml:"exporter"`     // 导出器配置。
	Mu                 sync.RWMutex
	Stats              *model.SelfMonitorStats `yaml:"Stats"` // 自监控状态对象
	// 是否转换上报的指标名。
	// 因为 prometheus 是不支持中文的，需要转换成英文才能正常工作。
	// 在伽利略中，指标名必须符合 OMP 规范，才能在 Web UI 上正常显示。
	// 所以此指标通常设置为 true.
	ConvertName bool `yaml:"convert_name"`
	SchemaURL   string
	// 用于数据上报身份认证
	APIKey string
}

// Traces 追踪配置。
type Traces struct {
	Log         *logs.Wrapper           `yaml:"-"`            // 日志 wrapper。
	Resource    model.Resource          `yaml:"resource"`     // 资源信息。
	SelfMonitor model.SelfMonitor       `yaml:"self_monitor"` // 自监控配置。
	Exporter    model.TracesExporter    `yaml:"exporter"`     // 导出器配置。
	Processor   model.TracesProcessor   `yaml:"processor"`    // 处理器配置。
	Stats       *model.SelfMonitorStats `yaml:"Stats"`        // 自监控状态对象
	SchemaURL   string
	// 用于数据上报身份认证
	APIKey string
}

// Logs 日志配置。
type Logs struct {
	Log         *logs.Wrapper           `yaml:"-"`
	Resource    model.Resource          `yaml:"resource"`
	Exporter    model.LogsExporter      `yaml:"exporter"`
	Processor   model.LogsProcessor     `yaml:"processor"`
	SelfMonitor model.SelfMonitor       `yaml:"self_monitor"`
	Stats       *model.SelfMonitorStats `yaml:"Stats"` // 自监控状态对象
	SchemaURL   string
	// 用于数据上报身份认证
	APIKey string
}

// Profiles 性能上报配置
type Profiles struct {
	// 是否开启，在配置更新（热加载）时，需将 ocp.pb ProfilesConfig 中的 Enable 传入该结构体，
	// 来控制是否开启 profile processor 和 exporter
	Enable      bool                    `yaml:"-"`
	Log         *logs.Wrapper           `yaml:"-"`
	Resource    model.Resource          `yaml:"resource"`
	Exporter    model.ProfilesExporter  `yaml:"exporter"`
	Processor   model.ProfilesProcessor `yaml:"processor"`
	SelfMonitor model.SelfMonitor       `yaml:"self_monitor"`
	Stats       *model.SelfMonitorStats `yaml:"Stats"` // 自监控状态对象
	// 用于数据上报身份认证
	APIKey string
}
