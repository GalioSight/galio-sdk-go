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

// Package model ...
package model

import (
	"strings"

	modelv1 "galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/semconv"
	"galiosight.ai/galio-sdk-go/version"
)

// NewResource 构造一个符合 trpc 语义的可观测资源，避免用户缺少必要字段导致伽利略平台体验不畅
func NewResource(
	telemetryTarget string, // 观测对象的唯一标识 ID
	deploymentNamespace modelv1.Namespace, // 区分正式环境和测试环境
	deploymentEnvironmentName string, // Name of the deployment environment (aka deployment tier)
	hostIP string, // 本机 IP 地址
	containerName string, // 容器名
	serviceSetName string, // 将服务分组
	deploymentCity string, // 部署城市
	serviceVersion string, // The version string of the service API or implementation
	rpcSystem string, // 框架协议，如 trpc、http、grpc 等
) *modelv1.Resource {
	platform, objectName, ok := strings.Cut(telemetryTarget, ".")
	if !ok {
		panic("invalid TelemetryTarget: " + telemetryTarget)
	}
	res := &modelv1.Resource{
		Target:        telemetryTarget,
		Platform:      platform,
		ObjectName:    objectName,
		ServiceName:   objectName,
		Namespace:     string(deploymentNamespace),
		EnvName:       deploymentEnvironmentName,
		SetName:       serviceSetName,
		Region:        serviceSetName,
		Instance:      hostIP,
		ContainerName: containerName,
		Version:       version.Number,
		City:          deploymentCity,
		FrameCode:     rpcSystem,
		Language:      semconv.TelemetrySDKLanguageGo.Value.AsString(),
		SdkName:       "galileo",
	}

	if app, server, ok := strings.Cut(objectName, "."); ok {
		res.App = app
		res.Server = server
	}
	return res
}
