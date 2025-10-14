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

//go:build apitest

package push

import (
	"testing"
	"time"

	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/version"
	"github.com/prometheus/client_golang/prometheus"
)

func TestPrometheusPushAPI(t *testing.T) {
	// Resource 用来标识资源信息，每个进程的 res 应该是唯一的，不会和其他进程相同，以避免与其他进程指标混淆
	res := model.Resource{
		// 以下是对各字段的具体赋值，应该根据实际情况修改
		Target:        "PCG-123.example.greeter", // target 需要在伽利略平台注册
		Namespace:     string(model.Production),  // 名字空间，只能是 Development 或 Production
		EnvName:       "andy-test",               // 环境名，
		Instance:      "127.0.0.1",               // 实例，通常使用机器 ip
		ContainerName: "formal.ContainerName.1",  // 容器名
		Version:       version.Number,            // 版本号，这个固定为 version.Number
		SetName:       "Set1",                    // SetName，
		City:          "Beijing",                 // 城市，
		App:           "example",                 // App，
		Server:        "greeter",                 // Server，
	}

	// Prometheus 运行时配置
	cfg := model.PrometheusPushConfig{
		// 开关，
		Enable: true,
		// Prometheus 推送的目标 URL，默认中国大陆内网，参考：https://iwiki.woa.com/p/4010767585
		//Url: "https://gotp.testsite.woa.com",
		Url: "http://wr.otp.woa.com",
		// 任务名，可以根据需要取个名字
		Job: "testjob",
		// 上报间隔，默认 20 秒，如果需要秒级监控，可以改成 1 秒。注意，间隔越小，相同时间上报的数据越多，成本越高。
		Interval: 20,
		Grouping: map[string]string{
			// tps_tenant_id 用于标识天机阁的租户，不是伽利略的租户。
			// 如果是天机阁迁移过来的服务，需要填充此字段，以便兼容天机阁的看板。此处 apitest 用于演示。
			"tps_tenant_id": "apitest",
			// 可以根据需要，再增加一些资源字段，以下是示例
			"pod_name":        "pod_name-001",
			"business_module": "business_module-001",
		},
		// http 头，如果有特殊需要可以添加一些自定义头
		HttpHeaders: nil,
	}
	// PrometheusPush 函数会在后台启动协程，将 Prometheus 指标定时 push 上报
	cancel, err := PrometheusPush(res, cfg)
	if err != nil {
		t.Fatalf("PrometheusPush() error = %v", err)
	}
	// 当需要停止上报的时候调用 cancel()，正常的服务应该持续上报指标，不需要调用 cancel()
	//defer cancel()
	_ = cancel

	go func() {
		tick := time.NewTicker(time.Second * 10)
		for range tick.C {
			requestCount.WithLabelValues("foo", "bar").Inc()
			RPCClientHandledTotal.WithLabelValues(prcLabelValues...).Inc()
			RPCClientStartedTotal.WithLabelValues(prcLabelValues...).Inc()
			RPCClientHandledSeconds.WithLabelValues(prcLabelValues...).Observe(11)
			RPCClientHandledSeconds.WithLabelValues(prcLabelValues...).Observe(10)
			requestHist.WithLabelValues().Observe(3)
		}
	}()

	// 此处演示，持续上报一段时间，正常服务应该持续上报，直到进程终止
	time.Sleep(2 * time.Hour)
}

var prcLabels = []string{
	"_caller_service_",
	"_caller_method_",
	"_caller_con_setid_",
	"_caller_ip_",
	"_caller_container_",
	"_callee_service_",
	"_callee_method_",
	"_callee_con_setid_",
	"_callee_ip_",
	"_callee_container_",
	"_code_",
	"_code_type_",
	"_caller_group_",
	"_user_ext1_",
	"_user_ext2_",
	"_user_ext3_",
	"_caller_target_",
	"_caller_server_",
	"_callee_target_",
	"_callee_server_",
	"_canary_",
	"_flow_tag_",
	"_monitor_name_",
}

var prcLabelValues = []string{
	"_caller_service_",
	"_caller_method_",
	"_caller_con_setid_",
	"_caller_ip_",
	"_caller_container_",
	"_callee_service_",
	"_callee_method_",
	"_callee_con_setid_",
	"_callee_ip_",
	"_callee_container_",
	"_code_",
	"success",
	"_caller_group_",
	"_user_ext1_",
	"_user_ext2_",
	"_user_ext3_",
	"_caller_target_",
	"_caller_server_",
	"_callee_target_",
	"_callee_server_",
	"_canary_",
	"_flow_tag_",
	"rpc_client",
}

var (
	// 创建一个指标，用于演示
	requestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "TestPrometheusPushAPI_demo_total",
			Help: "Test PrometheusPushAPI demo",
		},
		[]string{"method", "endpoint"},
	)

	requestHist = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "TestPrometheusPushAPI_hist_data", Help: "RPCClientHandledSecondsDescription"},
		nil)

	RPCClientHandledSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "rpc_client_handled_seconds", Help: "RPCClientHandledSecondsDescription"},
		prcLabels)
	RPCClientHandledTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "rpc_client_handled_total", Help: "RPCClientHandledTotalDescription"},
		prcLabels)
	RPCClientStartedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "rpc_client_started_total", Help: "RPCClientStartedTotalDescription"},
		prcLabels)
)

func init() {
	// 注册指标到 Prometheus
	prometheus.MustRegister(requestCount)
	prometheus.MustRegister(requestHist)
	prometheus.MustRegister(RPCClientHandledSeconds)
	prometheus.MustRegister(RPCClientHandledTotal)
	prometheus.MustRegister(RPCClientStartedTotal)
}
