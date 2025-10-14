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
	"go.opentelemetry.io/otel/attribute"
	otelres "go.opentelemetry.io/otel/sdk/resource"

	"galiosight.ai/galio-sdk-go/model"
	omp3 "galiosight.ai/galio-sdk-go/semconv"
	semconv "galiosight.ai/galio-sdk-go/semconv/v1.0.0"
)

type SchemaType int

const (
	SchemaTypeMetric SchemaType = iota
	SchemaTypeTrace
	SchemaTypeLog
)

// GenResource generates an OpenTelemetry resource based on the provided schema URL and model resource.
// It determines which version of the resource to generate based on the schema URL.
func GenResource(schemaURL string, res *model.Resource, t SchemaType) *otelres.Resource {
	if schemaURL == semconv.SchemaURL {
		switch t {
		case SchemaTypeLog:
			return genLogResourceV1(schemaURL, res)
		case SchemaTypeTrace:
			return genTraceResourceV1(schemaURL, res)
		}
	}
	return GenResourceV3(schemaURL, res)
}

// func genMetricResourceV1(schemaURL string, res *model.Resource) *otelres.Resource {
// 	// omp v1.0.0 metric schema 比较混乱，默认的 v1.0.0 的 metric schemaURL 是 prom-v1.0.0 即 prometheus 类型的 _target_
// 	// 这里直接默认只支持 v3.0.0, 即不应该调用到该函数
// 	return otelres.NewWithAttributes(schemaURL, omitEmptyString(
// 		semconv.TargetKey.String(res.GetTarget()),                    // 观测对象的唯一标识 ID，需要全局唯一
// 		semconv.NamespaceKey.String(res.GetNamespace()),              // 物理环境
// 		semconv.EnvNameKey.String(res.GetEnvName()),                  // 用户环境
// 		semconv.ConSetidKey.String(res.GetSetName()),                 // 本机 setID
// 		semconv.InstanceKey.String(res.GetInstance()),                // 实例，IP
// 		semconv.ContainerNameSnakeKey.String(res.GetContainerName()), // 容器名
// 		semconv.VersionKey.String(version.Number),                    // 版本
// 		semconv.TpsTenantIDKey.String(res.GetTenantId()),             // 租户
// 	)...)
// }

func genTraceResourceV1(schemaURL string, res *model.Resource) *otelres.Resource {
	return otelres.NewWithAttributes(schemaURL, omitEmptyString(
		semconv.TargetKey.String(res.GetTarget()),
		semconv.TpsTenantIDKey.String(res.GetTenantId()),
		semconv.CmdbModuleIDKey.String(res.GetCmdbId()),
		omp3.TelemetrySDKLanguageGo,
		omp3.TelemetrySDKNameKey.String(model.TpsTelemetryName),
		omp3.ServiceNameKey.String(res.GetObjectName()),
		semconv.SetNameKey.String(res.GetSetName()),
		omp3.ContainerNameKey.String(res.GetContainerName()),
	)...)
}

func genLogResourceV1(schemaURL string, res *model.Resource) *otelres.Resource {
	return otelres.NewWithAttributes(schemaURL, omitEmptyString(
		semconv.TpsTenantIDKey.String(res.GetTenantId()),
		semconv.ServerKey.String(res.GetObjectName()),
		semconv.EnvKey.String(res.GetEnvName()),
		semconv.TargetKey.String(res.GetTarget()),
		semconv.NamespaceKey.String(res.GetNamespace()),
		semconv.SetNameCamelKey.String(res.GetSetName()),
		semconv.ContainerNameCamelKey.String(res.GetContainerName()),
		semconv.InstanceKey.String(res.GetInstance()),
	)...)
}

// GenResourceV3 生成 OMP v3 版本的 resource
func GenResourceV3(schemaURL string, r *model.Resource) *otelres.Resource {
	return otelres.NewWithAttributes(schemaURL, omitEmptyString(
		omp3.TelemetryTargetKey.String(r.GetTarget()),
		omp3.DeploymentNamespaceKey.String(r.GetNamespace()),
		omp3.DeploymentEnvironmentNameKey.String(r.GetEnvName()),
		omp3.TelemetrySDKNameKey.String(model.TpsTelemetryName),
		omp3.ContainerNameKey.String(r.GetContainerName()),
		omp3.HostIPKey.String(r.GetInstance()),
		omp3.ServiceSetNameKey.String(r.GetSetName()),
		omp3.TelemetrySDKLanguageGo,
	)...)
}

func omitEmptyString(kv ...attribute.KeyValue) []attribute.KeyValue {
	var ret []attribute.KeyValue
	for i := range kv {
		if kv[i].Value.Type() == attribute.STRING && kv[i].Value.AsString() == "" {
			continue
		}
		ret = append(ret, kv[i])
	}
	return ret
}
