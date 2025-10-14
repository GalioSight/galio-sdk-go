// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2024 Tencent Galileo Authors

// Code generated from semantic convention specification. DO NOT EDIT.

package semconv // import "go.opentelemetry.io/otel/semconv/"

// Namespace: container

// Namespace: deployment

// Enum values for deployment.namespace
const (

	// 测试环境
	// Stability: development

	DeploymentNamespaceDevelopmentValue = "Development"
	// 正式环境
	// Stability: development

	DeploymentNamespaceProductionValue = "Production"
)

// Namespace: galileo

// Namespace: gen_ai

// Enum values for gen_ai.event
const (

	// llm 原始返回值
	// Stability: none

	GenAIEventPromptsValue = "gen_ai.prompts"
	// llm  通过原始返回值 提取出来的文本值
	// Stability: none

	GenAIEventCompletionsValue = "gen_ai.completions"
	// llm 第一个流事件
	// Stability: none

	GenAIEventFirstEventValue = "gen_ai.first_event"
	// llm 最后一个流事件
	// Stability: none

	GenAIEventLastEventValue = "gen_ai.last_event"
)

// Enum values for gen_ai.monitor
const (

	// llm 主调总体观测监控项
	// Stability: none

	GenAIMonitorClientValue = "LLMClient"
	// llm 主调流式事件监控项
	// Stability: none

	GenAIMonitorClientStreamValue = "LLMClientStream"
)

// Enum values for gen_ai.operation.name
const (

	// 文生文
	// Stability: none

	GenAIOperationNameChatValue = "chat"
)

// Enum values for gen_ai.system
const (

	// 太极提供商
	// Stability: none

	GenAISystemTaijiValue = "taiji"
	// 混元提供商
	// Stability: none

	GenAISystemHunyuanValue = "hunyuan"
	// 混元视频助手提供商
	// Stability: none

	GenAISystemHunyuanVideoAsstValue = "hunyuan_video_asst"
	// openai 提供商
	// Stability: none

	GenAISystemOpenAIValue = "openai"
)

// Namespace: host

// Namespace: message

// Enum values for message.type
const (

	// llm 原始返回值
	// Stability: none

	MessageTypeLLMPromptsValue = "llm_prompts"
	// llm llm 通过原始返回值 提取出来的文本值
	// Stability: none

	MessageTypeLLMCompletionsValue = "llm_completions"
	// llm llm 首 token
	// Stability: none

	MessageTypeLLMFirstTokenDataValue = "llm_first_token_data"
	// llm llm 尾 token
	// Stability: none

	MessageTypeLLMLastTokenDataValue = "llm_last_token_data"
)

// Namespace: messaging

// Namespace: other

// Enum values for llm_metrics
const (

	// llm 主调总体观测监控项
	// Stability: none

	LLMMetricsClientMonitorValue = "LLMClient"
	// llm 主调流式事件监控项
	// Stability: none

	LLMMetricsClientStreamMonitorValue = "LLMClientStream"
	// llm 请求总量指标
	// Stability: none

	LLMMetricsRequestCntValue = "request_cnt"
	// llm prompt token 指标
	// Stability: none

	LLMMetricsPromptTokensCntValue = "prompt_tokens_cnt"
	// llm completion token 指标
	// Stability: none

	LLMMetricsCompletionTokensCntValue = "completion_tokens_cnt"
	// llm completion events 指标
	// Stability: none

	LLMMetricsCompletionStreamEventsCntValue = "completion_stream_events_cnt"
	// llm 首 token 延迟指标
	// Stability: none

	LLMMetricsFirstTokenLatencySecondsValue = "first_token_latency_seconds"
	// llm 请求耗时指标
	// Stability: none

	LLMMetricsHandledSecondsValue = "handled_seconds"
	// llm token 秒级速率
	// Stability: none

	LLMMetricsPerSecondCompletionTokensNumValue = "per_second_completion_tokens_num"
)

// Enum values for llm_request_type
const (

	// 文生文
	// Stability: none

	LLMRequestTypeChatValue = "chat"
)

// Enum values for llm_vendor
const (

	// 太极提供商
	// Stability: none

	LLMVendorTaijiValue = "taiji"
	// 混元提供商
	// Stability: none

	LLMVendorHunyuanValue = "hunyuan"
	// 混元视频助手提供商
	// Stability: none

	LLMVendorHunyuanVideoAsstValue = "hunyuan_video_asst"
)

// Enum values for sse_metrics
const (

	// sse 被调总体观测监控项
	// Stability: none

	SSEMetricsSSEServerMonitorValue = "SSEServer"
	// sse 被调包粒度观测监控项
	// Stability: none

	SSEMetricsSSEServerChunkMonitorValue = "SSEServerChunk"
	// sse 主调总体观测监控项
	// Stability: none

	SSEMetricsSSEClientMonitorValue = "SSEClient"
	// sse 主调包粒度观测监控项
	// Stability: none

	SSEMetricsSSEClientChunkMonitorValue = "SSEClientChunk"
	// sse 请求总量指标
	// Stability: none

	SSEMetricsRequestCntValue = "request_cnt"
	// sse 首包耗时
	// Stability: none

	SSEMetricsFirstChunkSecondsValue = "first_chunk_seconds"
	// sse 请求耗时指标
	// Stability: none

	SSEMetricsHandledSecondsValue = "handled_seconds"
	// sse 包数
	// Stability: none

	SSEMetricsChunkCntValue = "chunk_cnt"
	// sse 每秒包数
	// Stability: none

	SSEMetricsPerSecondChunkCntValue = "per_second_chunk_cnt"
	// sse 总数据大小(B)
	// Stability: none

	SSEMetricsTotalDataBytesValue = "total_data_bytes"
)

// Enum values for sse_span_event
const (

	// sse 首包
	// Stability: none

	SSESpanEventFirstChunkValue = "sse_first_chunk"
	// sse 尾包
	// Stability: none

	SSESpanEventLastChunkValue = "sse_last_chunk"
)

// Namespace: rpc

// Enum values for rpc.error_code_type
const (

	// 成功
	// Stability: development

	RPCErrorCodeTypeSuccessValue = "success"
	// 异常
	// Stability: development

	RPCErrorCodeTypeExceptionValue = "exception"
	// 超时
	// Stability: development

	RPCErrorCodeTypeTimeoutValue = "timeout"
)

// Enum values for rpc.flow_tag
const (

	// 灰度流量
	// Stability: none

	RPCFlowTagGrayValue = "Gray"
	// 降级流量
	// Stability: none

	RPCFlowTagDowngradeValue = "Downgrade"
	// 重试流量
	// Stability: none

	RPCFlowTagRetryValue = "Retry"
)

// Enum values for rpc.system
const (

	// tRPC
	// Stability: development

	RPCSystemTrpcValue = "trpc"
	// gRPC
	// Stability: development

	RPCSystemGRPCValue = "grpc"
	// Java RMI
	// Stability: development

	RPCSystemJavaRmiValue = "java_rmi"
	// .NET WCF
	// Stability: development

	RPCSystemDotnetWcfValue = "dotnet_wcf"
	// Apache Dubbo
	// Stability: development

	RPCSystemApacheDubboValue = "apache_dubbo"
	// Connect RPC
	// Stability: development

	RPCSystemConnectRPCValue = "connect_rpc"
)

// Namespace: service

// Namespace: telemetry

// Enum values for telemetry.sdk.language
const (

	// go
	// Stability: stable

	TelemetrySDKLanguageGoValue = "go"
	// cpp
	// Stability: stable

	TelemetrySDKLanguageCPPValue = "cpp"
	// dotnet
	// Stability: stable

	TelemetrySDKLanguageDotnetValue = "dotnet"
	// erlang
	// Stability: stable

	TelemetrySDKLanguageErlangValue = "erlang"
	// java
	// Stability: stable

	TelemetrySDKLanguageJavaValue = "java"
	// nodejs
	// Stability: stable

	TelemetrySDKLanguageNodejsValue = "nodejs"
	// php
	// Stability: stable

	TelemetrySDKLanguagePHPValue = "php"
	// python
	// Stability: stable

	TelemetrySDKLanguagePythonValue = "python"
	// ruby
	// Stability: stable

	TelemetrySDKLanguageRubyValue = "ruby"
	// rust
	// Stability: stable

	TelemetrySDKLanguageRustValue = "rust"
	// swift
	// Stability: stable

	TelemetrySDKLanguageSwiftValue = "swift"
	// webjs
	// Stability: stable

	TelemetrySDKLanguageWebjsValue = "webjs"
)

// Namespace: tps

// Namespace: trace

// Namespace: workflow
