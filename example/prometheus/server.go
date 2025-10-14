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
)

// reportServerMetrics 上报服务端 (被调) 监控。
func reportServerMetrics(ctx context.Context, next func(context.Context) error) error {
	startTime := time.Now()
	err := next(ctx)
	endTime := time.Now()
	codeType := getCodeType(err) // 判断成功、失败、异常
	values := rpcValues("rpc_server", codeType)
	RPCServerHandledTotal.WithLabelValues(values...).Inc()
	RPCServerHandledSeconds.WithLabelValues(values...).Observe(endTime.Sub(startTime).Seconds())
	return err
}

func serverMethod(ctx context.Context) error {
	return chain{reportServerMetrics, serverHandle}.Run(ctx)
}

func reportCustomMetrics() error {
	CustomCounter.WithLabelValues("bar").Inc()
	CustomGauge.WithLabelValues("bar").Set(1.0)
	CustomHistogram.WithLabelValues("bar").Observe(0.5)
	return nil
}

func serverHandle(ctx context.Context, next func(ctx context.Context) error) error {
	err := reportCustomMetrics()
	time.Sleep(time.Second)
	return err
}
