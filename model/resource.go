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

package model

import (
	"galiosight.ai/galio-sdk-go/version"
)

type Namespace string // Namespace 物理空间的类型别名

const (
	Production  Namespace = "Production"  // Production 物理空间枚举值
	Development Namespace = "Development" // Development 物理空间枚举值
)

// NewResource 构造一个符合 trpc 语义的可观测资源，避免用户缺少必要字段导致伽利略平台体验不畅
func NewResource(
	platform string, // 资源所在平台，必填，取值建议："PCG-123"、"STKE"等
	app string, // 应用名称，必填，如 "mobagame"
	server string, // 服务名称，必填，如 "apiserver"
	service string, // service 名称，此参数暂时无用了，为了兼容性，函数签名里面继续保留此参数。
	namespace Namespace, //  物理环境，必填只能在 Production 和 Development 枚举，建议正式环境使用 Production
	envName string, // 用户环境，必填，一般是 formal、test 或 形如 3c170118 等自定义
	setName string, // set 名称，可选，如 "set.sz.2"
	city string, // 地域，可选，如 "sz"、"sh"、"gz"
	instance string, // 实例 IP，可选
	containerName string, // 容器逻辑名，可选，如 "test.galileo.sdkserver.sz10010"
) *Resource {
	return &Resource{
		Target:     platform + "." + app + "." + server, // 必须字段，观测对象（模块）的唯一标识 ID，须全局唯一，若无 Target 需到伽利略平台注册
		Platform:   platform,
		ObjectName: app + "." + server, // 对象名称，必填，一般是 Target 剔除 Platform 后的值，不填影响拉取 OCP 配置
		App:        app,
		Server:     server,
		// 如果不填 ServiceName 会影响伽利略平台查询不到 trace
		ServiceName:   app + "." + server, // 等于 ObjectName
		Namespace:     string(namespace),
		EnvName:       envName,
		SetName:       setName,        // setName
		Region:        setName,        // setName
		Instance:      instance,       // 实例 IP
		ContainerName: containerName,  // 容器逻辑名
		Version:       version.Number, // SDK 版本号
		City:          city,
		FrameCode:     "trpc",
		Language:      "go",
		SdkName:       "galileo",
	}
}

// FixNamespace 如果不是测试环境，则是正式环境。namespace 只有这两个取值。
func (r *Resource) FixNamespace() {
	if r.Namespace != string(Development) {
		r.Namespace = string(Production)
	}
}
