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

	"go.opentelemetry.io/otel/metric"
)

// reportClientMetrics 上报客户端侧 (主调) 监控。
func reportClientMetrics(ctx context.Context, next func(context.Context)) {
	values := rpcValues()
	next(ctx)
	RPCClientHandledTotal.Add(ctx, 1, metric.WithAttributes(values...))
	RPCClientHandledSeconds.Record(ctx, 0.5, metric.WithAttributes(values...))
}

func clientMethod(ctx context.Context) {
	chain{reportClientMetrics, callServer}.Run(ctx)
}

func callServer(ctx context.Context, next func(context.Context)) {
	// 实际情况中，client 和 server 不在同一个进程，所以需要重置 ctx
	serverMethod(context.Background())
}
