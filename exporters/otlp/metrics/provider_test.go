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

// Package metrics 用于将 OpenTelemetry 指标通过 push 方式上报到 OpenTelemetry collector
package metrics

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"galiosight.ai/galio-sdk-go/model"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

var (
	testProvider = newTestProvider() // provider，全局唯一。
	// 被调监控查看：
	testServerMeter            = newTestServerMeter(testProvider)               // meter：被调监控，全局唯一。
	testServerHandledCounter   = newTestServerHandledCounter(testServerMeter)   // 指标：被调请求数，全局唯一。
	testServerHandledHistogram = newTestServerHandledHistogram(testServerMeter) // 指标：被调耗时直方图，全局唯一。
	// 主调监控查看：
	testClientMeter            = newTestClientMeter(testProvider)               // meter：主调，全局唯一。
	testClientHandledCounter   = newTestClientHandledCounter(testClientMeter)   // 指标：主调请求数，全局唯一。
	testClientHandledHistogram = newTestClientHandledHistogram(testClientMeter) // 指标：主调耗时直方图，全局唯一。
	// 自定义监控查看：
	xxxMeter   = newCustomXXXMeter(testProvider) // meter：monitor_xxx，全局唯一。
	xxxTotal   = newCustomXXXTotal(xxxMeter)     // 指标：xxx_total，全局唯一。
	xxxSeconds = newCustomXXXSeconds(xxxMeter)   // 指标：xxx_seconds，全局唯一。
	xxxBytes   = newCustomXXXBytes(xxxMeter)     // 指标：xxx_bytes，全局唯一。
)

const (
	serverMeterName = "server_metrics"
	clientMeterName = "client_metrics"
)

const (
	codeSuccess       = "0"
	codeTimeout       = "err_21"
	codeException     = "err_161"
	codeTypeSuccess   = "success"   // 成功
	codeTypeTimeout   = "timeout"   // 超时
	codeTypeException = "exception" // 异常
)

var (
	attrCallerService   = attribute.Key("caller_service")
	attrCallerMethod    = attribute.Key("caller_method")
	attrCallerConSetid  = attribute.Key("caller_con_setid")
	attrCallerIP        = attribute.Key("caller_ip")
	attrCallerContainer = attribute.Key("caller_container")
	attrCalleeService   = attribute.Key("callee_service")
	attrCalleeMethod    = attribute.Key("callee_method")
	attrCalleeConSetid  = attribute.Key("callee_con_setid")
	attrCalleeIP        = attribute.Key("callee_ip")
	attrCalleeContainer = attribute.Key("callee_container")
	attrCallerGroup     = attribute.Key("caller_group")
	attrUserExt1        = attribute.Key("user_ext1")
	attrUserExt2        = attribute.Key("user_ext2")
	attrUserExt3        = attribute.Key("user_ext3")
	attrCallerServer    = attribute.Key("caller_server")
	attrCalleeServer    = attribute.Key("callee_server")
	attrCode            = attribute.Key("code")
	attrCodeType        = attribute.Key("code_type")
)

func newTestProvider() *sdkmetric.MeterProvider {
	res := model.Resource{
		Target:        "RPC.example.galileohttp", // 观测对象的唯一标识 ID，需要全局唯一，如：PCG-123.galileo.metaserver
		Namespace:     "Development",             // 物理环境，如：Development
		EnvName:       "test",                    // 用户环境，如：test
		Instance:      "LocalIP",                 // 实例，如：10.20.30.40
		ContainerName: "ContainerName",           // 容器，如：test.galileo.metaserver.sz100012
		App:           "example",                 // 业务名，如 galileo。
		Server:        "galileohttp",             // 服务名，如 metaserver。
		SetName:       "SetName",                 // 分 set 时的 set 名，如 set.sz.1
		TenantId:      "default",                 // 租户 ID。
	}
	cfg := model.OpenTelemetryPushConfig{
		Enable: true,                // 开启上报。
		Url:    "otlp.j.woa.com:80", // 伽利略 OpenTelemetry collector 地址。
	}
	if provider, err := NewMeterProvider(res, cfg, WithAPIKey("abcd")); err == nil {
		return provider
	}
	return nil
}

func newTestServerMeter(provider *sdkmetric.MeterProvider) otelmetric.Meter {
	serverMeter := provider.Meter(serverMeterName)
	return serverMeter
}

func newTestServerHandledCounter(serverMeter otelmetric.Meter) otelmetric.Int64Counter {
	serverHandledCounter, _ := serverMeter.Int64Counter(
		"rpc_server_handled_total",
		otelmetric.WithDescription("Total number of RPCs handled on the server."),
	)
	return serverHandledCounter
}

func newTestServerHandledHistogram(serverMeter otelmetric.Meter) otelmetric.Float64Histogram {
	serverHandledHistogram, _ := serverMeter.Float64Histogram(
		"rpc_server_handled_seconds",
		otelmetric.WithDescription(
			"Histogram of response latency (seconds) of tRPC that "+
				"had been application-level handled by the server.",
		),
		otelmetric.WithExplicitBucketBoundaries(0, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 5), // 单位：秒。
	)
	return serverHandledHistogram
}

func newTestClientMeter(provider *sdkmetric.MeterProvider) otelmetric.Meter {
	clientMeter := provider.Meter(clientMeterName)
	return clientMeter
}

func newTestClientHandledCounter(clientMeter otelmetric.Meter) otelmetric.Int64Counter {
	clientHandledCounter, _ := clientMeter.Int64Counter(
		"rpc_client_handled_total",
		otelmetric.WithDescription("Total number of RPCs handled on the client."),
	)
	return clientHandledCounter
}

func newTestClientHandledHistogram(clientMeter otelmetric.Meter) otelmetric.Float64Histogram {
	clientHandledHistogram, _ := clientMeter.Float64Histogram(
		"rpc_client_handled_seconds",
		otelmetric.WithDescription(
			"Histogram of response latency (seconds) of tRPC that "+
				"had been application-level handled by the client.",
		),
		otelmetric.WithExplicitBucketBoundaries(0, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 5), // 单位：秒。
	)
	return clientHandledHistogram
}

func newCustomXXXMeter(provider *sdkmetric.MeterProvider) otelmetric.Meter {
	// 上报自定义指标。
	monitorXXX := provider.Meter("monitor_xxx") // 监控项名：monitor_xxx。
	return monitorXXX
}

func newCustomXXXTotal(monitorXXX otelmetric.Meter) otelmetric.Int64Counter {
	xxxTotal, _ := monitorXXX.Int64Counter("xxx_total") // 指标名：xxx_total。伽利略暂时只支持 delta sum。
	return xxxTotal
}

func newCustomXXXSeconds(monitorXXX otelmetric.Meter) otelmetric.Float64Histogram {
	xxxSeconds, _ := monitorXXX.Float64Histogram("xxx_seconds") // 指标名：xxx_seconds。
	return xxxSeconds
}

var yyyBytesVal int64

func newCustomXXXBytes(monitorXXX otelmetric.Meter) otelmetric.Int64ObservableGauge {
	yyyBytes, _ := monitorXXX.Int64ObservableGauge("yyy_bytes") // 指标名：yyy_bytes。
	_, _ = monitorXXX.RegisterCallback(
		func(ctx context.Context, o otelmetric.Observer) error {
			o.ObserveInt64(yyyBytes, yyyBytesVal)
			fmt.Printf("callback yyyBytesVal address=%p value=%d\n", &yyyBytesVal, yyyBytesVal)
			return nil
		}, yyyBytes,
	)
	yyyBytesVal = rand.Int63n(100)
	fmt.Printf("init yyyBytesVal address=%p value=%d\n", &yyyBytesVal, yyyBytesVal)
	return yyyBytes
}

func TestProvider(t *testing.T) {
	sleepDuration := 0 * time.Second // 为了完成异步上报，测试结束需要等待上报，默认睡眠时间为 0 秒
	if len(os.Args) > 1 {            // 检查命令行参数，report.sh 脚本会有参数。
		durationStr := os.Args[1]                                         // 获取第一个参数
		if duration, err := time.ParseDuration(durationStr); err == nil { // 解析时间字符串
			sleepDuration = duration
		}
	}
	// 测试空 provider
	provider, err := NewMeterProvider(model.Resource{}, model.OpenTelemetryPushConfig{}, WithAPIKey("abcd"))
	assert.Nil(t, provider)
	assert.Error(t, err)
	// 测试正常构造的 provider
	assert.NotNil(t, testProvider)
	// 测试上报被调。
	reportRPCServerMetric(context.Background(), testServerHandledCounter, testServerHandledHistogram)
	// 测试上报主调。
	reportRPCClientMetric(context.Background(), testClientHandledCounter, testClientHandledHistogram)
	// 测试上报自定义。
	reportCustomXXXTotal(context.Background(), xxxTotal)
	reportCustomXXXSeconds(context.Background(), xxxSeconds)
	reportCustomXXXBytes()
	// 等待异步上报完成。
	time.Sleep(sleepDuration)
}

func reportRPCServerMetric(
	ctx context.Context,
	serverHandledCounter otelmetric.Int64Counter,
	serverHandledHistogram otelmetric.Float64Histogram,
) {
	startTime := time.Now()
	// 如下 rpc 属性字段，都需要替换成每次调用真实的值。。
	// 字段含义见：https://iwiki.woa.com/p/4009259982
	kvs := make([]attribute.KeyValue, 0, 18) // 7 个主调信息、6 个被调信息、3 个扩展字段、2 个错误码字段。
	// 主调信息。
	kvs = append(kvs, attrCallerServer.String("CallerServer"))       // 主调服务
	kvs = append(kvs, attrCallerService.String("CallerService"))     // 主调 service
	kvs = append(kvs, attrCallerMethod.String("CallerMethod"))       // 主调接口
	kvs = append(kvs, attrCallerConSetid.String(""))                 // 主调，set id
	kvs = append(kvs, attrCallerIP.String("CallerIP"))               // 主调 ip
	kvs = append(kvs, attrCallerContainer.String("CallerContainer")) // 主调容器。
	kvs = append(kvs, attrCallerGroup.String("CallerGroup"))         // 主调流量组
	// 被调信息。
	kvs = append(kvs, attrCalleeServer.String("CalleeServer"))     // 被调服务
	kvs = append(kvs, attrCalleeService.String("CalleeService"))   // 被调 service
	kvs = append(kvs, attrCalleeMethod.String("CalleeMethod"))     // 被调接口
	kvs = append(kvs, attrCalleeConSetid.String("FullSetName"))    // 被调 set id
	kvs = append(kvs, attrCalleeIP.String("LocalIP"))              // 被调 IP。
	kvs = append(kvs, attrCalleeContainer.String("ContainerName")) // 被调容器。
	// 扩展字段。
	kvs = append(kvs, attrUserExt1.String("")) // 预留字段 1
	kvs = append(kvs, attrUserExt2.String("")) // 预留字段 2
	kvs = append(kvs, attrUserExt3.String("")) // 预留字段 3
	// 模拟 rpc 处理
	code, codeType := fakeRPCHandle()
	// 错误码。
	kvs = append(kvs, attrCode.String(code))
	kvs = append(kvs, attrCodeType.String(codeType))
	// 上报数据。
	serverHandledCounter.Add(ctx, 1, otelmetric.WithAttributes(kvs...))
	serverHandledHistogram.Record(ctx, time.Since(startTime).Seconds(), otelmetric.WithAttributes(kvs...)) // 单位秒。
}

func reportRPCClientMetric(
	ctx context.Context,
	clientHandledCounter otelmetric.Int64Counter,
	clientHandledHistogram otelmetric.Float64Histogram,
) {
	startTime := time.Now()
	// 如下 rpc 属性字段，都需要替换成每次调用真实的值。。
	// 字段含义见：https://iwiki.woa.com/p/4009259982
	kvs := make([]attribute.KeyValue, 0, 18) // 7 个主调信息、6 个被调信息、3 个扩展字段、2 个错误码字段。
	// 主调信息。
	kvs = append(kvs, attrCallerServer.String("CallerServer"))       // 主调服务
	kvs = append(kvs, attrCallerService.String("CallerService"))     // 主调 service
	kvs = append(kvs, attrCallerMethod.String("CallerMethod"))       // 主调接口
	kvs = append(kvs, attrCallerConSetid.String(""))                 // 主调，set id
	kvs = append(kvs, attrCallerIP.String("CallerIP"))               // 主调 ip
	kvs = append(kvs, attrCallerContainer.String("CallerContainer")) // 主调容器。
	kvs = append(kvs, attrCallerGroup.String("CallerGroup"))         // 主调流量组
	// 被调信息。
	kvs = append(kvs, attrCalleeServer.String("CalleeServer"))     // 被调服务
	kvs = append(kvs, attrCalleeService.String("CalleeService"))   // 被调 service
	kvs = append(kvs, attrCalleeMethod.String("CalleeMethod"))     // 被调接口
	kvs = append(kvs, attrCalleeConSetid.String("FullSetName"))    // 被调 set id
	kvs = append(kvs, attrCalleeIP.String("LocalIP"))              // 被调 IP。
	kvs = append(kvs, attrCalleeContainer.String("ContainerName")) // 被调容器。
	// 扩展字段。
	kvs = append(kvs, attrUserExt1.String("")) // 预留字段 1
	kvs = append(kvs, attrUserExt2.String("")) // 预留字段 2
	kvs = append(kvs, attrUserExt3.String("")) // 预留字段 3
	// 模拟 rpc 处理
	code, codeType := fakeRPCHandle()
	// 错误码。
	kvs = append(kvs, attrCode.String(code))
	kvs = append(kvs, attrCodeType.String(codeType))
	// 上报数据。
	clientHandledCounter.Add(ctx, 1, otelmetric.WithAttributes(kvs...))
	clientHandledHistogram.Record(ctx, time.Since(startTime).Seconds(), otelmetric.WithAttributes(kvs...)) // 单位秒。
}

// fakeRPCHandle 模拟 rpc server 处理。
func fakeRPCHandle() (string, string) {
	costMS := rand.Intn(100) // 随机处理耗时。
	time.Sleep(time.Millisecond * time.Duration(costMS))
	if costMS > 90 { // 大于 90 毫秒假设超时。
		return codeTimeout, codeTypeTimeout
	} else if costMS > 50 { // 大于 50 毫秒假设异常。
		return codeException, codeTypeException
	}
	// 假设成功。
	return codeSuccess, codeTypeSuccess
}

func reportCustomXXXTotal(ctx context.Context, xxxTotal otelmetric.Int64Counter) {
	xxxTotal.Add(ctx, rand.Int63n(100))
}

func reportCustomXXXSeconds(ctx context.Context, xxxSeconds otelmetric.Float64Histogram) {
	xxxSeconds.Record(ctx, rand.Float64())
}

func reportCustomXXXBytes() {
	yyyBytesVal = rand.Int63n(100)
	fmt.Printf("report yyyBytesVal address=%p value=%d\n", &yyyBytesVal, yyyBytesVal)
}

func Test_getAggregationSelector(t *testing.T) {
	aggregationSelector := getAggregationSelector([]float64{0, 1, 5})
	assert.Equal(t, sdkmetric.AggregationSum{}, aggregationSelector(sdkmetric.InstrumentKindCounter))
	assert.Equal(t, sdkmetric.AggregationLastValue{}, aggregationSelector(sdkmetric.InstrumentKindGauge))
	assert.Equal(
		t, sdkmetric.AggregationExplicitBucketHistogram{
			Boundaries: []float64{0, 1, 5},
			NoMinMax:   false,
		}, aggregationSelector(sdkmetric.InstrumentKindHistogram),
	)
	assert.Equal(t, sdkmetric.AggregationSum{}, aggregationSelector(sdkmetric.InstrumentKind(0)))
}
