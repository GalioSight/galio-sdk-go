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

// Package model 定义伽利略中的数据模型
package model

const (
	// TpsTelemetryName 本 SDK 名字
	TpsTelemetryName = "galileo"
	// Language 本 SDK 的语言
	Language = "go"
	// TenantHeaderKey 租户 ID 在 HTTP 头里面的字段名。不区分大小写。
	TenantHeaderKey = "X-Tps-TenantID"
	// TargetHeaderKey Target 在 HTTP 头里面的字段名。不区分大小写。
	TargetHeaderKey = "X-Galileo-Target"
	// SchemaURLHeaderKey SchemaURL 在 HTTP 头里面的字段名，不区分大小写。
	SchemaURLHeaderKey = "X-Galileo-Schema-URL"
	// APIKeyHeaderKey 用于数据上报身份认证
	APIKeyHeaderKey = "X-Galileo-API-Key"

	// TraceparentHeader w3c trace header 字段
	TraceparentHeader = "traceparent"
	// TracestateHeader w3c trace  state
	TracestateHeader = "tracestate"
	// BaggageHeader 业务方自定义的全链路感知信息
	BaggageHeader = "baggage"
	// IngestionTime collector 摄入数据时间，单位纳秒。
	IngestionTime = "ingestion.time"
)
