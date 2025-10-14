// Copyright 2025 Tencent Galileo Authors
//
// Copyright 2025 Tencent OpenTelemetry Oteam
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

// Package resource ...
package resource

import (
	otelres "go.opentelemetry.io/otel/sdk/resource"

	ires "galiosight.ai/galio-sdk-go/internal/resource"
	"galiosight.ai/galio-sdk-go/model"
)

// GenResource 对外暴露的 API，只用于生成 OMP v3 版本 resource
func GenResource(schemaURL string, res *model.Resource) *otelres.Resource {
	return ires.GenResourceV3(schemaURL, res)
}
