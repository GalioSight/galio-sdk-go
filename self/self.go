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

// Package self 自监控组件的初始化。
package self

import (
	"galiosight.ai/galio-sdk-go/configs/ocp"
	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
	selfmetric "galiosight.ai/galio-sdk-go/self/metric"
)

// Init 初始化自监控，此处使用默认配置。
// 初始化自监控之前，必须要通过 ocp.RegisterResource 注册配置，以获得正确的配置。
// 同一个 target 只有第一次 RegisterResource 会生效。
// 所以此处增加一次 RegisterResource 调用，避免用户忘记 RegisterResource 导致问题。
func Init(resource *model.Resource, opts ...selfmetric.InitOption) {
	_ = ocp.RegisterResource(resource)
	config := ocp.GetUpdater(resource.Target).GetConfig().Config
	SetupObserver(resource, logs.DefaultWrapper(), config.SelfMonitor, config.ConfigServer, opts...)
}

// SetupObserver 设置自监控组件，包括 self log, self metrics, self schema。
func SetupObserver(
	resource *model.Resource,
	logWrapper *logs.Wrapper,
	selfMetric model.SelfMonitor,
	schemaURL string,
	opts ...selfmetric.InitOption,
) {
	selfmetric.Init(
		resource,
		selfMetric,
		logWrapper,
		opts...,
	)
}
