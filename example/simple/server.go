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

// Package main ...
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"galiosight.ai/galio-sdk-go"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/semconv"
)

func newSpanContextDemo(ctx context.Context) context.Context {
	sc := trace.NewSpanContext(
		trace.SpanContextConfig{
			TraceID:    generateTraceID(),
			SpanID:     trace.SpanID([8]byte{255}),
			TraceFlags: trace.FlagsSampled, // 假设上游采样了
		},
	)
	// 注入 上游 trace id, span id
	// 之后 tracer.Start 时，继承 trace id, span id 还是会新生成
	return trace.ContextWithRemoteSpanContext(ctx, sc)
}

func injectSpanIDDemo(ctx context.Context) context.Context {
	// 如果需要强制指定 SpanID，需要在 init 函数里面覆盖 IDGenerator
	// galio.WithTracerProviderOptions(sdk.WithIDGenerator(galio.NewSpanIDInjector(nil)))
	return galio.InjectSpanID(newSpanContextDemo(ctx), trace.SpanID([8]byte{1}))
}

func tracesServerDemo(ctx context.Context, next func(context.Context) error) error {
	// 这里假设从上游获得了请求，再从 header 中还原出来 SpanContext
	req := ctx.Value("req").(*http.Request)
	ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(req.Header))
	// 如果需要完全新建 SpanContext, 可以使用下面的 demo
	_ = newSpanContextDemo(ctx)
	// 如果需要强制指定 span id，可以使用下面的 demo
	_ = injectSpanIDDemo(ctx)

	// 将 ctx 传给 tracer，创建新的 span，用于将父子 span 串联起来。
	ctx, span := tracer.Start(ctx, "server",
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			// 按照 OMP v3 规范上报字段
			semconv.RPCCalleeMethodKey.String("DemoCalleeMethod"),
			semconv.RPCCalleeServerKey.String("DemoCalleeServer"),
			semconv.RPCCalleeServiceKey.String("DemoCalleeService"),
			semconv.RPCCallerMethodKey.String("DemoCallerMethod"),
			semconv.RPCCallerServerKey.String("DemoCallerServer"),
			semconv.RPCCallerServiceKey.String("DemoCallerService"),
			semconv.RPCCallerIPKey.String("127.0.0.1"),
			semconv.RPCCalleeIPKey.String("127.0.0.1"),
			attribute.String("foo-demo", "bar"), // 用户可增加自定义属性 foo:bar
		),
	)
	defer span.End()
	err := next(ctx)
	span.SetAttributes(semconv.RPCErrorCodeKey.String("0"))
	span.SetAttributes(semconv.RPCErrorCodeTypeSuccess)
	span.SetAttributes(semconv.RPCErrorMessageKey.String("no promblem"))
	// 用户可增加自定义属性 foo:bar
	span.SetAttributes(attribute.String("foo", "bar"))
	span.SetStatus(codes.Ok, "") // 标识是否发生错误。枚举值：codes.Ok codes.Error codes.Unset ,对应组合查询页面 调用状态的 ok, Error, unSet
	span.AddEvent("SENT", trace.WithAttributes(attribute.String("message.detail", `{"req": "test_req"}`)))
	span.AddEvent("RECEIVED", trace.WithAttributes(attribute.String("message.detail", "success")))
	return err
}

// reportServerMetrics 上报服务端 (被调) 监控。
func reportServerMetrics(ctx context.Context, next func(context.Context) error) error {
	startTime := time.Now()
	err := next(ctx)
	endTime := time.Now()

	codeType := getCodeType(err) // 判断成功、失败、异常
	fields := rpcLabels(true, codeType)
	serverMetrics := model.GetServerMetrics(len(fields))
	defer model.PutServerMetrics(serverMetrics)
	serverMetrics.RpcLabels.Fields = fields
	serverMetrics.Metrics[model.ServerMetricHandledTotalPoint].Value = 1
	serverMetrics.Metrics[model.ServerMetricHandledSecondsPoint].Value = endTime.Sub(startTime).Seconds()
	metrics.ProcessServerMetrics(serverMetrics)
	return err
}

func logsDemo(ctx context.Context) {
	logger.Debug(
		"log debug example", zap.String("foo", "bar"),
		zap.String("k2", "v2"), zap.String("k3", "v3"),
	)
	// 这里要确保前面已经运行过 tracer.Start，不然 ctx 里面不包含 traceID 等信息。
	sc := trace.SpanContextFromContext(ctx)
	logger.Info(
		"log info example trace demo", zap.String("foo", "bar"),
		zap.String("k0", "v0"), zap.String("k1", "v1"),
		zap.String("traceID", sc.TraceID().String()),
		zap.String("spanID", sc.SpanID().String()),
		zap.String("sampled", fmt.Sprintf("%v", sc.TraceFlags() == trace.FlagsSampled)),
	)
	logger.Warn(
		"log warn example", zap.String("foo", "bar"),
		zap.String("k0", "v0"), zap.String("k1", "v1"),
	)
	logger.Error(
		"log error example", zap.String("foo", "bar"),
		zap.String("k0", "v0"), zap.String("k1", "v1"),
		zap.String("k5", "v5"),
	)
}

func serverMethod(ctx context.Context) error {
	return chain{reportServerMetrics, tracesServerDemo, serverHandle}.Run(ctx)
}

func serverHandle(ctx context.Context, next func(ctx context.Context) error) error {
	reportCustomMetrics()
	logsDemo(ctx)
	reportInnerSpan(ctx)
	time.Sleep(time.Second)
	return nil
}

func reportInnerSpan(ctx context.Context) {
	_, span := tracer.Start(ctx, "workHard",
		trace.WithAttributes(attribute.String("extra.key", "extra.value")))
	defer span.End()
	time.Sleep(50 * time.Millisecond)
}
