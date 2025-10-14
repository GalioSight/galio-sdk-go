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

package profiles

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"galiosight.ai/galio-sdk-go/configs"
	otphttp "galiosight.ai/galio-sdk-go/exporters/otp/http"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/version"
	"github.com/stretchr/testify/require"
)

var data = &model.ProfilesBatch{
	Sequence: 1,
	Start:    1672502400,
	End:      1672502460,
	Profiles: []*model.Profile{
		{
			Name: "cpu.pprof",
			Type: "cpu",
		},
	},
	Resource: &model.Resource{
		Target:        "PCG-123.example.greeter",  // 观测对象的唯一标识 ID，需要全局唯一
		Namespace:     "Development",              // 物理环境
		EnvName:       "test",                     // 用户环境
		Region:        "sz",                       // 地域
		Instance:      "aaa.bbb.ccc.ddd",          // 实例 ip
		Node:          "cls-as9z3nec-2",           // 节点
		ContainerName: "test.example.greeter.sz1", // 容器
		Version:       version.Number,             // SDK 版本号
		Platform:      "PCG-123",                  // 平台
		ObjectName:    "example.greeter",          // 对象名称
	},
}

func Test_profilesExporter_Export(t *testing.T) {
	var ts = httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"code":0,"msg":"success"}`))
			},
		),
	)
	defer ts.Close()
	exporter, err := NewExporter(
		&configs.Profiles{
			Exporter: model.ProfilesExporter{
				Protocol:     "otp",
				Collector:    model.Collector{Addr: ts.URL},
				BufferSize:   10,
				TimeoutMs:    10000,
				ExportToFile: false,
			},
			Stats: &model.SelfMonitorStats{},
		},
	)
	require.Nil(t, err)
	exporter.Export(data)
	time.Sleep(time.Second * 1)
	stats := exporter.(*profilesExporter).stats
	require.Equal(t, int64(len(data.Profiles)), stats.ProfilesStats.EnqueueCounter.Load())
	require.Equal(t, int64(0), stats.ProfilesStats.DropCounter.Load())
}

func Test_profilesExporter_Export_Error(t *testing.T) {
	exporter, err := NewExporter(
		&configs.Profiles{
			Exporter: model.ProfilesExporter{
				Protocol:     "otp",
				Collector:    model.Collector{Addr: ""},
				BufferSize:   10,
				TimeoutMs:    10000,
				ExportToFile: false,
			},
		},
	)
	require.Nil(t, err)
	exit := make(chan struct{})
	exporter.Export(data)
	time.Sleep(time.Second * 1)
	close(exit)
	stats := exporter.(*profilesExporter).stats
	require.Equal(t, int64(len(data.Profiles)), stats.ProfilesStats.EnqueueCounter.Load())
	require.Equal(t, int64(0), stats.ProfilesStats.DropCounter.Load())
	require.Equal(t, int64(len(data.Profiles)), stats.ProfilesStats.FailedExportCounter.Load())
	require.Less(t, int64(0), stats.ProfilesStats.FailedWriteByteSize.Load())
}

func Test_profilesExporter_UpdateConfig(t *testing.T) {
	var ts = httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"code":0,"msg":"success"}`))
			},
		),
	)
	defer ts.Close()
	cfg := &configs.Profiles{
		Exporter: model.ProfilesExporter{
			Protocol:   "otp",
			Collector:  model.Collector{Addr: ""},
			BufferSize: 10,
		},
	}
	exporter, err := NewExporter(
		cfg,
	)
	p, ok := exporter.(*profilesExporter)
	if !ok {
		t.Errorf("exproter should be profilesExporter")
	}
	t.Logf("exporter=%+v", p.cfg)
	require.Nil(t, err)

	cfg.Exporter.Collector.Addr = ts.URL
	exporter.UpdateConfig(cfg)
	require.Equal(t, ts.URL, p.httpExporter.(*otphttp.HTTPGeneralExporter).CollectorAddr.FullURL)
}
