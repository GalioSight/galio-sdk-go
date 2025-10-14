// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0
//
// 2024 Tencent Galileo Authors

package main

import (
	"context"
	"errors"
	stdlog "log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	oteltrace "go.opentelemetry.io/otel/trace"

	logconf "galiosight.ai/galio-sdk-go/configs/logs"
	"galiosight.ai/galio-sdk-go/configs/ocp"
	traceconf "galiosight.ai/galio-sdk-go/configs/traces"
	expres "galiosight.ai/galio-sdk-go/exporters/otlp/resource"
	"galiosight.ai/galio-sdk-go/helper"
	"galiosight.ai/galio-sdk-go/model"
	semconv "galiosight.ai/galio-sdk-go/semconv"
)

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func setupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := newTraceProvider()
	if err != nil {
		stdlog.Fatal(err)
		return
	}
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, _ := newMeterProvider()
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Set up logger provider.
	loggerProvider, err := newLoggerProvider()
	if err != nil {
		stdlog.Fatal(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func initResource() *model.Resource {
	resource := model.NewResource(
		"Galileo-Dial", // 资源所在平台，如 PCG-123、STKE
		"galileo",      // 应用名
		"SDK",          // 服务名
		"DemoService",
		model.Production,           // 物理环境，只能在 model.Production 和 model.Development 枚举，正式环境必须是 model.Production
		"formal",                   // 用户环境，一般是 formal 和 test (或形如 3c170118 等自定义), 正式环境必须是 formal
		"set.sz.1",                 // set 名，可以为空
		"sz",                       // 城市，可以为空
		"127.0.0.1",                // 实例 IP，可以为空
		"test.galileo.SDK.sz10010", // 容器名，可以为空
	)
	if ocp.GetUpdater(resource.Target) != nil {
		// 已经初始化过了
		return resource
	}
	local := func(to *ocp.GalileoConfig) error {
		// 实际情况可以用
		// yaml.Unmarshal(cfg, &to.Config)
		to.Config.TracesConfig.Processor.Sampler.Fraction = 1
		to.Config.TracesConfig.Exporter.WindowSeconds = 1 // 测试用
		to.Config.LogsConfig.Exporter.WindowSeconds = 1   // 测试用
		return nil
	}
	// 初始化 Ocp 远程配置，每分钟拉取 ocp 配置，进行配置热更新
	ocp.RegisterResource(
		resource, ocp.WithLocalDecoder(ocp.DecodeFunc(local)),
		ocp.WithDuration(time.Minute),
	)
	return resource
}

func newTraceProvider() (oteltrace.TracerProvider, error) {
	// 改造代码开始，参考 example/simple 构造出 galileo SDK tracer
	tracesConfig := traceconf.NewConfig(initResource(),
		traceconf.WithSchemaURL(semconv.SchemaURL),
	)
	helper.GetTracesExporter(tracesConfig) // Galileo 创建 tracer 后会注册到全局 Provider 中，后续按照 otel 官方用法即可。
	return otel.GetTracerProvider(), nil
	// 改造结束
}

func newMeterProvider() (*metric.MeterProvider, error) {
	metricExporter, _ := stdoutmetric.New()
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			// Default is 1m. Set to 3s for demonstrative purposes.
			metric.WithInterval(3*time.Second))),
	)
	return meterProvider, nil
}

func newLoggerProvider() (*log.LoggerProvider, error) {
	res := initResource()

	logConfig := logconf.NewConfig(res, logconf.WithSchemaURL(""))
	// galileo 的 log 做得比较早，当时 otel 的 logger provider 还没有出来，所以没
	// 有办法方便的构造出来，等后续 galileo 代码迁移成 otel 的 log provider, 先手
	// 动构造一个

	exporter, err := otlploghttp.New(context.Background(),
		otlploghttp.WithEndpoint(logConfig.Exporter.GetCollector().Addr),
		otlploghttp.WithInsecure(),
		otlploghttp.WithCompression(otlploghttp.GzipCompression),
		otlploghttp.WithHeaders(map[string]string{
			model.TargetHeaderKey: res.GetTarget(),
		}))
	if err != nil {
		return nil, err
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(exporter)), // log.WithExportInterval(time.Second), // 测试用，正式场景请去掉
		log.WithResource(expres.GenResource(semconv.SchemaURL, res)),
	)
	return loggerProvider, nil
}
