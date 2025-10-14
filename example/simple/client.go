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
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/semconv"
)

func tracesClientDemo(ctx context.Context, next func(context.Context) error) error {
	ctx, span := tracer.Start(ctx, "client",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.RPCCalleeMethodKey.String("DemoCalleeMethod"),
			semconv.RPCCalleeServerKey.String("DemoCalleeServer"),
			semconv.RPCCalleeServiceKey.String("DemoCalleeService"),
			semconv.RPCCallerMethodKey.String("DemoCallerMethod"),
			semconv.RPCCallerServerKey.String("DemoCallerServer"),
			semconv.RPCCallerServiceKey.String("DemoCallerService"),
			semconv.RPCCallerIPKey.String("127.0.0.1"),
			semconv.RPCCalleeIPKey.String("127.0.0.1"),
		),
	)
	defer span.End()
	// 假设程序是发送的 http 请求，通过 carrier 和 inject 将 ctx 中的 span 注入到 http header 中，即随即发送到后端了。
	req, _ := http.NewRequest("", "", nil)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	// 这里创建一个与上游无关的 ctx，并且将 header 放进去，模拟实际情况中，跨服务的 rpc 调用，span 是通过 carrier 传递而非 ctx 内存传递
	err := next(context.WithValue(context.Background(), "req", req))
	span.SetAttributes(semconv.RPCErrorCodeKey.String("0"))
	span.SetAttributes(semconv.RPCErrorCodeTypeSuccess)
	span.SetAttributes(semconv.RPCErrorMessageKey.String("no promblem"))
	return err
}

// reportClientMetrics 上报客户端侧 (主调) 监控。
func reportClientMetrics(ctx context.Context, next func(context.Context) error) error {
	startTime := time.Now()
	err := next(ctx)
	endTime := time.Now()
	codeType := getCodeType(err) // 判断成功、失败、异常
	fields := rpcLabels(false, codeType)
	clientMetrics := model.GetClientMetrics(len(fields))
	defer model.PutClientMetrics(clientMetrics)
	clientMetrics.RpcLabels.Fields = fields
	clientMetrics.Metrics[model.ClientMetricHandledTotalPoint].Value = 1
	clientMetrics.Metrics[model.ClientMetricHandledSecondsPoint].Value = endTime.Sub(startTime).Seconds()
	metrics.ProcessClientMetrics(clientMetrics)
	return err
}

func getCodeType(err error) string {
	if err == nil {
		return semconv.RPCErrorCodeTypeSuccessValue
	}
	if checkIsTimeout(err) {
		return semconv.RPCErrorCodeTypeTimeoutValue
	} else {
		return semconv.RPCErrorCodeTypeExceptionValue
	}
}

func checkIsTimeout(err error) bool {
	return false
}

func clientMethod(ctx context.Context) error {
	return chain{reportClientMetrics, tracesClientDemo, callServer}.Run(ctx)
}

func callServer(ctx context.Context, next func(context.Context) error) error {
	return serverMethod(ctx)
}
