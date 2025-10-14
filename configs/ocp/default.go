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

package ocp

import (
	// 使用 embed 加载默认配置，由于默认配置数据较多，单独放到 default.yaml 文件中
	_ "embed"
	"log"

	"galiosight.ai/galio-sdk-go/model"
	"gopkg.in/yaml.v3"
)

//go:embed default.yaml
var defaultConfigData []byte

// DefaultConfig ocp server 的默认配置。
// 必须传入正确的 tenantID 才能上报。
func DefaultConfig(tenantID string) *model.GetConfigResponse {
	config := &model.GetConfigResponse{}
	err := yaml.Unmarshal(defaultConfigData, config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	config.TenantId = tenantID
	return config
}
