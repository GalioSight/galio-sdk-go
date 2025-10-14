// Copyright 2021 Tencent Galileo Authors
//
// Copyright 2021 Tencent OpenTelemetry Oteam
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

// Package main  示例项目的主函数。
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"time"

	sdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"galiosight.ai/galio-sdk-go/components"
	"galiosight.ai/galio-sdk-go/self"

	"galiosight.ai/galio-sdk-go"
	logconf "galiosight.ai/galio-sdk-go/configs/logs"
	metriconf "galiosight.ai/galio-sdk-go/configs/metrics"
	"galiosight.ai/galio-sdk-go/configs/ocp"
	"galiosight.ai/galio-sdk-go/configs/profiles"
	traceconf "galiosight.ai/galio-sdk-go/configs/traces"
	"galiosight.ai/galio-sdk-go/exporters/otlp/traces"
	"galiosight.ai/galio-sdk-go/helper"
	"galiosight.ai/galio-sdk-go/lib/flowtag"
	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/semconv"
	modelv3 "galiosight.ai/galio-sdk-go/v3/model"
)

var (
	tracer  trace.Tracer
	metrics components.MetricsProcessor
	logger  *zap.Logger
)

// 伽利略数据上报示例。
func main() {
	// 资源描述，见文档 https://galiosight.ai/semantic-conventions/blob/toraxie-omp-3.0/semconv/doc/v3.0.0/index.md
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
	// 本地配置
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
		// to.OcpAddr =
		// 数据接入点：中国大陆内网（默认）
		to.Config.AccessPoint = model.AccessPoint_ACCESS_POINT_CN_PRIVATE
		// APIKey 用于鉴权，避免别人伪造数据上报，此 APIKey 需要找伽利略小助手申请请
		to.APIKey = ""
		// 修改 trace 采样率
		to.Config.TracesConfig.Processor.Sampler.Fraction = 1.0
		// 是否开启 runtime 分析指标上报（默认开启）
		to.Config.MetricsConfig.Processor.EnableProcessMetrics = true
		return nil
	}
	setupTelemetry(resv3, local)
	go Run() // 若看不到指标请检查网络环境，建议在 IDC 环境测试。因上报服务在 DevNet、办公网不能直连 IDC 非 80、443、8080 端口。
	go eventsDemo(resv3)
	go profilesDemo(resv3)
	time.Sleep(time.Hour)
}

func setupTelemetry(resv3 *model.Resource, local func(to *ocp.GalileoConfig) error) {
	// 初始化 Ocp 远程配置，每分钟拉取 ocp 配置，进行配置热更新
	_ = ocp.RegisterResource(
		resv3, ocp.WithLocalDecoder(ocp.DecodeFunc(local)),
		ocp.WithDuration(time.Minute),
	)
	// 初始化自监控上报，使用海外接入等非默认配置场景需要设置
	config := ocp.GetUpdater(resv3.Target).GetConfig().Config
	self.SetupObserver(resv3, logs.DefaultWrapper(), config.SelfMonitor, config.ConfigServer)

	var err error
	{ // 构造 tracer, 初始化，只能执行一次
		tracesConfig := traceconf.NewConfig(
			resv3,
			traceconf.WithSchemaURL(semconv.SchemaURL), // 声明 OMP v3 版本协议
		)
		exporter, err := galio.NewTracesExporter(tracesConfig) // 全局持有，不要重复创建。
		if err != nil {
			panic(fmt.Errorf("GetTracesExporter err=%v, tracesConfig=%+v", err, tracesConfig))
		}
		galio.SetDefaultTracesExporter(exporter) // 支持 galio.WithSpan 等 API 使用
		tracer = exporter
	}

	{ // 构造 metric, 初始化，只能执行一次
		metricConfig := metriconf.NewConfig(resv3, metriconf.WithSchemaURL(semconv.SchemaURL))
		// 获取指标处理器，注意 base SDK 是没有热更新的，配置在初始化时就确定了，后面不会再修改。
		// 在伽利略平台上修改配置，对已经初始化好的 processor 是没有用的；建议若在平台修改完配置，等待几分钟缓存清空后再重启服务
		metrics, err = helper.GetMetricsProcessor(metricConfig) // 全局持有，不要重复创建。
		if err != nil {
			panic(err)
		}
	}

	{ // 构造 logger, 初始化，只能执行一次
		logsConfig := logconf.NewConfig(resv3, logconf.WithSchemaURL(semconv.SchemaURL))
		logsConfig.Processor.Level = "INFO"
		logger, err = galio.NewLogger(logsConfig) // 全局持有，不要重复创建。
		if err != nil {
			panic(fmt.Errorf("GetTracesExporter err=%v, logsConfig=%+v", err, logsConfig))
		}
	}
}

type chain []func(ctx context.Context, next func(context.Context) error) error

// Run ...
func (c chain) Run(ctx context.Context) error {
	if len(c) == 1 {
		return c[0](ctx, nil)
	}
	f := c[0]
	left := c[1:]
	return f(ctx, left.Run)
}

// Run ...
func Run() {
	for {
		clientMethod(context.Background())
		time.Sleep(time.Second)
	}
}

// generateTraceID TraceID 应该由上游传过来，此为显式模拟父 span 故使用了随机生成。实际中 otel.Tracer Start() 自动生成 TraceID 无需关注
func generateTraceID() [16]byte {
	var traceID [16]byte
	_, err := rand.Read(traceID[:])
	if err != nil {
		// 处理错误
		panic(err)
	}
	return traceID
}

func profilesDemo(resource *model.Resource) {
	profilesConfig := profiles.NewConfig(resource)
	profilesConfig.Log.Level = logs.LevelDebug
	processor, err := helper.GetProfilesProcessor(profilesConfig)
	if err != nil {
		fmt.Printf("GetProfilesProcessor err=%v, profilesConfig=%+v", err, profilesConfig)
		return
	}
	processor.Start()
}

// reportCustomMetrics 上报自定义监控。
func reportCustomMetrics() {
	monitorName := "monitorItem"                  // 监控项名称
	customMetrics := model.GetCustomMetrics(1, 2) // 一个维度，两个数据点
	defer model.PutCustomMetrics(customMetrics)
	customMetrics.CustomLabels[0].Name = "test_label_name"
	customMetrics.CustomLabels[0].Value = "test_label_value"
	customMetrics.Metrics[0].Name = "a_counter"
	customMetrics.Metrics[0].Aggregation = model.Aggregation_AGGREGATION_COUNTER
	customMetrics.Metrics[0].Value = 100
	customMetrics.Metrics[1].Name = "a_histogram"
	// 注意若在伽利略平台改完分位值配置后，需要等待几分钟重启服务才能生效，因 baseSDK 没有热加载功能
	customMetrics.Metrics[1].Aggregation = model.Aggregation_AGGREGATION_HISTOGRAM
	customMetrics.Metrics[1].Value = generateRandomFloat() // 模拟上报随机数体现分位值效果
	customMetrics.MonitorName = monitorName
	metrics.ProcessCustomMetrics(customMetrics)
}

var rd = rand.New(rand.NewSource(time.Now().UnixNano()))

// generateRandomFloat 返回 [0.0, 1.0] 之间随机数
func generateRandomFloat() float64 {
	return rd.Float64()
}

func rpcLabels(isServer bool, codeType string) []model.RPCLabels_Field {
	// 模调 schema 见文档：https://galiosight.ai/semantic-conventions/blob/master/semconv/doc/v3.0.0/index.md
	fields := []model.RPCLabels_Field{
		modelv3.RPCLabelsField(modelv3.RPCLabels_rpc_callee_method, "DemoCalleeMethod"),
		modelv3.RPCLabelsField(modelv3.RPCLabels_rpc_callee_server, "DemoCalleeServer"),
		modelv3.RPCLabelsField(modelv3.RPCLabels_rpc_callee_service, "DemoCalleeService"),
		modelv3.RPCLabelsField(modelv3.RPCLabels_rpc_caller_method, "DemoCallerMethod"),
		modelv3.RPCLabelsField(modelv3.RPCLabels_rpc_caller_server, "DemoCallerServer"),
		modelv3.RPCLabelsField(modelv3.RPCLabels_rpc_caller_service, "DemoCallerService"),
		modelv3.RPCLabelsField(modelv3.RPCLabels_rpc_error_code, "0"),
		modelv3.RPCLabelsField(modelv3.RPCLabels_rpc_error_code_type, codeType),
		modelv3.RPCLabelsField(modelv3.RPCLabels_rpc_callee_set, "set.sz1.abc1"),
		modelv3.RPCLabelsField(modelv3.RPCLabels_rpc_caller_set, "set.gz1.4s8g"),
		modelv3.RPCLabelsField(modelv3.RPCLabels_rpc_caller_group, ""),
		modelv3.RPCLabelsField(modelv3.RPCLabels_rpc_canary, ""),
		modelv3.RPCLabelsField(modelv3.RPCLabels_rpc_user_ext1, ""),
		modelv3.RPCLabelsField(modelv3.RPCLabels_rpc_user_ext2, ""),
		modelv3.RPCLabelsField(modelv3.RPCLabels_rpc_user_ext3, ""),
		// 此字段正常应该是空，当需要给流量打特殊标签时，才需要填充值，此处演示 flowtag 的用法
		modelv3.RPCLabelsField(modelv3.RPCLabels_rpc_flow_tag, (flowtag.Downgrade | flowtag.Gray).String()),
	}
	if isServer {
		fields = append(fields, modelv3.RPCLabelsField(modelv3.RPCLabels_rpc_caller_ip, "127.0.0.1"))
		fields = append(
			fields, modelv3.RPCLabelsField(
				modelv3.RPCLabels_rpc_caller_container,
				"test.galileo.sdkserver.sz10010",
			),
		)
		// 另外一边 ip 由 resource 的 host.ip 填充
	} else {
		fields = append(fields, modelv3.RPCLabelsField(modelv3.RPCLabels_rpc_callee_ip, "127.0.0.1"))
		fields = append(
			fields, modelv3.RPCLabelsField(
				modelv3.RPCLabels_rpc_callee_container,
				"test.example.demoserver.sz10020",
			),
		)
		// 另外一边 ip 由 resource 的 host.ip 填充
	}
	return fields
}

func printSpan(span trace.Span) {
	rspan := span.(traces.Span)
	kind := "server"
	if rspan.SpanKind() == trace.SpanKindClient {
		kind = "client"
	}
	for _, attr := range rspan.Attributes() {
		log.Printf("%s span attribute %s %s", kind, attr.Key, attr.Value.AsString())
	}
	for _, attr := range rspan.Resource().Attributes() {
		log.Printf("%s resource span attribute %s %s", kind, attr.Key, attr.Value.AsString())
	}
}

// RecoveryHandler 伽利略 recovery handler，用于捕获 panic 事件。
func RecoveryHandler(e interface{}, labels ...interface{}) {
	const PanicBufLen = 1024 * 1024
	buf := make([]byte, PanicBufLen)
	buf = buf[:runtime.Stack(buf, false)]
	stackMsg := fmt.Sprintf("[PANIC][GALILEO][%v]%v\n%s\n", labels, e, buf)
	galio.ReportEvent(stackMsg, "recovery", "runtime", "panic") // 上报时间
}

// Recover 捕获 panic
func Recover(labels ...interface{}) {
	if err := recover(); err != nil {
		RecoveryHandler(err, labels...)
	}
}

func init() {
	// 如果需要强制指定 SpanID，才可以覆盖 IDGenerator
	galio.WithTracerProviderOptions(sdk.WithIDGenerator(galio.NewSpanIDInjector(nil)))
}
