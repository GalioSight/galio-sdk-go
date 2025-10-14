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
	"context"

	"go.opentelemetry.io/otel/trace"
	logpb "go.opentelemetry.io/proto/otlp/logs/v1"

	"galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/configs/ocp"
	"galiosight.ai/galio-sdk-go/errs"
	"galiosight.ai/galio-sdk-go/model"
)

// MetricsExporter 监控导出器。
type MetricsExporter interface {
	// GetStats 获取自监控统计数据。
	GetStats() *model.SelfMonitorStats
	// Export 导出监控。
	Export(metrics *model.Metrics)
	// UpdateConfig 更新配置。
	UpdateConfig(cfg *configs.Metrics)
}

// TracesExporter 追踪导出器。
type TracesExporter interface {
	trace.Tracer
	// Watcher 为了观察 ocp 的配置更新
	ocp.Watcher
	// UpdateConfig 更新配置。
	UpdateConfig(cfg *configs.Traces)
}

// LogsExporter 日志导出器。
type LogsExporter interface {
	// ExportLogs 导出日志。
	ExportLogs(context.Context, []*logpb.ResourceLogs) error
	// Shutdown 关闭导出器。
	Shutdown(ctx context.Context) error
}

// ProfilesExporter 性能数据导出器
type ProfilesExporter interface {
	// Export 导出（上传）性能数据
	Export(profiles *model.ProfilesBatch)
	UpdateConfig(cfg *configs.Profiles)
	Shutdown()
}

// ExporterFactory 导出器工厂。
type ExporterFactory interface {
	BaseFactory
	// CreateMetricsExporter 创建监控导出器。
	CreateMetricsExporter(cfg *configs.Metrics) (MetricsExporter, error)
	// CreateTracesExporter 创建追踪导出器。
	CreateTracesExporter(cfg *configs.Traces) (TracesExporter, error)
	// CreateLogsExporter 创建日志导出器。
	CreateLogsExporter(cfg *configs.Logs) (LogsExporter, error)
	// CreateProfilesExporter 创建性能数据导出器。
	CreateProfilesExporter(cfg *configs.Profiles) (ProfilesExporter, error)
}

// createMetricsExporter 创建监控导出器函数，定义成类型，方便各个实现 with option。
type createMetricsExporter func(cfg *configs.Metrics) (MetricsExporter, error)

// CreateMetricsExporter 创建监控导出器，其实就是调用自身，函数类型实现接口。
func (c createMetricsExporter) CreateMetricsExporter(cfg *configs.Metrics) (MetricsExporter, error) {
	if c == nil {
		return nil, errs.ErrCreateMetricsExporter
	}
	return c(cfg)
}

// createTracesExporter 创建追踪导出器函数，定义成类型，方便各个实现 with option。
type createTracesExporter func(cfg *configs.Traces) (TracesExporter, error)

// CreateTracesExporter 创建追踪导出器，其实就是调用自身，函数类型实现接口。
func (c createTracesExporter) CreateTracesExporter(cfg *configs.Traces) (TracesExporter, error) {
	if c == nil {
		return nil, errs.ErrCreateTracesExporter
	}
	return c(cfg)
}

// createLogsExporter 创建日志导出器函数，定义成类型，方便各个实现 with option。
type createLogsExporter func(cfg *configs.Logs) (LogsExporter, error)

// CreateLogsExporter 创建日志导出器，其实就是调用自身，函数类型实现接口。
func (c createLogsExporter) CreateLogsExporter(cfg *configs.Logs) (LogsExporter, error) {
	if c == nil {
		return nil, errs.ErrCreateLogsExporter
	}
	return c(cfg)
}

// createProfilesExporter 创建性能数据导出器函数，定义成类型，方便各个实现 with option。
type createProfilesExporter func(cfg *configs.Profiles) (ProfilesExporter, error)

// CreateProfilesExporter 创建日志导出器，其实就是调用自身，函数类型实现接口。
func (c createProfilesExporter) CreateProfilesExporter(cfg *configs.Profiles) (ProfilesExporter, error) {
	if c == nil {
		return nil, errs.ErrCreateProfilesExporter
	}
	return c(cfg)
}

type exporterFactory struct {
	createMetricsExporter
	createTracesExporter
	createLogsExporter
	createProfilesExporter
	baseFactory
}

// ExporterFactoryOption 构造导出器工厂的 option。
type ExporterFactoryOption func(e *exporterFactory)

// WithCreateMetricsExporter 设置监控导出器的生成函数。
func WithCreateMetricsExporter(c createMetricsExporter) ExporterFactoryOption {
	return func(e *exporterFactory) {
		e.createMetricsExporter = c
	}
}

// WithCreateTracesExporter 设置追踪导出器的生成函数。
func WithCreateTracesExporter(c createTracesExporter) ExporterFactoryOption {
	return func(e *exporterFactory) {
		e.createTracesExporter = c
	}
}

// WithCreateLogsExporter 设置日志导出器的生成函数。
func WithCreateLogsExporter(c createLogsExporter) ExporterFactoryOption {
	return func(e *exporterFactory) {
		e.createLogsExporter = c
	}
}

// WithCreateProfilesExporter 设置日志导出器的生成函数。
func WithCreateProfilesExporter(c createProfilesExporter) ExporterFactoryOption {
	return func(e *exporterFactory) {
		e.createProfilesExporter = c
	}
}

// NewExporterFactory 构造导出器工厂，由各个实现调用。
func NewExporterFactory(protocol string, options ...ExporterFactoryOption) ExporterFactory {
	e := &exporterFactory{
		baseFactory: baseFactory{protocol: protocol},
	}
	for _, option := range options {
		option(e)
	}
	return e
}
