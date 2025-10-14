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

package ocp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"galiosight.ai/galio-sdk-go/model"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v3"
)

// testResponse 重置 response 并返回一个可更改的对象
func testResponse() *model.GetConfigResponse {
	var rsp model.GetConfigResponse
	json.Unmarshal([]byte(`{
  "msg": "success",
  "target": "PCG-123.example.greeter",
  "config_server": "",
  "register_server": "",
  "self_monitor": {
    "protocol": "otp",
    "collector": {
      "addr": "",
      "telemetry_data": 1,
      "data_protocol": 5,
      "data_transmission": 1
    }
  },
  "metrics_config": {
    "enable": true,
    "processor": {
      "protocol": "omp",
      "window_seconds": 20,
      "clear_seconds": 3600,
      "expires_seconds": 3600,
      "point_limit": 200000,
      "enable_process_metrics": true,
      "process_metrics_seconds": 20,
      "histogram_buckets": [
        {
          "name": "rpc_server_handled_seconds",
          "buckets": [
            0.005,
            0.01,
            0.025,
            0.05,
            0.1,
            0.25,
            0.5,
            1,
            5
          ]
        },
        {
          "name": "rpc_client_handled_seconds",
          "buckets": [
            0.005,
            0.01,
            0.025,
            0.05,
            0.1,
            0.25,
            0.5,
            1,
            5
          ]
        }
      ]
    },
    "exporter": {
      "protocol": "otp",
      "collector": {
        "addr": "",
        "telemetry_data": 1,
        "data_protocol": 5,
        "data_transmission": 1
      },
      "thread_count": 10,
      "buffer_size": 10000,
      "page_size": 1000,
      "timeout_ms": 2000,
      "window_seconds": 10
    }
  },
  "traces_config": {
    "enable": true,
    "processor": {
      "protocol": "omp",
      "sampler": {
        "enable": true,
        "fraction": 0.0001,
        "error_fraction": 1
      },
      "disable_trace_body": true,
      "deferred_sample_slow_duration_ms": 1000,
      "enable_profile": true
    },
    "exporter": {
      "protocol": "oltp",
      "collector": {
        "addr": "",
        "telemetry_data": 3,
        "data_protocol": 1,
        "data_transmission": 2
      }
    }
  },
  "logs_config": {
    "enable": true,
    "processor": {
      "protocol": "omp",
      "sampler": {
        "enable": false,
        "fraction": 1,
        "error_fraction": 1
      },
      "level": "error",
      "enable_recovery": true
    },
    "exporter": {
      "protocol": "oltp",
      "collector": {
        "addr": "",
        "telemetry_data": 2,
        "data_protocol": 1,
        "data_transmission": 2
      }
    }
  },
  "profiles_config": {
    "enable": true,
    "processor": {
      "protocol": "omp",
      "profile_types": ["CPU", "heap"],
      "period_seconds": 60,
      "cpu_duration_seconds": 60,
      "cpu_profile_rate": 100,
      "mutex_profile_fraction": 10,
      "block_profile_rate": 100000000,
      "enable_delta_profiles": true,
      "enable_link_trace": false
    },
    "exporter": {
      "protocol": "otp",
      "collector": {
        "addr": "",
        "telemetry_data": 4,
        "data_protocol": 5,
        "data_transmission": 1
      },
      "export_to_file": false
    }
  },
  "tenant_id": "galileo"
}`), &rsp)
	return &rsp
}

func newTestServer() (*httptest.Server, *model.GetConfigResponse) {
	rsp := testResponse()
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				body, _ := json.Marshal(rsp)
				_, _ = w.Write(body)
			},
		),
	)
	return ts, rsp
}

// Test_getOcpConfig ocp 接口协议测试。
// 使用本地 testServer 模拟服务端返回。
func Test_getOcpConfig(t *testing.T) {
	testServer, _ := newTestServer()
	defer testServer.Close()
	type args struct {
		ocpURL string
	}
	ocps := []args{
		{testServer.URL},
	}
	for i := range ocps {
		ocp := ocps[i].ocpURL
		t.Run(
			ocp, func(t *testing.T) {
				configResponse, err := GetOcpConfig(
					ocp, &model.Resource{Platform: DefaultPlatform, ObjectName: "example.greeter"},
				)
				assert.Nil(t, err)
				t.Logf("configResponse=%v, err=%v", configResponse, err)
				assert.Equal(t, int32(0), configResponse.Code)
				expected := &model.GetConfigResponse{
					Code: 0, Msg: "success",
					Target:         "PCG-123.example.greeter",
					ConfigServer:   "",
					RegisterServer: "", SelfMonitor: model.SelfMonitor{
						Protocol: "otp", Collector: model.Collector{
							Addr: "", TelemetryData: 1,
							DataProtocol:     5,
							DataTransmission: 1, Version: 0,
						},
					}, MetricsConfig: model.MetricsConfig{
						Enable: true, Processor: model.MetricsProcessor{
							Protocol: "omp", WindowSeconds: 20, ClearSeconds: 3600, ExpiresSeconds: 3600,
							PointLimit:           200000,
							EnableProcessMetrics: true, ProcessMetricsSeconds: 20,
							HistogramBuckets: []model.HistogramBucket{
								{
									Name:    "rpc_server_handled_seconds",
									Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 5},
								},
								{
									Name:    "rpc_client_handled_seconds",
									Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 5},
								},
							},
						}, Exporter: model.MetricsExporter{
							Protocol: "otp", Collector: model.Collector{
								Addr: "", TelemetryData: 1,
								DataProtocol:     5,
								DataTransmission: 1, Version: 0,
							}, ThreadCount: 10, BufferSize: 10000, PageSize: 1000, TimeoutMs: 2000, WindowSeconds: 10,
						},
					}, TracesConfig: model.TracesConfig{
						Enable: true, Processor: model.TracesProcessor{
							Protocol:         "omp",
							Sampler:          model.SamplerConfig{Enable: true, Fraction: 0.0001, ErrorFraction: 1},
							DisableTraceBody: true, EnableDeferredSample: false, DeferredSampleError: false,
							DeferredSampleSlowDurationMs: 1000, DisableParentSampling: false, EnableProfile: true,
						}, Exporter: model.TracesExporter{
							Protocol: "oltp", Collector: model.Collector{
								Addr: "", TelemetryData: 3, DataProtocol: 1, DataTransmission: 2,
								Version: 0,
							},
						},
					}, LogsConfig: model.LogsConfig{
						Enable: true, Processor: model.LogsProcessor{
							Protocol:     "omp",
							OnlyTraceLog: false,
							TraceLogMode: 0, Level: "error", EnableRecovery: true,
						}, Exporter: model.LogsExporter{
							Protocol: "oltp", Collector: model.Collector{
								Addr: "", TelemetryData: 2, DataProtocol: 1, DataTransmission: 2,
								Version: 0,
							},
						},
					}, ProfilesConfig: model.ProfilesConfig{
						Enable: true, Processor: model.ProfilesProcessor{
							Protocol:             "omp",
							ProfileTypes:         []string{"CPU", "heap"},
							PeriodSeconds:        60,
							CpuDurationSeconds:   60,
							CpuProfileRate:       100,
							MutexProfileFraction: 10,
							BlockProfileRate:     100000000,
							EnableDeltaProfiles:  true,
						}, Exporter: model.ProfilesExporter{
							Protocol: "otp", Collector: model.Collector{
								Addr:             "",
								TelemetryData:    4,
								DataProtocol:     5,
								DataTransmission: 1,
								Version:          0,
							},
							ExportToFile: false,
						},
					}, TenantId: "galileo", Version: 0,
				}
				assert.Equal(t, expected, configResponse)
				bytes, err := yaml.Marshal(configResponse)
				assert.Nil(t, err)
				assert.True(t, len(bytes) > 0)
			},
		)
	}
}
