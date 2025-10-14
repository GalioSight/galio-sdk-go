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

// Package metrics 指标配置
package metrics

import (
	"galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/configs/ocp"
	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/self/metric"
	semconv "galiosight.ai/galio-sdk-go/semconv/v1.0.0"
)

type option func(*configs.Metrics)

// WithConvertName 是否转换中文指标名。
func WithConvertName(b bool) option {
	return func(metrics *configs.Metrics) {
		metrics.ConvertName = b
	}
}

// WithExporter 设置 MetricsExporter.
// Deprecated: 直接覆盖整个对象有可能会导致配置错误，使用 ocp.WithLocalDecoder 代替。
func WithExporter(exporter *model.MetricsExporter) option {
	return func(m *configs.Metrics) {
		m.Exporter = *exporter
	}
}

// WithProcessor 设置 MetricsProcessor.
// Deprecated: 直接覆盖整个对象有可能会导致配置错误，使用 ocp.WithLocalDecoder 代替。
func WithProcessor(processor *model.MetricsProcessor) option {
	return func(m *configs.Metrics) {
		m.Processor = *processor
	}
}

// WithSchemaURL 设置 SchemaURL.
func WithSchemaURL(schemaURL string) option {
	return func(m *configs.Metrics) {
		m.SchemaURL = schemaURL
	}
}

// NewConfig 创建 Metrics 配置。
// 需要注意，此函数会被定时调用。
// 不要在此函数中创建重的对象。
// 坚决不能在此函数中创建协程。
func NewConfig(
	resource *model.Resource,
	opts ...option,
) *configs.Metrics {
	_ = ocp.RegisterResource(resource)
	getConfig := ocp.GetUpdater(resource.Target).GetConfig()
	config := getConfig.Config
	m := &configs.Metrics{
		Log:         logs.DefaultWrapper(),
		Resource:    *resource,
		SelfMonitor: config.SelfMonitor,
		Exporter:    config.MetricsConfig.Exporter,
		Processor:   config.MetricsConfig.Processor,
		Stats:       metric.GetSelfMonitor().Stats,
		ConvertName: true,
		SchemaURL:   semconv.SchemaURL,
		APIKey:      getConfig.APIKey,
	}
	for _, o := range opts {
		o(m)
	}
	return m
}
