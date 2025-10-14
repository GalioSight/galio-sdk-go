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

package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"galiosight.ai/galio-sdk-go/configs"
	otphttp "galiosight.ai/galio-sdk-go/exporters/otp/http"
	"galiosight.ai/galio-sdk-go/model"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/stretchr/testify/require"
)

var data = &model.Metrics{
	TimestampMs: time.Now().Unix() * 1000,
	NormalLabels: &model.NormalLabels{
		Fields: []model.NormalLabels_Field{
			{
				Name:  model.NormalLabels_target,
				Value: "PCG-123.galileo.SDK",
			},
			{
				Name:  model.NormalLabels_namespace,
				Value: "Production",
			},
			{
				Name:  model.NormalLabels_env_name,
				Value: "formal",
			},
			{
				Name:  model.NormalLabels_region,
				Value: "gz",
			},
			{
				Name:  model.NormalLabels_instance,
				Value: "a.b.c.d",
			},
			{
				Name:  model.NormalLabels_node,
				Value: "cls-j7om2txw-808101ce6c0bbe4954ebb7fc901de0ee-0",
			},
			{
				Name:  model.NormalLabels_container_name,
				Value: "formal.galileo.apiserver.gz100001",
			},
		},
	},
	ClientMetrics: []*model.ClientMetricsOTP{
		{
			RpcClientStartedTotal: 102,
			RpcClientHandledTotal: 203,
			RpcClientHandledSeconds: &model.Histogram{
				Sum:   101.2,
				Count: 3,
				Buckets: []*model.Bucket{
					{
						Range: "23",
						Count: 19,
					},
					{
						Range: "367",
						Count: 10,
					},
				},
			},
			RpcLabels: &model.RPCLabels{
				Fields: []model.RPCLabels_Field{
					{
						Name:  model.RPCLabels_caller_service,
						Value: "trpc.galileo.apiserver.apiserver",
					},
					{
						Name:  model.RPCLabels_callee_ip,
						Value: "a.b.c.d",
					},
				},
			},
		},
		{
			RpcClientStartedTotal: 102,
			RpcClientHandledTotal: 203,
			RpcClientHandledSeconds: &model.Histogram{
				Sum:   101.2,
				Count: 3,
				Buckets: []*model.Bucket{
					{
						Range: "33",
						Count: 19,
					},
					{
						Range: "44",
						Count: 10,
					},
				},
			},
			RpcLabels: &model.RPCLabels{
				Fields: []model.RPCLabels_Field{
					{
						Name:  model.RPCLabels_caller_service,
						Value: "trpc.galileo.apiserver.apiserver",
					},
					{
						Name:  model.RPCLabels_callee_ip,
						Value: "a.b.c.d",
					},
				},
			},
		},
	},
	ServerMetrics: []*model.ServerMetricsOTP{
		{
			RpcServerStartedTotal: 102,
			RpcServerHandledTotal: 203,
			RpcServerHandledSeconds: &model.Histogram{
				Sum:   101.2,
				Count: 3,
				Buckets: []*model.Bucket{
					{
						Range: "33",
						Count: 19,
					},
					{
						Range: "44",
						Count: 10,
					},
				},
			},
			RpcLabels: &model.RPCLabels{
				Fields: []model.RPCLabels_Field{
					{
						Name:  model.RPCLabels_caller_service,
						Value: "trpc.galileo.apiserver.apiserver",
					},
					{
						Name:  model.RPCLabels_callee_ip,
						Value: "a.b.c.d",
					},
				},
			},
		},
		{
			RpcServerStartedTotal: 102,
			RpcServerHandledTotal: 203,
			RpcServerHandledSeconds: &model.Histogram{
				Sum:   101.2,
				Count: 3,
				Buckets: []*model.Bucket{
					{
						Range: "33",
						Count: 19,
					},
					{
						Range: "44",
						Count: 10,
					},
				},
			},
			RpcLabels: &model.RPCLabels{
				Fields: []model.RPCLabels_Field{
					{
						Name:  model.RPCLabels_caller_service,
						Value: "trpc.galileo.apiserver.apiserver",
					},
					{
						Name:  model.RPCLabels_callee_ip,
						Value: "a.b.c.d",
					},
				},
			},
		},
	},
	NormalMetrics: []*model.NormalMetricOTP{
		{
			Metric: &model.MetricOTP{
				Name: "trpc.test",
				V: &model.MetricOTP_Value{
					Value: 100,
				},
				Aggregation: model.Aggregation_AGGREGATION_SUM,
			},
		},
		{
			Metric: &model.MetricOTP{
				Name: "trpc.test",
				V: &model.MetricOTP_Value{
					Value: 100,
				},
				Aggregation: model.Aggregation_AGGREGATION_SUM,
			},
		},
	},
	CustomMetrics: []*model.CustomMetricsOTP{
		{
			Metrics: []*model.MetricOTP{
				{
					Name: "trpc.test",
					V: &model.MetricOTP_Value{
						Value: 100,
					},
					Aggregation: model.Aggregation_AGGREGATION_SUM,
				},
			},
		},
		{
			Metrics: []*model.MetricOTP{
				{
					Name: "trpc.test",
					V: &model.MetricOTP_Value{
						Value: 100,
					},
					Aggregation: model.Aggregation_AGGREGATION_SUM,
				},
			},
		},
	},
}

func Test_metricsExporter_Export(t *testing.T) {
	var ts = httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"code":0,"msg":"success"}`))
			},
		),
	)
	defer ts.Close()
	exporter, err := NewExporter(
		&configs.Metrics{
			Exporter: model.MetricsExporter{
				Protocol:      "otp",
				Collector:     model.Collector{Addr: ts.URL},
				ThreadCount:   1,
				BufferSize:    10,
				WindowSeconds: 1,
				PageSize:      100,
				TimeoutMs:     10000,
			},
			Stats: &model.SelfMonitorStats{},
		},
	)
	require.Nil(t, err)
	exporter.Export(proto.Clone(data).(*model.Metrics))
	// WindowSeconds 是 1 秒，此处等 2 秒，让数据异步报完，期望数据在 2 秒内报完
	time.Sleep(time.Second * 2)
	stats := exporter.(*metricsExporter).stats
	require.Equal(t, int64(0), stats.ReportErrorTotal.Load())
	require.Equal(t, int64(1), stats.ReportHandledTotal.Load())
}

func Test_metricsExporter_Export_Page(t *testing.T) {
	var ts = httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"code":0,"msg":"success"}`))
			},
		),
	)
	defer ts.Close()
	exporter, err := NewExporter(
		&configs.Metrics{
			Exporter: model.MetricsExporter{
				Protocol:      "otp",
				Collector:     model.Collector{Addr: ts.URL},
				ThreadCount:   1,
				BufferSize:    10,
				WindowSeconds: 1,
				PageSize:      100,
				TimeoutMs:     1000,
			},
			Stats: &model.SelfMonitorStats{},
		},
	)
	require.Nil(t, err)
	exporter.Export(proto.Clone(data).(*model.Metrics))
	// WindowSeconds 是 1 秒，此处等 2 秒，让数据异步报完，期望数据在 2 秒内报完
	time.Sleep(time.Second * 2)
	stats := exporter.(*metricsExporter).stats
	require.Equal(t, int64(0), stats.ReportErrorTotal.Load())
	require.Equal(t, int64(0), stats.ReportErrorRowsTotal.Load())
	require.Equal(t, int64(1), stats.ReportHandledTotal.Load())
	require.Less(t, int64(1), stats.ReportHandledRowsTotal.Load())
}

func Test_metricsExporter_Export_Error(t *testing.T) {
	cfg := &configs.Metrics{
		Exporter: model.MetricsExporter{
			Protocol:      "otp",
			Collector:     model.Collector{Addr: ""},
			ThreadCount:   1,
			BufferSize:    10,
			WindowSeconds: 1,
			PageSize:      100,
		},
	}
	exporter, err := NewExporter(
		cfg,
	)
	m := exporter.(*metricsExporter)
	t.Logf("exporter=%+v", m.cfg)
	require.Nil(t, err)
	exporter.Export(proto.Clone(data).(*model.Metrics))
	// WindowSeconds 是 1 秒，此处等 2 秒，让数据异步报完，期望数据在 2 秒内报完
	time.Sleep(time.Second * 2)
	stats := m.stats
	require.Equal(t, int64(1), stats.ReportErrorTotal.Load())
	require.Equal(t, int64(1), stats.ReportHandledTotal.Load())
	require.Less(t, int64(1), stats.ReportErrorRowsTotal.Load())
	require.Less(t, int64(1), stats.ReportHandledRowsTotal.Load())
}

func Test_metricsExporter_UpdateConfig(t *testing.T) {
	var ts = httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"code":0,"msg":"success"}`))
			},
		),
	)
	defer ts.Close()
	cfg := &configs.Metrics{
		Exporter: model.MetricsExporter{
			Protocol:      "otp",
			Collector:     model.Collector{Addr: ""},
			ThreadCount:   1,
			BufferSize:    10,
			WindowSeconds: 1,
			PageSize:      100,
		},
	}
	exporter, err := NewExporter(
		cfg,
	)
	m := exporter.(*metricsExporter)
	t.Logf("exporter=%+v", m.cfg)
	require.Nil(t, err)

	cfg.Exporter.Collector.Addr = ts.URL
	exporter.UpdateConfig(cfg)
	require.Equal(t, ts.URL, m.httpExporter.(*otphttp.HTTPGeneralExporter).CollectorAddr.FullURL)

	exporter.Export(proto.Clone(data).(*model.Metrics))
	// WindowSeconds 是 1 秒，此处等 2 秒，让数据异步报完，期望数据在 2 秒内报完
	time.Sleep(time.Second * 2)
	stats := exporter.(*metricsExporter).stats
	require.Equal(t, int64(0), stats.ReportErrorTotal.Load())
	require.Equal(t, int64(0), stats.ReportErrorRowsTotal.Load())
	require.Equal(t, int64(1), stats.ReportHandledTotal.Load())
	require.Less(t, int64(1), stats.ReportHandledRowsTotal.Load())
}

func Test_snappy(t *testing.T) {
	s := "x"
	buf := make([]byte, 10)
	ret := snappy.Encode(buf, []byte(s))
	require.Equal(t, []byte{1, 0, 120}, ret)
}
