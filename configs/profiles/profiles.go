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

// Package profiles ...
package profiles

import (
	"galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/configs/ocp"
	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/self/metric"
)

type option func(*configs.Profiles)

// WithProfilesEnable 设置是否开启 Profiles
func WithProfilesEnable(enable bool) option {
	return func(t *configs.Profiles) {
		t.Enable = enable
	}
}

// WithExporter 设置 ProfilesExporter
// Deprecated: 直接覆盖整个对象有可能会导致配置错误，使用 ocp.WithLocalDecoder 代替。
func WithExporter(exporter *model.ProfilesExporter) option {
	return func(t *configs.Profiles) {
		t.Exporter = *exporter
	}
}

// WithProcessor 设置 ProfilesProcessor
// Deprecated: 直接覆盖整个对象有可能会导致配置错误，使用 ocp.WithLocalDecoder 代替。
func WithProcessor(processor *model.ProfilesProcessor) option {
	return func(t *configs.Profiles) {
		t.Processor = *processor
	}
}

// NewConfig 创建 Profiles 配置。
// 需要注意，此函数会被定时调用。
// 不要在此函数中创建重的对象。
// 坚决不能在此函数中创建协程。
func NewConfig(
	resource *model.Resource,
	opts ...option,
) *configs.Profiles {
	_ = ocp.RegisterResource(resource)
	getConfig := ocp.GetUpdater(resource.Target).GetConfig()
	config := getConfig.Config
	m := &configs.Profiles{
		Log:         logs.DefaultWrapper(),
		Resource:    *resource,
		SelfMonitor: config.SelfMonitor,
		Exporter:    config.ProfilesConfig.Exporter,
		Processor:   config.ProfilesConfig.Processor,
		Stats:       metric.GetSelfMonitor().Stats,
		APIKey:      getConfig.APIKey,
	}
	for _, o := range opts {
		o(m)
	}
	return m
}
