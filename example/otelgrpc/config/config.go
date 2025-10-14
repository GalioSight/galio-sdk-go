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

// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package config ...
package config // import "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc/example/config"

import (
	"fmt"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"galiosight.ai/galio-sdk-go"
	"galiosight.ai/galio-sdk-go/configs/ocp"
	traceconf "galiosight.ai/galio-sdk-go/configs/traces"
	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/self"
	modelv3 "galiosight.ai/galio-sdk-go/v3/model"
)

// Init configures an OpenTelemetry exporter and trace provider.
func Init() (*sdktrace.TracerProvider, error) {

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

	local := func(to *ocp.GalileoConfig) error {
		// 实际情况可以用
		// yaml.Unmarshal(cfg, &to.Config)
		// 如果上报数据到伽利略失败，需要调试，可以修改成 info 或 debug
		// 日志会输出到 ./galileo/galileo.log 中
		to.Verbose = "error"
		// 接入地址请根据需要修改，参考：
		// ocp 管控地址：中国大陆内网（默认）
		to.OcpAddr = ""
		// 测试环境
		// to.OcpAddr = ""
		// 数据接入点：中国大陆内网（默认）
		to.Config.AccessPoint = model.AccessPoint_ACCESS_POINT_CN_PRIVATE

		// 修改 trace 采样率
		to.Config.TracesConfig.Processor.Sampler.Fraction = 1.0
		// 是否开启 runtime 分析指标上报（默认开启）
		to.Config.MetricsConfig.Processor.EnableProcessMetrics = true
		return nil
	}
	// 初始化 Ocp 远程配置，每分钟拉取 ocp 配置，进行配置热更新
	_ = ocp.RegisterResource(
		resv3, ocp.WithLocalDecoder(ocp.DecodeFunc(local)),
		ocp.WithDuration(time.Minute),
	)
	// 初始化自监控上报，使用海外接入等非默认配置场景需要设置
	config := ocp.GetUpdater(resv3.Target).GetConfig().Config
	self.SetupObserver(resv3, logs.DefaultWrapper(), config.SelfMonitor, config.ConfigServer)

	{ // 构造 tracer, 初始化，只能执行一次
		tracesConfig := traceconf.NewConfig(resv3,
			traceconf.WithSchemaURL(modelv3.OtelSchemaURL), // 声明 OpenTelemetry 版本协议
		)
		// XXX 这里副作用会调用 otel.SetTraceProvider
		exporter, err := galio.NewTracesExporter(tracesConfig) // 全局持有，不要重复创建。
		if err != nil {
			panic(fmt.Errorf("GetTracesExporter err=%v, tracesConfig=%+v", err, tracesConfig))
		}
		galio.SetDefaultTracesExporter(exporter) // 支持 galio.WithSpan 等 API 使用
	}

	return nil, nil
}
