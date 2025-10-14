// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2024 Tencent Galileo Authors

// Code generated from semantic convention specification. DO NOT EDIT.

package semconv // import "go.opentelemetry.io/otel/semconv/"

// Namespace: cmdb

// Namespace: net

// Namespace: other

// Enum values for code_type
const (

	// 成功
	// Stability: development

	CodeTypeSuccessValue = "success"
	// 异常
	// Stability: development

	CodeTypeExceptionValue = "exception"
	// 超时
	// Stability: development

	CodeTypeTimeoutValue = "timeout"
)

// Enum values for flow_tag
const (

	// 灰度流量
	// Stability: none

	FlowTagGrayValue = "Gray"
	// 降级流量
	// Stability: none

	FlowTagDowngradeValue = "Downgrade"
	// 重试流量
	// Stability: none

	FlowTagRetryValue = "Retry"
)

// Enum values for namespace
const (

	// 测试环境
	// Stability: development

	NamespaceDevelopmentValue = "Development"
	// 正式环境
	// Stability: development

	NamespaceProductionValue = "Production"
)

// Namespace: server

// Namespace: set

// Namespace: tps

// Namespace: trpc
