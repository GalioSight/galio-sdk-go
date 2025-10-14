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

// Package traces 追踪配置
package traces

import (
	"galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/configs/ocp"
	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/self/metric"
	semconv "galiosight.ai/galio-sdk-go/semconv/v1.0.0"
)

type option func(t *configs.Traces)

// WithExporter 设置 MetricsExporter .
// Deprecated: 直接覆盖整个对象有可能会导致配置错误，使用 ocp.WithLocalDecoder 代替。
func WithExporter(exporter *model.TracesExporter) option {
	return func(t *configs.Traces) {
		t.Exporter = *exporter
	}
}

// WithProcessor 设置 TracesProcessor .
// Deprecated: 直接覆盖整个对象有可能会导致配置错误，使用 ocp.WithLocalDecoder 代替。
func WithProcessor(processor *model.TracesProcessor) option {
	return func(t *configs.Traces) {
		t.Processor = *processor
	}
}

// WithSamplerFraction 设置 采样率 .
func WithSamplerFraction(f float64) option {
	return func(t *configs.Traces) {
		t.Processor.Sampler.Fraction = f
	}
}

// WithEnableProfile 设置 是否开启 span 关联 profile
func WithEnableProfile(enable bool) option {
	return func(t *configs.Traces) {
		t.Processor.EnableProfile = enable
	}
}

// WithSchemaURL 设置 SchemaURL.
func WithSchemaURL(schemaURL string) option {
	return func(t *configs.Traces) {
		t.SchemaURL = schemaURL
	}
}

// NewConfig 创建 Traces 配置。
// 需要注意，此函数会被定时调用。
// 不要在此函数中创建重的对象。
// 坚决不能在此函数中创建协程。
func NewConfig(
	resource *model.Resource,
	opts ...option,
) *configs.Traces {
	_ = ocp.RegisterResource(resource)
	getConfig := ocp.GetUpdater(resource.Target).GetConfig()
	config := getConfig.Config
	m := &configs.Traces{
		Log:         logs.DefaultWrapper(),
		Resource:    *resource,
		SelfMonitor: config.SelfMonitor,
		Exporter:    config.TracesConfig.Exporter,
		Processor:   config.TracesConfig.Processor,
		Stats:       metric.GetSelfMonitor().Stats,
		SchemaURL:   semconv.SchemaURL,
		APIKey:      getConfig.APIKey,
	}
	for _, o := range opts {
		o(m)
	}
	return m
}
