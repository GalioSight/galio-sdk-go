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

// Package helper 辅助类，用于简化对象的创建
package helper

import (
	"galiosight.ai/galio-sdk-go/components"
	"galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/errs"
	"galiosight.ai/galio-sdk-go/exporters/otlp"
	"galiosight.ai/galio-sdk-go/exporters/otp"
	"galiosight.ai/galio-sdk-go/processors/omp"
	"galiosight.ai/galio-sdk-go/self"
	selfmetric "galiosight.ai/galio-sdk-go/self/metric"
)

func buildFactories() *components.Factories {
	factories := &components.Factories{}
	factories.ProcessorFactories = components.BuildProcessorFactories(
		omp.NewFactory(),
	)
	factories.ExporterFactories = components.BuildExporterFactories(
		otp.NewFactory(),
		otlp.NewFactory(),
	)
	return factories
}

var globalFactories = buildFactories()

func getMetricsProcessor(
	cfg *configs.Metrics,
	exporter components.MetricsExporter,
) (components.MetricsProcessor, error) {
	factory, ok := globalFactories.ProcessorFactories[cfg.Processor.Protocol]
	if !ok {
		return nil, errs.ErrFactoryEmpty
	}
	return factory.CreateMetricsProcessor(cfg, exporter)
}

func getMetricsExporter(cfg *configs.Metrics) (components.MetricsExporter, error) {
	factory, ok := globalFactories.ExporterFactories[cfg.Exporter.Protocol]
	if !ok {
		return nil, errs.ErrFactoryEmpty
	}
	return factory.CreateMetricsExporter(cfg)
}

// GetMetricsProcessor 获取监控处理器。
func GetMetricsProcessor(cfg *configs.Metrics) (components.MetricsProcessor, error) {
	self.Init(&cfg.Resource, selfmetric.WithAPIKey(cfg.APIKey))
	exporter, err := getMetricsExporter(cfg)
	if err != nil {
		return nil, err
	}
	return getMetricsProcessor(cfg, exporter)
}

// GetLogsExporter 获取日志导出器。
func GetLogsExporter(cfg *configs.Logs) (components.LogsExporter, error) {
	self.Init(&cfg.Resource, selfmetric.WithAPIKey(cfg.APIKey))
	factory, ok := globalFactories.ExporterFactories[fixProtocol(cfg.Exporter.Protocol)]
	if !ok {
		return nil, errs.ErrFactoryEmpty
	}
	return factory.CreateLogsExporter(cfg)
}

// GetTracesExporter  获取追踪导出器。
// Deprecated: use base.NewTracesExporter instead.
func GetTracesExporter(cfg *configs.Traces) (components.TracesExporter, error) {
	self.Init(&cfg.Resource, selfmetric.WithAPIKey(cfg.APIKey))
	factory, ok := globalFactories.ExporterFactories[fixProtocol(cfg.Exporter.Protocol)]
	if !ok {
		return nil, errs.ErrFactoryEmpty
	}
	return factory.CreateTracesExporter(cfg)
}

func fixProtocol(p string) string {
	if p == "oltp" { // 修复老版本的拼写错误。
		return "otlp"
	}
	return p
}

func getProfilesProcessor(
	cfg *configs.Profiles,
	exporter components.ProfilesExporter,
) (components.ProfilesProcessor, error) {
	self.Init(&cfg.Resource, selfmetric.WithAPIKey(cfg.APIKey))
	factory, ok := globalFactories.ProcessorFactories[cfg.Processor.Protocol]
	if !ok {
		return nil, errs.ErrFactoryEmpty
	}
	return factory.CreateProfilesProcessor(cfg, exporter)
}

func getProfilesExporter(cfg *configs.Profiles) (components.ProfilesExporter, error) {
	factory, ok := globalFactories.ExporterFactories[cfg.Exporter.Protocol]
	if !ok {
		return nil, errs.ErrFactoryEmpty
	}
	return factory.CreateProfilesExporter(cfg)
}

// GetProfilesProcessor 获取性能数据处理器。
func GetProfilesProcessor(cfg *configs.Profiles) (components.ProfilesProcessor, error) {
	exporter, err := getProfilesExporter(cfg)
	if err != nil {
		return nil, err
	}
	return getProfilesProcessor(cfg, exporter)
}
