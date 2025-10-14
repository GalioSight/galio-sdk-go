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

// Package prometheus ...
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/expfmt"

	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/semconv"
	"galiosight.ai/galio-sdk-go/version"
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
	h := http.Header{}
	h.Set(model.TargetHeaderKey, "STKE.example.omp_v3_go")
	h.Set(model.SchemaURLHeaderKey, semconv.SchemaURL) // 声明使用 OMP v3.0.0 规范
	pushURL := "<push url>"
	pusher = push.New(pushURL, "galileo").Gatherer(prometheus.DefaultGatherer).
		Format(expfmt.NewFormat(expfmt.TypeProtoDelim)).Header(h)
	pusher.Grouping(V(semconv.TelemetryTargetKey.String("STKE.example.omp_v3_go"))) // 可观测对象
	pusher.Grouping(V(semconv.DeploymentNamespaceProduction))                       // 物理环境 Production 或者 Development
	pusher.Grouping(V(semconv.DeploymentEnvironmentNameKey.String("formal")))       // 用户环境
	pusher.Grouping(V(semconv.ServiceSetNameKey.String("set.sz.1")))                // 容器集合
	pusher.Grouping(V(semconv.HostIPKey.String("127.0.0.1")))                       // IP
	pusher.Grouping(V(semconv.ContainerNameKey.String("test.galileo.SDK.sz10010"))) // 容器名
	pusher.Grouping(V(semconv.TelemetrySDKVersionKey.String(version.Number)))       // sdk 版本
	pusher.Grouping(V(semconv.DeploymentCityKey.String("sz")))                      // city
	pusher.Grouping(V(semconv.ServiceVersionKey.String("push-v0.1.0")))             // 服务发布版本 (镜像)
	pusher.Grouping(V(semconv.TelemetrySDKLanguageGo))                              // 上报语言
	pusher.Grouping(V(semconv.TelemetrySDKNameKey.String("prom-push")))             // sdk 名称
	return startTicker(pusher, 10*time.Second)                                      // 上报间隔为 10 秒，可以自定义。
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
	// 创建一个计数器
	labels = semconv.PrometheusLabels(
		semconv.RPCCallerMethodKey,
		semconv.RPCCallerServerKey,
		semconv.RPCCallerServiceKey,
		semconv.RPCCallerContainerKey,
		semconv.RPCCallerIPKey,
		semconv.RPCCallerSetKey,
		semconv.RPCCalleeMethodKey,
		semconv.RPCCalleeServerKey,
		semconv.RPCCalleeServiceKey,
		semconv.RPCCalleeContainerKey,
		semconv.RPCCalleeIPKey,
		semconv.RPCCalleeSetKey,
		semconv.RPCErrorCodeKey,
		semconv.RPCErrorCodeTypeKey,
		semconv.RPCCallerGroupKey,
		semconv.RPCCanaryKey,
		semconv.RPCUserExt1Key,
		semconv.RPCUserExt2Key,
		semconv.RPCUserExt3Key,
		semconv.MonitorNameKey,
	)
	customs                 = []string{"foo"}
	RPCClientHandledSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: semconv.RPCClientHandledSecondsName, Help: semconv.RPCClientHandledSecondsDescription},
		labels)
	RPCClientHandledTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: semconv.RPCClientHandledTotalName, Help: semconv.RPCClientHandledTotalDescription},
		labels)
	RPCClientStartedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: semconv.RPCClientStartedTotalName, Help: semconv.RPCClientStartedTotalDescription},
		labels)
	RPCServerHandledSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: semconv.RPCServerHandledSecondsName, Help: semconv.RPCServerHandledSecondsDescription},
		labels)
	RPCServerHandledTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: semconv.RPCServerHandledTotalName, Help: semconv.RPCServerHandledTotalDescription},
		labels)
	RPCServerStartedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: semconv.RPCServerStartedTotalName, Help: semconv.RPCServerStartedTotalDescription},
		labels)
	CustomCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "custom_counter_total", Help: "custom counter demo",
			ConstLabels: prometheus.Labels{P(semconv.MonitorNameKey): "custom_monitor"}}, customs)
	CustomGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "custom_gauge_set", Help: "custom gauge demo",
			ConstLabels: prometheus.Labels{P(semconv.MonitorNameKey): "custom_monitor"}}, customs)
	CustomHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "custom_histogram", Help: "custom histogram demo",
			ConstLabels: prometheus.Labels{P(semconv.MonitorNameKey): "custom_monitor"}}, customs)
)

func init() {
	// 注册计数器到 Prometheus
	prometheus.MustRegister(RPCClientHandledTotal)
	prometheus.MustRegister(RPCClientHandledSeconds)
	prometheus.MustRegister(RPCClientStartedTotal)
	prometheus.MustRegister(RPCServerHandledTotal)
	prometheus.MustRegister(RPCServerHandledSeconds)
	prometheus.MustRegister(RPCServerStartedTotal)
	prometheus.MustRegister(CustomCounter)
	prometheus.MustRegister(CustomGauge)
	prometheus.MustRegister(CustomHistogram)
}

// startTicker 启动定时器以定期推送指标。
// 如果配置的间隔有效，返回一个取消函数；否则返回相应的错误。
func startTicker(pusher *push.Pusher, interval time.Duration) (context.CancelFunc, error) {
	if interval <= 0 {
		return nil, ErrCfgInterval
	}
	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(interval)
	_ = doPush(pusher)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = doPush(pusher)
			case <-ctx.Done():
				return
			}
		}
	}()
	return cancel, nil
}

func doPush(pusher *push.Pusher) error {
	if err := pusher.Push(); err != nil {
		log.Printf("failed to push prometheus metrics: %v\n", err)
		return err
	} else {
		log.Printf("success to push prometheus metrics\n")
	}
	return nil
}

// ErrCfgInterval 表示配置的间隔时间无效。
// 当 cfg.Interval <= 0 时返回该错误。
var ErrCfgInterval = errors.New("cfg.Interval <= 0")

type chain []func(ctx context.Context, next func(context.Context) error) error

func (c chain) Run(ctx context.Context) error {
	if len(c) == 1 {
		return c[0](ctx, nil)
	}
	f := c[0]
	left := c[1:]
	return f(ctx, left.Run)
}

func rpcValues(monitorName, codeType string) []string {
	return []string{
		"DemoCallerMethod",
		"DemoCallerSever",
		"DemoCallerService",
		"test.galileo.sdkserver.sz10010",
		"127.0.0.1",
		"set.gz1.4s8g",
		"DemoCalleeMethod",
		"DemoCalleeServer",
		"DemoCalleeService",
		"test.example.demoserver.sz10020",
		"127.0.0.1",
		"set.sz1.abc1",
		"0",
		codeType,
		"",
		"",
		"",
		"",
		"",
		monitorName,
	}
}
