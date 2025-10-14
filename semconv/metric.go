// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Code generated from semantic convention specification. DO NOT EDIT.

package semconv // import "go.opentelemetry.io/otel/semconv/"

const (

	// GenAIRequestCnt is the metric conforming to the "gen_ai.request_cnt" semantic conventions. It represents the 请求量
	// Instrument: counter
	// Unit: {request}
	// Stability:

	GenAIRequestCntName = "gen_ai.request_cnt"

	GenAIRequestCntUnit = "{request}"

	GenAIRequestCntDescription = "请求量"

	// GenAIUsageInputTokens is the metric conforming to the "gen_ai.usage.input_tokens" semantic conventions. It represents the llm usage prompts token 指标
	// Instrument: histogram
	// Unit: {token}
	// Stability:

	GenAIUsageInputTokensName = "gen_ai.usage.input_tokens"

	GenAIUsageInputTokensUnit = "{token}"

	GenAIUsageInputTokensDescription = "llm usage prompts token 指标"

	// GenAIUsageOutputTokens is the metric conforming to the "gen_ai.usage.output_tokens" semantic conventions. It represents the llm usage completion token 指标
	// Instrument: histogram
	// Unit: {token}
	// Stability:

	GenAIUsageOutputTokensName = "gen_ai.usage.output_tokens"

	GenAIUsageOutputTokensUnit = "{token}"

	GenAIUsageOutputTokensDescription = "llm usage completion token 指标"

	// GenAIClientOperationDuration is the metric conforming to the "gen_ai.client.operation.duration" semantic conventions. It represents the 请求耗时
	// Instrument: histogram
	// Unit: s
	// Stability:

	GenAIClientOperationDurationName = "gen_ai.client.operation.duration"

	GenAIClientOperationDurationUnit = "s"

	GenAIClientOperationDurationDescription = "请求耗时"

	// GenAIServerTimePerOutputToken is the metric conforming to the "gen_ai.server.time_per_output_token" semantic conventions. It represents the 每秒 token 数
	// Instrument: histogram
	// Unit: {token}
	// Stability:

	GenAIServerTimePerOutputTokenName = "gen_ai.server.time_per_output_token"

	GenAIServerTimePerOutputTokenUnit = "{token}"

	GenAIServerTimePerOutputTokenDescription = "每秒 token 数"

	// GenAIServerTimeToFirstToken is the metric conforming to the "gen_ai.server.time_to_first_token" semantic conventions. It represents the 首 token 耗时
	// Instrument: histogram
	// Unit: s
	// Stability:

	GenAIServerTimeToFirstTokenName = "gen_ai.server.time_to_first_token"

	GenAIServerTimeToFirstTokenUnit = "s"

	GenAIServerTimeToFirstTokenDescription = "首 token 耗时"

	// GenAIServerEvents is the metric conforming to the "gen_ai.server.events" semantic conventions. It represents the 流消息事件数
	// Instrument: histogram
	// Unit: {events}
	// Stability:

	GenAIServerEventsName = "gen_ai.server.events"

	GenAIServerEventsUnit = "{events}"

	GenAIServerEventsDescription = "流消息事件数"

	// RPCServerHandledSeconds is the metric conforming to the "rpc_server_handled_seconds" semantic conventions. It represents the measures the duration of inbound RPC
	// Instrument: histogram
	// Unit: s
	// Stability: development

	RPCServerHandledSecondsName = "rpc_server_handled_seconds"

	RPCServerHandledSecondsUnit = "s"

	RPCServerHandledSecondsDescription = "Measures the duration of inbound RPC."

	// RPCServerStartedTotal is the metric conforming to the "rpc_server_started_total" semantic conventions. It represents the 服务端（被调方上报）接收到的请求量
	// Instrument: counter
	// Unit: {count}
	// Stability: development

	RPCServerStartedTotalName = "rpc_server_started_total"

	RPCServerStartedTotalUnit = "{count}"

	RPCServerStartedTotalDescription = "服务端（被调方上报）接收到的请求量"

	// RPCServerHandledTotal is the metric conforming to the "rpc_server_handled_total" semantic conventions. It represents the 服务端（被调方上报）处理完成的请求量
	// Instrument: counter
	// Unit: {count}
	// Stability: development

	RPCServerHandledTotalName = "rpc_server_handled_total"

	RPCServerHandledTotalUnit = "{count}"

	RPCServerHandledTotalDescription = "服务端（被调方上报）处理完成的请求量"

	// RPCClientHandledSeconds is the metric conforming to the "rpc_client_handled_seconds" semantic conventions. It represents the measures the duration of outbound RPC
	// Instrument: histogram
	// Unit: s
	// Stability: development

	RPCClientHandledSecondsName = "rpc_client_handled_seconds"

	RPCClientHandledSecondsUnit = "s"

	RPCClientHandledSecondsDescription = "Measures the duration of outbound RPC."

	// RPCClientStartedTotal is the metric conforming to the "rpc_client_started_total" semantic conventions. It represents the 客户端（主调方上报）发出的请求量
	// Instrument: counter
	// Unit: {count}
	// Stability: development

	RPCClientStartedTotalName = "rpc_client_started_total"

	RPCClientStartedTotalUnit = "{count}"

	RPCClientStartedTotalDescription = "客户端（主调方上报）发出的请求量"

	// RPCClientHandledTotal is the metric conforming to the "rpc_client_handled_total" semantic conventions. It represents the 客户端（主调方上报）处理完成的请求量
	// Instrument: counter
	// Unit: {count}
	// Stability: development

	RPCClientHandledTotalName = "rpc_client_handled_total"

	RPCClientHandledTotalUnit = "{count}"

	RPCClientHandledTotalDescription = "客户端（主调方上报）处理完成的请求量"
)
