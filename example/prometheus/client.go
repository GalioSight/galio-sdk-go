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

	"galiosight.ai/galio-sdk-go/semconv"
)

// reportClientMetrics 上报客户端侧 (主调) 监控。
func reportClientMetrics(ctx context.Context, next func(context.Context) error) error {
	startTime := time.Now()
	err := next(ctx)
	endTime := time.Now()
	codeType := getCodeType(err) // 判断成功、失败、异常
	values := rpcValues("rpc_client", codeType)
	RPCClientHandledTotal.WithLabelValues(values...).Inc()
	RPCClientHandledSeconds.WithLabelValues(values...).Observe(endTime.Sub(startTime).Seconds())
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
	return chain{reportClientMetrics, callServer}.Run(ctx)
}

func callServer(ctx context.Context, next func(context.Context) error) error {
	// 实际情况中，client 和 server 不在同一个进程，所以需要重置 ctx
	return serverMethod(context.Background())
}
