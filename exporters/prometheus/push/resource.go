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

// Package push ...
package push

import (
	"github.com/prometheus/client_golang/prometheus/push"

	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/semconv"
	"galiosight.ai/galio-sdk-go/version"
)

func addGroupingV1(pusher *push.Pusher, res *model.Resource) {
	// 伽利略保留字段，以下划线开头结尾
	pusher.Grouping("_target_", res.Target)
	pusher.Grouping("_namespace_", res.Namespace)
	pusher.Grouping("_env_name_", res.EnvName)
	pusher.Grouping("_instance_", res.Instance)
	pusher.Grouping("_container_name_", res.ContainerName)
	pusher.Grouping("_version_", res.Version)
	pusher.Grouping("_con_setid_", res.SetName)
	pusher.Grouping("_city_", res.City)
	// 以下 app server 字段，为了兼容天机阁而保留，不属于伽利略保留字段
	pusher.Grouping("app", res.App)
	pusher.Grouping("server", res.Server)
}

func addGroupingV3(pusher *push.Pusher, res *model.Resource) {
	V := semconv.PrometheusValue
	pusher.Grouping(V(semconv.TelemetryTargetKey.String(res.GetTarget())))
	pusher.Grouping(V(semconv.DeploymentNamespaceKey.String(res.GetNamespace())))
	pusher.Grouping(V(semconv.DeploymentEnvironmentNameKey.String(res.GetEnvName())))
	pusher.Grouping(V(semconv.ServiceSetNameKey.String(res.GetSetName())))
	pusher.Grouping(V(semconv.HostIPKey.String(res.GetInstance())))
	pusher.Grouping(V(semconv.ContainerNameKey.String(res.GetContainerName())))
	pusher.Grouping(V(semconv.TelemetrySDKVersionKey.String(version.Number)))
	pusher.Grouping(V(semconv.DeploymentCityKey.String(res.GetCity())))
	pusher.Grouping(V(semconv.ServiceVersionKey.String(res.GetReleaseVersion())))
	pusher.Grouping(V(semconv.TelemetrySDKLanguageGo))
	pusher.Grouping(V(semconv.TelemetrySDKNameKey.String(model.TpsTelemetryName)))
}
