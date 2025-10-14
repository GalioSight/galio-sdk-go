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

package profiles

import (
	"testing"

	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/version"
	"github.com/stretchr/testify/assert"
)

func TestNewProfilesConfig(t *testing.T) {
	resource := &model.Resource{
		Target:        "PCG-123.example.greeter",  // 观测对象的唯一标识 ID，需要全局唯一
		Namespace:     "Development",              // 物理环境
		EnvName:       "test",                     // 用户环境
		Region:        "sz",                       // 地域
		Instance:      "aaa.bbb.ccc.ddd",          // 实例 ip
		Node:          "cls-as9z3nec-2",           // 节点
		ContainerName: "test.example.greeter.sz1", // 容器
		Version:       version.Number,             // SDK 版本号
		Platform:      "PCG-123",                  // 平台
		ObjectName:    "example.greeter",          // 对象名称
	}
	profiles := NewConfig(
		resource,
		WithProfilesEnable(false),
		WithExporter(&model.ProfilesExporter{}),
		WithProcessor(&model.ProfilesProcessor{}),
	)
	assert.False(t, profiles.Enable)
	assert.Equal(t, "", profiles.Exporter.Protocol)
	assert.Equal(t, "", profiles.Processor.Protocol)
}
