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

package otp

import (
	"galiosight.ai/galio-sdk-go/components"
	"galiosight.ai/galio-sdk-go/exporters/otp/metrics"
	"galiosight.ai/galio-sdk-go/exporters/otp/profiles"
	"galiosight.ai/galio-sdk-go/protocols"
)

const (
	protocol = protocols.OTP
)

// NewFactory 创建 otp 协议的导出器工厂。
func NewFactory() components.ExporterFactory {
	return components.NewExporterFactory(
		protocol,
		components.WithCreateMetricsExporter(metrics.NewExporter),
		components.WithCreateProfilesExporter(profiles.NewExporter),
	)
}
