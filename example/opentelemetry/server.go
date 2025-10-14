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
	"time"

	"go.opentelemetry.io/otel/metric"
)

// reportServerMetrics 上报服务端 (被调) 监控。
func reportServerMetrics(ctx context.Context, next func(context.Context)) {
	values := rpcValues()
	next(ctx)
	RPCServerHandledTotal.Add(ctx, 1, metric.WithAttributes(values...))
	RPCServerHandledSeconds.Record(ctx, 1, metric.WithAttributes(values...))
}

func serverMethod(ctx context.Context) {
	chain{reportServerMetrics, serverHandle}.Run(ctx)
}

func reportCustomMetrics(ctx context.Context) {
	CustomCounter.Add(ctx, 1, metric.WithAttributes(customs))
	CustomGauge.Record(ctx, 1.0, metric.WithAttributes(customs))
	CustomHistogram.Record(ctx, 0.5, metric.WithAttributes(customs))
}

func serverHandle(ctx context.Context, next func(ctx context.Context)) {
	reportCustomMetrics(ctx)
	time.Sleep(time.Second)
}
