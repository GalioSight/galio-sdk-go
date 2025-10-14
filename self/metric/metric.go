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

// Package metric 自监控指标。
package metric

import (
	"sync"
	"time"

	otphttp "galiosight.ai/galio-sdk-go/exporters/otp/http"
	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
)

// SelfMonitor 自监控。
type SelfMonitor struct {
	// Stats 自监控统计数据，1 s 精度。
	Stats *model.SelfMonitorStats
	// exporter 自监控导出器。
	exporter otphttp.HTTPExporter
	// log 日志 wrapper。
	log *logs.Wrapper
	// 上次上报时的指标数据，用于和当前指标数据计算增量值。初始为 nil。
	lastStats *model.SelfMonitorStats
}

// selfMonitor 自监控对象。默认的空对象是不会工作的，需要调用 Init 方法之后才能正常工作。
// 默认创建一个空对象，避免业务未初始化时导致 panic。
var selfMonitor = &SelfMonitor{
	Stats:     &model.SelfMonitorStats{},
	exporter:  otphttp.NewHTTPGeneralExporter(1000, "", logs.DefaultWrapper()),
	log:       logs.DefaultWrapper(),
	lastStats: &model.SelfMonitorStats{},
}

// GetSelfMonitor 获取自监控单例对象。
// 调用此函数之前，必须先调用 Init 方法，否则会返回空指针。
func GetSelfMonitor() *SelfMonitor {
	return selfMonitor
}

var once sync.Once

const selfTenant = "galileo"
const selfTarget = "PCG-123.galileo.otp"

type options struct {
	apiKey string // 保存 APIKey 的可选参数
}

// InitOption 定义选项函数类型
type InitOption func(*options)

// WithAPIKey 创建选项函数，用于设置 APIKey
func WithAPIKey(apiKey string) InitOption {
	return func(o *options) {
		o.apiKey = apiKey
	}
}

// Init 初始化默认自监控对象。
// 自监控对象只能有一个，所以此方法只需要调用一次。
// 多次调用的话，只有第一次会执行。
// 通常在启动的时候进行初始化。
func Init(
	resource *model.Resource,
	monitor model.SelfMonitor,
	log *logs.Wrapper,
	opts ...InitOption, // 添加可变长选项参数
) {
	once.Do(
		func() {
			config := options{
				apiKey: "", // 默认值为空，保持兼容性
			}
			for _, opt := range opts {
				opt(&config)
			}

			// 使用 config.apiKey 替换硬编码的空字符串
			selfMonitor.exporter = otphttp.NewHTTPGeneralExporter(
				1000*10, monitor.Collector.Addr, log,
				otphttp.WithHeaders(
					map[string]string{
						model.TenantHeaderKey: selfTenant,
						model.TargetHeaderKey: selfTarget,
						model.APIKeyHeaderKey: config.apiKey, // 动态设置 APIKey
					},
				),
				otphttp.WithMaxRetryCount(2),
			)
			selfMonitor.log = log
			n := selfMetricNormalLabels(resource)
			go selfMonitor.Report(n, resource.Target, monitor.ReportSeconds)
		},
	)
}

// selfMetricNormalLabels 将 Resource 转成 NormalLabels。
// target 使用 "PCG-123.galileo.otp"，将数据全部报到伽利略的服务上，方便统一查看。
func selfMetricNormalLabels(r *model.Resource) *model.NormalLabels {
	normalLabels := model.NewNormalLabels()
	if r == nil {
		return normalLabels
	}
	normalLabels.Fields[model.NormalLabels_target].Value = selfTarget
	normalLabels.Fields[model.NormalLabels_namespace].Value = "Production"
	normalLabels.Fields[model.NormalLabels_env_name].Value = r.EnvName
	normalLabels.Fields[model.NormalLabels_region].Value = ""
	normalLabels.Fields[model.NormalLabels_instance].Value = ""
	normalLabels.Fields[model.NormalLabels_node].Value = ""
	normalLabels.Fields[model.NormalLabels_container_name].Value = ""
	normalLabels.Fields[model.NormalLabels_version].Value = r.Version
	return normalLabels
}

// Report 定时上报自监控数据，默认自监控 10 s 上报一次。
func (s *SelfMonitor) Report(
	normalLabels *model.NormalLabels, target string, seconds int32,
) {
	if seconds < 1 {
		seconds = 1
	}
	ticker := time.NewTicker(time.Second * time.Duration(seconds))
	defer ticker.Stop()
	reuseObj := otphttp.NewReuseObject()
	for range ticker.C {
		s.report(normalLabels, target, reuseObj)
	}
}

func (s *SelfMonitor) report(
	normalLabels *model.NormalLabels, target string, r *otphttp.ReuseObject,
) {
	curStats := *s.Stats
	metrics := model.GetDeltaMetrics(s.lastStats, &curStats, target)
	s.lastStats = &curStats
	metrics.NormalLabels = normalLabels
	metrics.TimestampMs = time.Now().Unix() * 1000
	s.Stats.SelfMonitorCount.Inc()
	if err := s.exporter.Export(metrics, r); err != nil {
		s.Stats.SelfMonitorError.Inc()
		s.log.Errorf("[galileo]selfMonitor.report|err=%v\n", err)
	} else {
		s.log.Infof(
			"[galileo]selfMonitor.report|stats=%+v,metrics=%+v\n",
			s.Stats.MetricsStats, metrics,
		)
	}
}
