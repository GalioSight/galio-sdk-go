// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2024 Tencent Galileo Authors

// Code generated from semantic convention specification. DO NOT EDIT.

package semconv // import "go.opentelemetry.io/otel/semconv/"

import "go.opentelemetry.io/otel/attribute"

// Namespace: container
const (

	// 容器名
	// Stability: Development
	// Type: string
	//
	// Examples: "test.galileo.apiserver.sz100012"

	ContainerNameKey = attribute.Key("container.name")
)

// Namespace: deployment
const (

	// 部署城市，类似于 otel 的 cloud.region
	// Stability: Development
	// Type: string
	//
	// Examples: "gz"

	DeploymentCityKey = attribute.Key("deployment.city")
	// Name of the [deployment environment] (aka deployment tier).
	//
	// Stability: Development
	// Type: string
	//
	// Examples:
	// "formal",
	// "test",
	// "d22b30d0",
	// "staging",
	// "production",
	//
	// Note: `deployment.environment.name` does not affect the uniqueness constraints defined through
	// the `service.namespace`, `service.name` and `service.instance.id` resource attributes.
	// This implies that resources carrying the following attribute combinations MUST be
	// considered to be identifying the same service:
	//
	// 细分的环境，例如每个人都可以在测试环境创建自己的特性环境，自行开发实现隔离。
	// env_name 除了固定的 formal、test 外，通常形如 d22b30d0 这样的格式
	//
	// [deployment environment]: https://wikipedia.org/wiki/Deployment_environment

	DeploymentEnvironmentNameKey = attribute.Key("deployment.environment.name")
	// 为 deployment.environment.name 的超集，区分正式环境和测试环境
	// Stability: Stable
	// Type: Enum
	//
	// Examples: undefined

	DeploymentNamespaceKey = attribute.Key("deployment.namespace")
)

// Enum values for deployment.namespace
var (

	// 测试环境
	// Stability: development

	DeploymentNamespaceDevelopment = DeploymentNamespaceKey.String("Development")
	// 正式环境
	// Stability: development

	DeploymentNamespaceProduction = DeploymentNamespaceKey.String("Production")
)

// Namespace: galileo
const (

	// 采样策略的名称
	// Stability: Development
	// Type: string
	//
	// Examples:
	// "dyeing",
	// "follow root dyeing",

	GalileoSamplerKey = attribute.Key("galileo.sampler")
	// tracestate string
	// Stability: Development
	// Type: string
	//
	// Examples: "g=w:1:33683436c17a0f980c156349;s:5;r:4"

	GalileoStateKey = attribute.Key("galileo.state")
)

// Namespace: gen_ai
const (

	// llm span body 事件
	// Stability: Development
	// Type: Enum
	//
	// Examples: undefined

	GenAIEventKey = attribute.Key("gen_ai.event")
	// 是否为流式
	// Stability: Development
	// Type: string
	//
	// Examples: "true or false"

	GenAIIsStreamKey = attribute.Key("gen_ai.is_stream")
	// llm 相关指标
	// Stability: Development
	// Type: Enum
	//
	// Examples: undefined

	GenAIMonitorKey = attribute.Key("gen_ai.monitor")
	// 请求类型
	// Stability: Development
	// Type: Enum
	//
	// Examples: "chat"

	GenAIOperationNameKey = attribute.Key("gen_ai.operation.name")
	// 请求模型
	// Stability: Development
	// Type: string
	//
	// Examples: "Hunyuan-T1-32K"

	GenAIRequestModelKey = attribute.Key("gen_ai.request.model")
	// completion 事件数
	// Stability: Development
	// Type: string
	//
	// Examples: undefined

	GenAIResponseEventsKey = attribute.Key("gen_ai.response.events")
	// 返回首包错误码
	// Stability: Development
	// Type: string
	//
	// Examples: undefined

	GenAIResponseFirstCodeKey = attribute.Key("gen_ai.response.first_code")
	// 返回尾包错误码
	// Stability: Development
	// Type: string
	//
	// Examples: undefined

	GenAIResponseLastCodeKey = attribute.Key("gen_ai.response.last_code")
	// 返回模型
	// Stability: Development
	// Type: string
	//
	// Examples: "Hunyuan-T1-32K"

	GenAIResponseModelKey = attribute.Key("gen_ai.response.model")
	// 首 token 耗时
	// Stability: Development
	// Type: string
	//
	// Examples: undefined

	GenAIServerTimeToFirstTokenKey = attribute.Key("gen_ai.server.time_to_first_token")
	// session_id
	// Stability: Development
	// Type: string
	//
	// Examples: undefined

	GenAISessionIDKey = attribute.Key("gen_ai.session_id")
	// 系统提供商 string
	// Stability: Development
	// Type: Enum
	//
	// Examples: "taiji"

	GenAISystemKey = attribute.Key("gen_ai.system")
	// prompt token 数量
	// Stability: Development
	// Type: string
	//
	// Examples: undefined

	GenAIUsageInputTokensKey = attribute.Key("gen_ai.usage.input_tokens")
	// completion token数量
	// Stability: Development
	// Type: string
	//
	// Examples: undefined

	GenAIUsageOutputTokensKey = attribute.Key("gen_ai.usage.output_tokens")
)

// Enum values for gen_ai.event
var (

	// llm 原始返回值
	// Stability: none

	GenAIEventPrompts = GenAIEventKey.String("gen_ai.prompts")
	// llm  通过原始返回值 提取出来的文本值
	// Stability: none

	GenAIEventCompletions = GenAIEventKey.String("gen_ai.completions")
	// llm 第一个流事件
	// Stability: none

	GenAIEventFirstEvent = GenAIEventKey.String("gen_ai.first_event")
	// llm 最后一个流事件
	// Stability: none

	GenAIEventLastEvent = GenAIEventKey.String("gen_ai.last_event")
)

// Enum values for gen_ai.monitor
var (

	// llm 主调总体观测监控项
	// Stability: none

	GenAIMonitorClient = GenAIMonitorKey.String("LLMClient")
	// llm 主调流式事件监控项
	// Stability: none

	GenAIMonitorClientStream = GenAIMonitorKey.String("LLMClientStream")
)

// Enum values for gen_ai.operation.name
var (

	// 文生文
	// Stability: none

	GenAIOperationNameChat = GenAIOperationNameKey.String("chat")
)

// Enum values for gen_ai.system
var (

	// 太极提供商
	// Stability: none

	GenAISystemTaiji = GenAISystemKey.String("taiji")
	// 混元提供商
	// Stability: none

	GenAISystemHunyuan = GenAISystemKey.String("hunyuan")
	// 混元视频助手提供商
	// Stability: none

	GenAISystemHunyuanVideoAsst = GenAISystemKey.String("hunyuan_video_asst")
	// openai 提供商
	// Stability: none

	GenAISystemOpenAI = GenAISystemKey.String("openai")
)

// Namespace: host
const (

	// 本机 IP 地址
	// Stability: Development
	// Type: string
	//
	// Examples: "10.20.30.40"

	HostIPKey = attribute.Key("host.ip")
)

// Namespace: message
const (

	// llm span body 事件
	// Stability: Development
	// Type: Enum
	//
	// Examples: undefined

	MessageTypeKey = attribute.Key("message.type")
)

// Enum values for message.type
var (

	// llm 原始返回值
	// Stability: none

	MessageTypeLLMPrompts = MessageTypeKey.String("llm_prompts")
	// llm llm 通过原始返回值 提取出来的文本值
	// Stability: none

	MessageTypeLLMCompletions = MessageTypeKey.String("llm_completions")
	// llm llm 首 token
	// Stability: none

	MessageTypeLLMFirstTokenData = MessageTypeKey.String("llm_first_token_data")
	// llm llm 尾 token
	// Stability: none

	MessageTypeLLMLastTokenData = MessageTypeKey.String("llm_last_token_data")
)

// Namespace: messaging
const (

	// The name of the consumer group with which a consumer is associated.
	//
	// Stability: Development
	// Type: string
	//
	// Examples:
	// "my-group",
	// "indexer",
	//
	// Note: Name of the Kafka Consumer Group that is handling the message. Only applies to consumers, not producers

	MessagingKafkaConsumerGroupKey = attribute.Key("messaging.kafka.consumer.group")
)

// Namespace: other
const (

	// prometheus 上报指定监控项名
	// Stability:
	// Type: string
	//
	// Examples: "custom_counter_total"

	MonitorNameKey = attribute.Key("_monitor_name_")
	// 用于统计 span 根节点来源
	// Stability: Development
	// Type: string
	//
	// Examples: "trpc.http.upserver.upservice│PCG-123.galileo.openserver│2"
	// Note: 经常有用户表示自己跟随上游命中采样，数量非预期，希望知道是哪个服务发起的命中采样，从而决定是否要调整 其格式为 caller|callee|who, who为1表示是caller发起的, who为2表示是callee发起的

	AffinityAttributeKey = attribute.Key("affinity_attribute")
	// 显示用户通过 opentelemetry baggage API 设置的内容
	// Stability: Development
	// Type: string
	//
	// Examples: ""

	BaggageKey = attribute.Key("baggage")
	// 首包 code
	// Stability: Development
	// Type: string
	//
	// Examples: undefined

	FirstChunkCodeKey = attribute.Key("first_chunk_code")
	// 是否卡顿
	// Stability: Development
	// Type: string
	//
	// Examples: undefined

	IsStuckKey = attribute.Key("is_stuck")
	// 尾包 code
	// Stability: Development
	// Type: string
	//
	// Examples: undefined

	LastChunkCodeKey = attribute.Key("last_chunk_code")
	// 行号
	// Stability: Development
	// Type: string
	//
	// Examples: "server.go:60"

	LineKey = attribute.Key("line")
	// completion 事件数
	// Stability: Development
	// Type: string
	//
	// Examples: undefined

	LLMCompletionStreamEventsKey = attribute.Key("llm_completion_stream_events")
	// 首 token 耗时
	// Stability: Development
	// Type: string
	//
	// Examples: undefined

	LLMFirstTokenLatencySecondsKey = attribute.Key("llm_first_token_latency_seconds")
	// 是否为流式
	// Stability: Development
	// Type: string
	//
	// Examples: "true or false"

	LLMIsStreamKey = attribute.Key("llm_is_stream")
	// llm 相关指标
	// Stability: Development
	// Type: Enum
	//
	// Examples: undefined

	LLMMetricsKey = attribute.Key("llm_metrics")
	// 请求模型
	// Stability: Development
	// Type: string
	//
	// Examples: "Hunyuan-T1-32K"

	LLMRequestModelKey = attribute.Key("llm_request_model")
	// 请求类型
	// Stability: Development
	// Type: Enum
	//
	// Examples: "chat"

	LLMRequestTypeKey = attribute.Key("llm_request_type")
	// 返回首包错误码
	// Stability: Development
	// Type: string
	//
	// Examples: undefined

	LLMResponseFirstCodeKey = attribute.Key("llm_response_first_code")
	// 返回尾包错误码
	// Stability: Development
	// Type: string
	//
	// Examples: undefined

	LLMResponseLastCodeKey = attribute.Key("llm_response_last_code")
	// 返回模型
	// Stability: Development
	// Type: string
	//
	// Examples: "Hunyuan-T1-32K"

	LLMResponseModelKey = attribute.Key("llm_response_model")
	// session_id
	// Stability: Development
	// Type: string
	//
	// Examples: undefined

	LLMSessionIDKey = attribute.Key("llm_session_id")
	// completion token数量
	// Stability: Development
	// Type: string
	//
	// Examples: undefined

	LLMUsageCompletionTokensKey = attribute.Key("llm_usage_completion_tokens")
	// prompt token 数量
	// Stability: Development
	// Type: string
	//
	// Examples: undefined

	LLMUsagePromptTokensKey = attribute.Key("llm_usage_prompt_tokens")
	// 提供商 string
	// Stability: Development
	// Type: Enum
	//
	// Examples: "taiji"

	LLMVendorKey = attribute.Key("llm_vendor")
	// 是否命中了 trace 采样
	// Stability: Development
	// Type: boolean
	//
	// Examples: undefined

	SampledKey = attribute.Key("sampled")
	// sse trace tag，包数量
	// Stability: Development
	// Type: string
	//
	// Examples: undefined

	SSEChunkCntKey = attribute.Key("sse_chunk_cnt")
	// sse trace tag，首包耗时
	// Stability: Development
	// Type: string
	//
	// Examples: undefined

	SSEFirstChunkSecondsKey = attribute.Key("sse_first_chunk_seconds")
	// sse trace tag，是否卡顿
	// Stability: Development
	// Type: string
	//
	// Examples: undefined

	SSEIsStuckKey = attribute.Key("sse_is_stuck")
	// llm 相关指标
	// Stability: Development
	// Type: Enum
	//
	// Examples: undefined

	SSEMetricsKey = attribute.Key("sse_metrics")
	// sse trace tag，每秒包数
	// Stability: Development
	// Type: string
	//
	// Examples: undefined

	SSEPerSecondChunkCntKey = attribute.Key("sse_per_second_chunk_cnt")
	// sse span body 事件
	// Stability: Development
	// Type: Enum
	//
	// Examples: undefined

	SSESpanEventKey = attribute.Key("sse_span_event")
)

// Enum values for llm_metrics
var (

	// llm 主调总体观测监控项
	// Stability: none

	LLMMetricsClientMonitor = LLMMetricsKey.String("LLMClient")
	// llm 主调流式事件监控项
	// Stability: none

	LLMMetricsClientStreamMonitor = LLMMetricsKey.String("LLMClientStream")
	// llm 请求总量指标
	// Stability: none

	LLMMetricsRequestCnt = LLMMetricsKey.String("request_cnt")
	// llm prompt token 指标
	// Stability: none

	LLMMetricsPromptTokensCnt = LLMMetricsKey.String("prompt_tokens_cnt")
	// llm completion token 指标
	// Stability: none

	LLMMetricsCompletionTokensCnt = LLMMetricsKey.String("completion_tokens_cnt")
	// llm completion events 指标
	// Stability: none

	LLMMetricsCompletionStreamEventsCnt = LLMMetricsKey.String("completion_stream_events_cnt")
	// llm 首 token 延迟指标
	// Stability: none

	LLMMetricsFirstTokenLatencySeconds = LLMMetricsKey.String("first_token_latency_seconds")
	// llm 请求耗时指标
	// Stability: none

	LLMMetricsHandledSeconds = LLMMetricsKey.String("handled_seconds")
	// llm token 秒级速率
	// Stability: none

	LLMMetricsPerSecondCompletionTokensNum = LLMMetricsKey.String("per_second_completion_tokens_num")
)

// Enum values for llm_request_type
var (

	// 文生文
	// Stability: none

	LLMRequestTypeChat = LLMRequestTypeKey.String("chat")
)

// Enum values for llm_vendor
var (

	// 太极提供商
	// Stability: none

	LLMVendorTaiji = LLMVendorKey.String("taiji")
	// 混元提供商
	// Stability: none

	LLMVendorHunyuan = LLMVendorKey.String("hunyuan")
	// 混元视频助手提供商
	// Stability: none

	LLMVendorHunyuanVideoAsst = LLMVendorKey.String("hunyuan_video_asst")
)

// Enum values for sse_metrics
var (

	// sse 被调总体观测监控项
	// Stability: none

	SSEMetricsSSEServerMonitor = SSEMetricsKey.String("SSEServer")
	// sse 被调包粒度观测监控项
	// Stability: none

	SSEMetricsSSEServerChunkMonitor = SSEMetricsKey.String("SSEServerChunk")
	// sse 主调总体观测监控项
	// Stability: none

	SSEMetricsSSEClientMonitor = SSEMetricsKey.String("SSEClient")
	// sse 主调包粒度观测监控项
	// Stability: none

	SSEMetricsSSEClientChunkMonitor = SSEMetricsKey.String("SSEClientChunk")
	// sse 请求总量指标
	// Stability: none

	SSEMetricsRequestCnt = SSEMetricsKey.String("request_cnt")
	// sse 首包耗时
	// Stability: none

	SSEMetricsFirstChunkSeconds = SSEMetricsKey.String("first_chunk_seconds")
	// sse 请求耗时指标
	// Stability: none

	SSEMetricsHandledSeconds = SSEMetricsKey.String("handled_seconds")
	// sse 包数
	// Stability: none

	SSEMetricsChunkCnt = SSEMetricsKey.String("chunk_cnt")
	// sse 每秒包数
	// Stability: none

	SSEMetricsPerSecondChunkCnt = SSEMetricsKey.String("per_second_chunk_cnt")
	// sse 总数据大小(B)
	// Stability: none

	SSEMetricsTotalDataBytes = SSEMetricsKey.String("total_data_bytes")
)

// Enum values for sse_span_event
var (

	// sse 首包
	// Stability: none

	SSESpanEventFirstChunk = SSESpanEventKey.String("sse_first_chunk")
	// sse 尾包
	// Stability: none

	SSESpanEventLastChunk = SSESpanEventKey.String("sse_last_chunk")
)

// Namespace: rpc
const (

	// 被调容器
	// Stability: Development
	// Type: string
	//
	// Examples: "test.galileo.metaserver.sz100012"

	RPCCalleeContainerKey = attribute.Key("rpc.callee.container")
	// 被调 ip
	// Stability: Development
	// Type: string
	//
	// Examples: "10.30.50.70"

	RPCCalleeIPKey = attribute.Key("rpc.callee.ip")
	// 被调方法
	// Stability: Development
	// Type: string
	//
	// Examples: "GetPromQueryURLByTarget"

	RPCCalleeMethodKey = attribute.Key("rpc.callee.method")
	// 被调服务，123 平台上的服务为 app.server，可能会和别的平台的服务重名
	// Stability: Development
	// Type: string
	//
	// Examples: "galileo.metaserver"

	RPCCalleeServerKey = attribute.Key("rpc.callee.server")
	// 被调 service
	// Stability: Development
	// Type: string
	//
	// Examples: "trpc.galileo.metaserver.TargetDataService"

	RPCCalleeServiceKey = attribute.Key("rpc.callee.service")
	// 被调 set
	// Stability: Development
	// Type: string
	//
	// Examples: "set.sz.1"

	RPCCalleeSetKey = attribute.Key("rpc.callee.set")
	// 主调容器
	// Stability: Development
	// Type: string
	//
	// Examples: "test.galileo.apiserver.sz100012"

	RPCCallerContainerKey = attribute.Key("rpc.caller.container")
	// 主调流量分组
	// Stability: Development
	// Type: string
	//
	// Examples: "qq"
	// Note: 常用于中台内部服务需要根据主调业务划分监控，通常需要用户主动填充，rpc 框架无法自动填充

	RPCCallerGroupKey = attribute.Key("rpc.caller.group")
	// 主调 ip
	// Stability: Development
	// Type: string
	//
	// Examples: "10.20.30.40"

	RPCCallerIPKey = attribute.Key("rpc.caller.ip")
	// 主调方法
	// Stability: Development
	// Type: string
	//
	// Examples: "getData"

	RPCCallerMethodKey = attribute.Key("rpc.caller.method")
	// 主调服务，123 平台上的服务为 app.server，可能会和别的平台的服务重名
	// Stability: Development
	// Type: string
	//
	// Examples: "galileo.apiserver"

	RPCCallerServerKey = attribute.Key("rpc.caller.server")
	// 主调 service
	// Stability: Development
	// Type: string
	//
	// Examples: "trpc.galileo.apiserver.DataService"

	RPCCallerServiceKey = attribute.Key("rpc.caller.service")
	// 主调 set
	// Stability: Development
	// Type: string
	//
	// Examples: "set.sz.1"

	RPCCallerSetKey = attribute.Key("rpc.caller.set")
	// 金丝雀流量
	// Stability: Development
	// Type: string
	//
	// Examples: "1"

	RPCCanaryKey = attribute.Key("rpc.canary")
	// 返回码
	// Stability: Development
	// Type: string
	//
	// Examples:
	// "err_101",
	// "0",
	// "ret_100",
	//
	// Note: err_开头的默认表示异常，ret_开头的默认表示成功，特别的 0 表示成功

	RPCErrorCodeKey = attribute.Key("rpc.error_code")
	// 返回码类型，区分成功，异常，超时
	// Stability: Development
	// Type: Enum
	//
	// Examples: undefined

	RPCErrorCodeTypeKey = attribute.Key("rpc.error_code_type")
	// 返回码描述
	// Stability: Development
	// Type: string
	//
	// Examples: "404 not found"

	RPCErrorMessageKey = attribute.Key("rpc.error_message")
	// 定义流量标签，如灰度、降级、重试
	// Stability: Development
	// Type: Enum
	//
	// Examples: "1"

	RPCFlowTagKey = attribute.Key("rpc.flow_tag")
	// A string identifying the remoting system. See below for a list of well-known identifiers.
	// Stability: Development
	// Type: Enum
	//
	// Examples: undefined

	RPCSystemKey = attribute.Key("rpc.system")
	// trpc 协议的 request ID.
	// Stability: Development
	// Type: int
	//
	// Examples: undefined

	RPCTrpcRequestIDKey = attribute.Key("rpc.trpc.request_id")
	// 用户扩展字段 1
	// Stability: Development
	// Type: string
	//
	// Examples: ""

	RPCUserExt1Key = attribute.Key("rpc.user.ext1")
	// 用户扩展字段 2
	// Stability: Development
	// Type: string
	//
	// Examples: ""

	RPCUserExt2Key = attribute.Key("rpc.user.ext2")
	// 用户扩展字段 3
	// Stability: Development
	// Type: string
	//
	// Examples: ""

	RPCUserExt3Key = attribute.Key("rpc.user.ext3")
)

// Enum values for rpc.error_code_type
var (

	// 成功
	// Stability: development

	RPCErrorCodeTypeSuccess = RPCErrorCodeTypeKey.String("success")
	// 异常
	// Stability: development

	RPCErrorCodeTypeException = RPCErrorCodeTypeKey.String("exception")
	// 超时
	// Stability: development

	RPCErrorCodeTypeTimeout = RPCErrorCodeTypeKey.String("timeout")
)

// Enum values for rpc.flow_tag
var (

	// 灰度流量
	// Stability: none

	RPCFlowTagGray = RPCFlowTagKey.String("Gray")
	// 降级流量
	// Stability: none

	RPCFlowTagDowngrade = RPCFlowTagKey.String("Downgrade")
	// 重试流量
	// Stability: none

	RPCFlowTagRetry = RPCFlowTagKey.String("Retry")
)

// Enum values for rpc.system
var (

	// tRPC
	// Stability: development

	RPCSystemTrpc = RPCSystemKey.String("trpc")
	// gRPC
	// Stability: development

	RPCSystemGRPC = RPCSystemKey.String("grpc")
	// Java RMI
	// Stability: development

	RPCSystemJavaRmi = RPCSystemKey.String("java_rmi")
	// .NET WCF
	// Stability: development

	RPCSystemDotnetWcf = RPCSystemKey.String("dotnet_wcf")
	// Apache Dubbo
	// Stability: development

	RPCSystemApacheDubbo = RPCSystemKey.String("apache_dubbo")
	// Connect RPC
	// Stability: development

	RPCSystemConnectRPC = RPCSystemKey.String("connect_rpc")
)

// Namespace: service
const (

	// 服务名
	// Stability: Development
	// Type: string
	// Deprecated: use telemetry.target instead
	//
	// Examples: "galileo.apiserver"

	ServiceNameKey = attribute.Key("service.name")
	// 将服务分组，或者理解成打标签
	// Stability: Development
	// Type: string
	//
	// Examples: "set.sz.1"

	ServiceSetNameKey = attribute.Key("service.set.name")
	// The version string of the service API or implementation. The format is not defined by these conventions.
	//
	// Stability: Stable
	// Type: string
	//
	// Examples:
	// "2.0.0",
	// "a01dbef8a",

	ServiceVersionKey = attribute.Key("service.version")
)

// Namespace: telemetry
const (

	// 组织名，用于商业版中代替 target 中的 platform
	// Stability: Development
	// Type: string
	//
	// Examples: "513F30A49CF9"

	TelemetryOrganizationKey = attribute.Key("telemetry.organization")
	// The language of the telemetry SDK.
	//
	// Stability: Stable
	// Type: Enum
	//
	// Examples: undefined

	TelemetrySDKLanguageKey = attribute.Key("telemetry.sdk.language")
	// The name of the telemetry SDK as defined above.
	//
	// Stability: Stable
	// Type: string
	//
	// Examples:
	// "opentelemetry",
	// "galileo",
	//
	// Note: The OpenTelemetry SDK MUST set the `telemetry.sdk.name` attribute to `opentelemetry`.
	// If another SDK, like a fork or a vendor-provided implementation, is used, this SDK MUST set the
	// `telemetry.sdk.name` attribute to the fully-qualified class or module name of this SDK's main entry point
	// or another suitable identifier depending on the language.
	// The identifier `opentelemetry` is reserved and MUST NOT be used in this case.
	// All custom identifiers SHOULD be stable across different versions of an implementation

	TelemetrySDKNameKey = attribute.Key("telemetry.sdk.name")
	// The version string of the telemetry SDK.
	//
	// Stability: Stable
	// Type: string
	//
	// Examples:
	// "v0.17.0",
	// "cpp-1.17.0",
	// "py-1.17.0",

	TelemetrySDKVersionKey = attribute.Key("telemetry.sdk.version")
	// 如果支持多个部署平台，等于 platform + service.name, 否则等于 service.name
	// Stability: Stable
	// Type: string
	//
	// Examples: "PCG-123.galileo.metaserver"
	// Note: target 是 galileo 的核心概念之一，任何观测数据，都必须绑定到一个 target 上。 <p>target 是观测对象的唯一标识 ID，必须全局唯一，target 相同的，就认为是同一个对象。</p> <p>target 必须全局唯一，为了避免冲突，格式上分为两部分，用点分割。</p>
	//
	//   - 第一部分是平台 (platform)，第二部分是平台内的对象名称 (service.name)。
	//   - 第一个点之前的部分，是 platform，如 PCG-123，不同平台的 platform 是不同的。
	//   - 第一个点之后的部分，是 service.name，如 galileo.metaserver。 target = platform . service.name

	TelemetryTargetKey = attribute.Key("telemetry.target")
)

// Enum values for telemetry.sdk.language
var (

	// go
	// Stability: stable

	TelemetrySDKLanguageGo = TelemetrySDKLanguageKey.String("go")
	// cpp
	// Stability: stable

	TelemetrySDKLanguageCPP = TelemetrySDKLanguageKey.String("cpp")
	// dotnet
	// Stability: stable

	TelemetrySDKLanguageDotnet = TelemetrySDKLanguageKey.String("dotnet")
	// erlang
	// Stability: stable

	TelemetrySDKLanguageErlang = TelemetrySDKLanguageKey.String("erlang")
	// java
	// Stability: stable

	TelemetrySDKLanguageJava = TelemetrySDKLanguageKey.String("java")
	// nodejs
	// Stability: stable

	TelemetrySDKLanguageNodejs = TelemetrySDKLanguageKey.String("nodejs")
	// php
	// Stability: stable

	TelemetrySDKLanguagePHP = TelemetrySDKLanguageKey.String("php")
	// python
	// Stability: stable

	TelemetrySDKLanguagePython = TelemetrySDKLanguageKey.String("python")
	// ruby
	// Stability: stable

	TelemetrySDKLanguageRuby = TelemetrySDKLanguageKey.String("ruby")
	// rust
	// Stability: stable

	TelemetrySDKLanguageRust = TelemetrySDKLanguageKey.String("rust")
	// swift
	// Stability: stable

	TelemetrySDKLanguageSwift = TelemetrySDKLanguageKey.String("swift")
	// webjs
	// Stability: stable

	TelemetrySDKLanguageWebjs = TelemetrySDKLanguageKey.String("webjs")
)

// Namespace: tps
const (

	// trpc 默认染色 key
	// Stability: Development
	// Type: string
	//
	// Examples: ""
	// Note: 伽利略支持任意key染色, 这里的tps.dyeing是平台无关的一个默认实现

	TpsDyeingKey = attribute.Key("tps.dyeing")
)

// Namespace: trace
const (

	// 用户可以使用该 key 作强制采样，供工具调试使用
	// Stability: Development
	// Type: string
	//
	// Examples: "todo"

	TraceForceSampleKey = attribute.Key("trace.force.sample")
)

// Namespace: workflow
const (

	// path 下一跳 hex 值
	// Stability: Development
	// Type: string
	//
	// Examples: "94e9a0e2d86a1686b123eb7e"

	WorkflowChildPathKey = attribute.Key("workflow.child_path")
	// path hex 值
	// Stability: Development
	// Type: string
	//
	// Examples: "94e9a0e2d86a1686b123eb7e"

	WorkflowPathKey = attribute.Key("workflow.path")
)
