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
	"testing"

	"galiosight.ai/galio-sdk-go/model"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig("default")
	assert.Equal(t, "default", config.TenantId)
	// 默认情况下，访问点必须是未配置的，由后台判断。
	// 除非用户主动配置。
	// 如果业务同一个环境有多个城市，包含国外城市的话，只能依赖自动判断。
	assert.Equal(t, model.AccessPoint_ACCESS_POINT_INVALID, config.AccessPoint)
	assert.Equal(t, "default", config.PrometheusPush.Grouping["tps_tenant_id"])
	_, err := yaml.Marshal(config)
	assert.Nil(t, err)
}
