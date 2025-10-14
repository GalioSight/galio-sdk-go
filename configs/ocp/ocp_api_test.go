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

package ocp

import (
	"testing"

	"galiosight.ai/galio-sdk-go/model"
	"github.com/stretchr/testify/assert"
)

func TestOcpApiTestSite(t *testing.T) {
	local := model.GetConfigResponse{}
	configResponse, err := GetOcpConfig(
		"",
		&model.Resource{Platform: DefaultPlatform, ObjectName: "example.greeter", Version: "v0.14.1"},
		Local(&local),
	)
	assert.Nil(t, err)
	assert.Equal(
		t, "", configResponse.MetricsConfig.Exporter.Collector.Addr,
	)
	assert.Equal(t, "", configResponse.TracesConfig.Exporter.Collector.Addr)
	assert.Equal(t, "", configResponse.LogsConfig.Exporter.Collector.Addr)
	assert.Equal(
		t, "", configResponse.ProfilesConfig.Exporter.Collector.Addr,
	)
	assert.NotNil(t, configResponse.MetricsConfig.Exporter.Collector.DirectIpPort)
	t.Logf("configResponse=%v, err=%v", configResponse, err)
}

func TestOcpApiTestSite_CN_PRIVATE(t *testing.T) {
	local := model.GetConfigResponse{
		AccessPoint: model.AccessPoint_ACCESS_POINT_CN_PRIVATE,
	}
	configResponse, err := GetOcpConfig(
		"",
		&model.Resource{Platform: DefaultPlatform, ObjectName: "example.greeter", Version: "v0.14.1"},
		Local(&local),
	)
	assert.Nil(t, err)
	assert.Equal(
		t, "", configResponse.MetricsConfig.Exporter.Collector.Addr,
	)
	assert.Equal(t, "", configResponse.TracesConfig.Exporter.Collector.Addr)
	assert.Equal(t, "", configResponse.LogsConfig.Exporter.Collector.Addr)
	assert.Equal(
		t, "", configResponse.ProfilesConfig.Exporter.Collector.Addr,
	)
	assert.NotNil(t, configResponse.MetricsConfig.Exporter.Collector.DirectIpPort)
	t.Logf("configResponse=%v, err=%v", configResponse, err)
}

func TestOcpApiTestSite_SG_PRIVATE(t *testing.T) {
	local := model.GetConfigResponse{
		AccessPoint: model.AccessPoint_ACCESS_POINT_SG_PRIVATE,
	}
	configResponse, err := GetOcpConfig(
		"",
		&model.Resource{Platform: DefaultPlatform, ObjectName: "example.greeter", Version: "v0.14.1"},
		Local(&local),
	)
	assert.Nil(t, err)
	assert.Equal(
		t, "",
		configResponse.MetricsConfig.Exporter.Collector.Addr,
	)
	assert.Equal(t, "", configResponse.TracesConfig.Exporter.Collector.Addr)
	assert.Equal(t, "", configResponse.LogsConfig.Exporter.Collector.Addr)
	assert.Equal(
		t, "",
		configResponse.ProfilesConfig.Exporter.Collector.Addr,
	)
	assert.NotNil(t, configResponse.MetricsConfig.Exporter.Collector.DirectIpPort)
	t.Logf("configResponse=%v, err=%v", configResponse, err)
}

func TestOcpApiTestSite_CN_PUBLIC(t *testing.T) {
	local := model.GetConfigResponse{
		AccessPoint: model.AccessPoint_ACCESS_POINT_CN_PUBLIC,
	}
	configResponse, err := GetOcpConfig(
		"",
		&model.Resource{Platform: DefaultPlatform, ObjectName: "example.greeter", Version: "v0.14.1"},
		Local(&local),
	)
	assert.Nil(t, err)
	assert.Equal(
		t, "https://galileotelemetry.tencent.com/api/v1/metrics/write",
		configResponse.MetricsConfig.Exporter.Collector.Addr,
	)
	assert.Equal(t, "galileotelemetry.tencent.com", configResponse.TracesConfig.Exporter.Collector.Addr)
	assert.Equal(t, "galileotelemetry.tencent.com", configResponse.LogsConfig.Exporter.Collector.Addr)
	assert.Equal(
		t, "https://galileotelemetry.tencent.com/api/v1/profile/write",
		configResponse.ProfilesConfig.Exporter.Collector.Addr,
	)
	assert.Nil(t, configResponse.MetricsConfig.Exporter.Collector.DirectIpPort)
	t.Logf("configResponse=%v, err=%v", configResponse, err)
}

func TestOcpApiTestSite_SG_PUBLIC(t *testing.T) {
	local := model.GetConfigResponse{
		AccessPoint: model.AccessPoint_ACCESS_POINT_SG_PUBLIC,
	}
	configResponse, err := GetOcpConfig(
		"",
		&model.Resource{Platform: DefaultPlatform, ObjectName: "example.greeter", Version: "v0.14.1"},
		Local(&local),
	)
	assert.Nil(t, err)
	assert.Equal(
		t, "https://sg.galileotelemetry.tencent.com/api/v1/metrics/write",
		configResponse.MetricsConfig.Exporter.Collector.Addr,
	)
	assert.Equal(t, "sg.galileotelemetry.tencent.com", configResponse.TracesConfig.Exporter.Collector.Addr)
	assert.Equal(t, "sg.galileotelemetry.tencent.com", configResponse.LogsConfig.Exporter.Collector.Addr)
	assert.Equal(
		t, "https://sg.galileotelemetry.tencent.com/api/v1/profile/write",
		configResponse.ProfilesConfig.Exporter.Collector.Addr,
	)
	assert.Nil(t, configResponse.MetricsConfig.Exporter.Collector.DirectIpPort)
	t.Logf("configResponse=%v, err=%v", configResponse, err)
}

func TestOcpApiPre(t *testing.T) {
	configResponse, err := GetOcpConfig(
		"",
		&model.Resource{Platform: DefaultPlatform, ObjectName: "example.greeter", Version: "v0.3.21"},
	)
	assert.Nil(t, err)
	assert.Equal(t, "", configResponse.LogsConfig.Exporter.Collector.Addr)
	t.Logf("configResponse=%v, err=%v", configResponse, err)
}

func TestOcpApiFormal_CN_PRIVATE(t *testing.T) {
	local := model.GetConfigResponse{
		AccessPoint: model.AccessPoint_ACCESS_POINT_CN_PRIVATE,
	}
	configResponse, err := GetOcpConfig(
		DefaultURL,
		&model.Resource{Platform: DefaultPlatform, ObjectName: "ilive.ilive_pid_indicators_push", Version: "v0.3.21"},
		Local(&local),
	)
	assert.Nil(t, err)
	assert.Equal(
		t, "",
		configResponse.MetricsConfig.Exporter.Collector.Addr,
	)
	assert.Equal(t, "", configResponse.LogsConfig.Exporter.Collector.Addr)
	assert.Equal(t, "", configResponse.TracesConfig.Exporter.Collector.Addr)
	assert.Equal(
		t, "",
		configResponse.ProfilesConfig.Exporter.Collector.Addr,
	)
	assert.NotNil(t, configResponse.MetricsConfig.Exporter.Collector.DirectIpPort)
	t.Logf("configResponse=%#v, err=%v", configResponse, err)
}

func TestOcpApiFormal_CN_PUBLIC(t *testing.T) {
	local := model.GetConfigResponse{
		AccessPoint: model.AccessPoint_ACCESS_POINT_CN_PUBLIC,
	}
	configResponse, err := GetOcpConfig(
		"https://galileotelemetry.tencent.com/ocp/api/v1/get_config",
		&model.Resource{Platform: DefaultPlatform, ObjectName: "ilive.ilive_pid_indicators_push", Version: "v0.3.21"},
		Local(&local),
	)
	assert.Nil(t, err)
	assert.Equal(
		t, "https://galileotelemetry.tencent.com/api/v1/metrics/write",
		configResponse.MetricsConfig.Exporter.Collector.Addr,
	)
	assert.Equal(t, "galileotelemetry.tencent.com", configResponse.LogsConfig.Exporter.Collector.Addr)
	assert.Equal(t, "galileotelemetry.tencent.com", configResponse.TracesConfig.Exporter.Collector.Addr)
	assert.Equal(
		t, "https://galileotelemetry.tencent.com/api/v1/profile/write",
		configResponse.ProfilesConfig.Exporter.Collector.Addr,
	)
	assert.NotNil(t, configResponse.MetricsConfig.Exporter.Collector.DirectIpPort)
	t.Logf("configResponse=%#v, err=%v", configResponse, err)
}

func TestOcpApiFormal_SG_PRIVATE(t *testing.T) {
	local := model.GetConfigResponse{
		AccessPoint: model.AccessPoint_ACCESS_POINT_SG_PRIVATE,
	}
	configResponse, err := GetOcpConfig(
		"",
		&model.Resource{Platform: DefaultPlatform, ObjectName: "ilive.ilive_pid_indicators_push", Version: "v0.3.21"},
		Local(&local),
	)
	assert.Nil(t, err)
	assert.Equal(
		t, "",
		configResponse.MetricsConfig.Exporter.Collector.Addr,
	)
	assert.Equal(t, "", configResponse.LogsConfig.Exporter.Collector.Addr)
	assert.Equal(t, "", configResponse.TracesConfig.Exporter.Collector.Addr)
	assert.Equal(
		t, "",
		configResponse.ProfilesConfig.Exporter.Collector.Addr,
	)
	assert.NotNil(t, configResponse.MetricsConfig.Exporter.Collector.DirectIpPort)
	t.Logf("configResponse=%#v, err=%v", configResponse, err)
}

func TestOcpApiFormal_SG_PUBLIC(t *testing.T) {
	local := model.GetConfigResponse{
		AccessPoint: model.AccessPoint_ACCESS_POINT_SG_PUBLIC,
	}
	configResponse, err := GetOcpConfig(
		"https://sg.galileotelemetry.tencent.com/ocp/api/v1/get_config",
		&model.Resource{Platform: DefaultPlatform, ObjectName: "ilive.ilive_pid_indicators_push", Version: "v0.3.21"},
		Local(&local),
	)
	assert.Nil(t, err)
	assert.Equal(
		t, "https://sg.galileotelemetry.tencent.com/api/v1/metrics/write",
		configResponse.MetricsConfig.Exporter.Collector.Addr,
	)
	assert.Equal(t, "sg.galileotelemetry.tencent.com", configResponse.LogsConfig.Exporter.Collector.Addr)
	assert.Equal(t, "sg.galileotelemetry.tencent.com", configResponse.TracesConfig.Exporter.Collector.Addr)
	assert.Equal(
		t, "https://sg.galileotelemetry.tencent.com/api/v1/profile/write",
		configResponse.ProfilesConfig.Exporter.Collector.Addr,
	)
	assert.Nil(t, configResponse.MetricsConfig.Exporter.Collector.DirectIpPort)
	t.Logf("configResponse=%#v, err=%v", configResponse, err)
}

func TestOcpApiFormal(t *testing.T) {
	configResponse, err := GetOcpConfig(
		DefaultURL,
		&model.Resource{Platform: DefaultPlatform, ObjectName: "ilive.ilive_pid_indicators_push", Version: "v0.3.21"},
	)
	assert.Nil(t, err)
	assert.Equal(
		t, "",
		configResponse.MetricsConfig.Exporter.Collector.Addr,
	)
	assert.Equal(t, "", configResponse.LogsConfig.Exporter.Collector.Addr)
	assert.Equal(t, "", configResponse.TracesConfig.Exporter.Collector.Addr)
	assert.Equal(
		t, "",
		configResponse.ProfilesConfig.Exporter.Collector.Addr,
	)
	assert.NotNil(t, configResponse.MetricsConfig.Exporter.Collector.DirectIpPort)
	t.Logf("configResponse=%#v, err=%v", configResponse, err)
}
