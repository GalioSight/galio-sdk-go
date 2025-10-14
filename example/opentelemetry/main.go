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

package main

import (
	"context"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus/push"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"galiosight.ai/galio-sdk-go/exporters/otlp/metrics"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/semconv"
	modelv3 "galiosight.ai/galio-sdk-go/v3/model"
)

var (
	pusher *push.Pusher
	P      = semconv.PrometheusLabel
	V      = semconv.PrometheusValue
)

// PrometheusPush 将 Prometheus 指标通过 push 方式进行上报。
// 此函数主要用于兼容天机阁等使用 Prometheus API 的场景。
// 如果推送成功，将返回一个取消函数和 nil；如果失败，将返回相应的错误。
func setup() (context.CancelFunc, error) {
	resv3 := modelv3.NewResource(
		"STKE.example.omp_v3_go",
		model.Production,
		"formal",
		"127.0.0.1",
		"test.galileo.SDK.sz10010",
		"set.sz.1",
		"sz",
		"test-v0.1.0",
		"",
	)

	cfg := model.OpenTelemetryPushConfig{
		Enable: true,
		Url:    "", // 伽利略 OpenTelemetry collector 地址。
	}
	// Galileo SDK 提供 helper 帮助构建 provider，用户不想依赖 galileo SDK 也可以将代码复制出去
	provider, err := metrics.NewMeterProvider(*resv3, cfg)
	if err != nil {
		return nil, err
	}

	otel.SetMeterProvider(provider)
	doInit()
	return func() {
		provider.Shutdown(context.Background())
	}, nil
}

func main() {
	cancel, err := setup()
	if err != nil {
		log.Fatalf("PrometheusPush() error = %v", err)
	}
	defer cancel()

	go Run()
	time.Sleep(time.Hour)
}

// Run ...
func Run() {
	for {
		clientMethod(context.Background())
		time.Sleep(time.Second)
	}
}

var (
	RPCClientHandledSeconds metric.Float64Histogram
	RPCClientHandledTotal   metric.Int64Counter
	RPCServerHandledSeconds metric.Float64Histogram
	RPCServerHandledTotal   metric.Int64Counter
	CustomCounter           metric.Int64Counter
	CustomGauge             metric.Float64Gauge
	CustomHistogram         metric.Float64Histogram
	customs                 = attribute.String("foo", "bar")
)

func doInit() {
	// 新建主调监控项，在伽利略 otlp 协议主调监控项名固定为 client_metrics。
	rpcClient := otel.Meter("client_metrics")
	RPCClientHandledTotal, _ = rpcClient.Int64Counter(
		semconv.RPCClientHandledTotalName,
		metric.WithDescription(semconv.RPCClientHandledTotalDescription),
		metric.WithUnit(semconv.RPCClientHandledTotalUnit),
	)
	RPCClientHandledSeconds, _ = rpcClient.Float64Histogram(
		semconv.RPCClientHandledSecondsName,
		metric.WithDescription(semconv.RPCClientHandledSecondsDescription),
		metric.WithUnit(semconv.RPCClientHandledSecondsUnit),
	)
	// 新建被调监控项，在伽利略 otlp 协议被调监控项名固定为 server_metrics。
	rpcServer := otel.Meter("server_metrics")
	RPCServerHandledTotal, _ = rpcServer.Int64Counter(
		semconv.RPCServerHandledTotalName,
		metric.WithDescription(semconv.RPCServerHandledTotalDescription),
		metric.WithUnit(semconv.RPCServerHandledTotalUnit),
	)
	RPCServerHandledSeconds, _ = rpcServer.Float64Histogram(
		semconv.RPCServerHandledSecondsName,
		metric.WithDescription(semconv.RPCServerHandledSecondsDescription),
		metric.WithUnit(semconv.RPCServerHandledSecondsUnit),
	)

	// 新建自定义监控项 demo_monitor，自定义监控项名用户可以任意指定。
	demoMonitor := otel.Meter("demo_monitor")
	CustomCounter, _ = demoMonitor.Int64Counter(
		"custom_counter_total", metric.WithDescription("custom counter demo"), metric.WithUnit("{count}"),
	)
	CustomGauge, _ = demoMonitor.Float64Gauge(
		"custom_gauge_set", metric.WithDescription("custom gauge demo"), metric.WithUnit("ms"),
	)
	CustomHistogram, _ = demoMonitor.Float64Histogram(
		"custom_histogram", metric.WithDescription("custom histogram demo"), metric.WithUnit("ms"),
	)
}

type chain []func(ctx context.Context, next func(context.Context))

func (c chain) Run(ctx context.Context) {
	if len(c) == 1 {
		c[0](ctx, nil)
		return
	}
	f := c[0]
	left := c[1:]
	f(ctx, left.Run)
}

func rpcValues() []attribute.KeyValue {
	return []attribute.KeyValue{
		semconv.RPCCallerMethodKey.String("DemoCallerMethod"),
		semconv.RPCCallerServerKey.String("DemoCallerServer"),
		semconv.RPCCallerServiceKey.String("DemoCallerService"),
		semconv.RPCCallerContainerKey.String("test.galileo.sdkserver.sz10010"),
		semconv.RPCCallerIPKey.String("127.0.0.1"),
		semconv.RPCCallerSetKey.String("set.gz1.4s8g"),
		semconv.RPCCalleeMethodKey.String("DemoCalleeMethod"),
		semconv.RPCCalleeServerKey.String("DemoCalleeServer"),
		semconv.RPCCalleeServiceKey.String("DemoCalleeService"),
		semconv.RPCCalleeContainerKey.String("test.galileo.sdkserver.sz10010"),
		semconv.RPCCalleeIPKey.String("127.0.0.1"),
		semconv.RPCCalleeSetKey.String("set.sz1.abc1"),
		semconv.RPCErrorCodeKey.String("0"),
		semconv.RPCErrorCodeTypeSuccess,
		semconv.RPCCallerGroupKey.String(""),
		semconv.RPCCanaryKey.String(""),
		semconv.RPCUserExt1Key.String(""),
		semconv.RPCUserExt2Key.String(""),
		semconv.RPCUserExt3Key.String(""),
	}
}
