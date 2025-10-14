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

package components

import (
	"galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/configs/ocp"
	"galiosight.ai/galio-sdk-go/errs"
	"galiosight.ai/galio-sdk-go/model"
)

// MetricsProcessor 监控处理器。
type MetricsProcessor interface {
	ocp.Watcher
	// GetStats 获取自监控统计数据。
	GetStats() *model.SelfMonitorStats
	// ProcessClientMetrics 处理主调监控。
	ProcessClientMetrics(clientMetrics *model.ClientMetrics)
	// ProcessServerMetrics 处理被调监控。
	ProcessServerMetrics(serverMetrics *model.ServerMetrics)
	// ProcessNormalMetric 处理属性监控。
	ProcessNormalMetric(normalMetric *model.NormalMetric)
	// ProcessCustomMetrics 处理用户自定义监控。
	ProcessCustomMetrics(customMetrics *model.CustomMetrics)
	// UpdateConfig 更新配置。
	UpdateConfig(cfg *configs.Metrics)
}

// TracesProcessor 追踪处理器。
type TracesProcessor interface {
	// Watcher 为了观察 ocp 的配置更新
	ocp.Watcher
	// ProcessTrace 处理追踪。
	ProcessTrace() // TODO(andyning): 协议定义。
}

// LogsProcessor 日志处理器。
type LogsProcessor interface {
	// Watcher 为了观察 ocp 的配置更新
	ocp.Watcher
	// ProcessLog 处理日志。
	ProcessLog() // TODO(andyning): 协议定义。
}

// ProfilesProcessor 性能数据处理器
type ProfilesProcessor interface {
	// Watcher 为了观察 ocp 的配置更新
	ocp.Watcher
	Start()
	// UpdateConfig 更新配置。
	UpdateConfig(cfg *configs.Profiles)
	Shutdown()
}

// ProcessorFactory 处理器工厂。
type ProcessorFactory interface {
	BaseFactory
	// CreateMetricsProcessor 创建监控处理器。
	CreateMetricsProcessor(cfg *configs.Metrics, exporter MetricsExporter) (MetricsProcessor, error)
	// CreateTracesProcessor 创建追踪处理器。
	CreateTracesProcessor(cfg *configs.Traces, exporter TracesExporter) (TracesProcessor, error)
	// CreateLogsProcessor 创建日志处理器。
	CreateLogsProcessor(cfg *configs.Logs, exporter LogsExporter) (LogsProcessor, error)
	// CreateProfilesProcessor 创建性能数据处理器
	CreateProfilesProcessor(cfg *configs.Profiles, exporter ProfilesExporter) (ProfilesProcessor, error)
}

// createMetricsProcessor 创建监控处理器函数，定义成类型，方便各个实现 with option。
type createMetricsProcessor func(cfg *configs.Metrics, exporter MetricsExporter) (MetricsProcessor, error)

// CreateMetricsProcessor 创建监控处理器，其实就是调用自身，函数类型实现接口。
func (c createMetricsProcessor) CreateMetricsProcessor(
	cfg *configs.Metrics,
	exporter MetricsExporter,
) (MetricsProcessor, error) {
	if c == nil {
		return nil, errs.ErrCreateMetricsProcessor
	}
	return c(cfg, exporter)
}

// createTracesProcessor 创建追踪处理器函数，定义成类型，方便各个实现 with option。
type createTracesProcessor func(cfg *configs.Traces, exporter TracesExporter) (TracesProcessor, error)

// CreateTracesProcessor 创建追踪处理器，其实就是调用自身，函数类型实现接口。
func (c createTracesProcessor) CreateTracesProcessor(
	cfg *configs.Traces,
	exporter TracesExporter,
) (TracesProcessor, error) {
	if c == nil {
		return nil, errs.ErrCreateTracesProcessor
	}
	return c(cfg, exporter)
}

// createLogsProcessor 创建日志处理器函数，定义成类型，方便各个实现 with option。
type createLogsProcessor func(cfg *configs.Logs, exporter LogsExporter) (LogsProcessor, error)

// CreateLogsProcessor 创建日志处理器，其实就是调用自身，函数类型实现接口。
func (c createLogsProcessor) CreateLogsProcessor(
	cfg *configs.Logs,
	exporter LogsExporter,
) (LogsProcessor, error) {
	if c == nil {
		return nil, errs.ErrCreateLogsProcessor
	}
	return c(cfg, exporter)
}

// createProfilesProcessor 创建性能数据处理器函数，定义成类型，方便各个实现 with option。
type createProfilesProcessor func(cfg *configs.Profiles, exporter ProfilesExporter) (ProfilesProcessor, error)

// CreateProfilesProcessor 创建性能数据处理器，其实就是调用自身，函数类型实现接口。
func (c createProfilesProcessor) CreateProfilesProcessor(
	cfg *configs.Profiles,
	exporter ProfilesExporter,
) (ProfilesProcessor, error) {
	if c == nil {
		return nil, errs.ErrCreateProfilesProcessor
	}
	return c(cfg, exporter)
}

type processorFactory struct {
	createMetricsProcessor
	createTracesProcessor
	createLogsProcessor
	createProfilesProcessor
	baseFactory
}

// ProcessorFactoryOption 构造处理器工厂的 option。
type ProcessorFactoryOption func(p *processorFactory)

// WithCreateMetricsProcessor 设置监控处理器的生成函数。
func WithCreateMetricsProcessor(c createMetricsProcessor) ProcessorFactoryOption {
	return func(p *processorFactory) {
		p.createMetricsProcessor = c
	}
}

// WithCreateTracesProcessor 设置追踪处理器的生成函数。
func WithCreateTracesProcessor(c createTracesProcessor) ProcessorFactoryOption {
	return func(p *processorFactory) {
		p.createTracesProcessor = c
	}
}

// WithCreateLogsProcessor 设置日志处理器的生成函数。
func WithCreateLogsProcessor(c createLogsProcessor) ProcessorFactoryOption {
	return func(p *processorFactory) {
		p.createLogsProcessor = c
	}
}

// WithCreateProfilesProcessor 设置性能数据处理器的生成函数。
func WithCreateProfilesProcessor(c createProfilesProcessor) ProcessorFactoryOption {
	return func(p *processorFactory) {
		p.createProfilesProcessor = c
	}
}

// NewProcessorFactory 构造处理器工厂，由各个实现调用。
func NewProcessorFactory(protocol string, options ...ProcessorFactoryOption) ProcessorFactory {
	m := &processorFactory{
		baseFactory: baseFactory{protocol: protocol},
	}
	for _, option := range options {
		option(m)
	}
	return m
}
