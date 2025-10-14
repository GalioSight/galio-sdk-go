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

// Package ocp Ocp 协议远程配置
package ocp

import (
	"fmt"
	"time"

	"galiosight.ai/galio-sdk-go/errs"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/self/log"
	"github.com/gogo/protobuf/proto"
)

const (
	// timeout 访问 ocp 服务的默认超时时间。
	timeout = time.Second * 2

	// totalTimeout 插件加载总超时时间。
	totalTimeout = timeout * 4

	// retCodeOK ocp 服务接口成功时的 code
	retCodeOK = 0

	// DefaultPlatform 默认平台
	DefaultPlatform = "PCG-123"
)

const (
	// DefaultURL 默认的 ocp 地址，是当前伽利略实现的 ocp 服务地址。
	DefaultURL = ""
)

type options struct {
	local  *model.GetConfigResponse
	apiKey string
}

type option func(*options)

// Local 指定本地配置
func Local(local *model.GetConfigResponse) option {
	return func(o *options) {
		o.local = local
	}
}

// WithApiKey 设置 apiKey
func WithApiKey(apiKey string) option {
	return func(o *options) {
		o.apiKey = apiKey
	}
}

// GetOcpConfig 从 Ocp server 获取配置
// 理论上 ocp 会将 local 配置与远程配置合并之后返回。
// 如果 ocp 协议发生了变化，SDK 新增了字段，需要注意发布 ocp 服务，保持 proto 协议是最新版本，否则远程接口返回的字段会不完整。
func GetOcpConfig(url string, resource *model.Resource, opts ...option) (*model.GetConfigResponse, error) {
	if url == "" {
		return nil, errs.ErrOcpInvalid
	}
	var o options
	for _, f := range opts {
		f(&o)
	}
	req := &model.GetConfigRequest{
		Platform:   resource.Platform,
		ObjectName: resource.ObjectName,
		Metrics: model.CollectorProtocol{
			TelemetryData:    model.TelemetryData_TELEMETRY_DATA_METRICS,
			DataProtocol:     model.DataProtocol_DATA_PROTOCOL_OTP,
			DataTransmission: model.DataTransmission_DATA_TRANSMISSION_HTTP,
		},
		Traces: model.CollectorProtocol{
			TelemetryData:    model.TelemetryData_TELEMETRY_DATA_TRACES,
			DataProtocol:     model.DataProtocol_DATA_PROTOCOL_OTLP,
			DataTransmission: model.DataTransmission_DATA_TRANSMISSION_gRPC,
		},
		Logs: model.CollectorProtocol{
			TelemetryData:    model.TelemetryData_TELEMETRY_DATA_LOGS,
			DataProtocol:     model.DataProtocol_DATA_PROTOCOL_OTLP,
			DataTransmission: model.DataTransmission_DATA_TRANSMISSION_gRPC,
		},
		Profiles: model.CollectorProtocol{
			TelemetryData:    model.TelemetryData_TELEMETRY_DATA_PROFILES,
			DataProtocol:     model.DataProtocol_DATA_PROTOCOL_OTP,
			DataTransmission: model.DataTransmission_DATA_TRANSMISSION_HTTP,
		},
		Env:      resource.EnvName,
		Set:      resource.SetName,
		Resource: *resource,
		Local:    o.local,
	}
	rsp := &model.GetConfigResponse{}
	if o.local != nil {
		// 复制本地配置对象作为结果对象，然后再将远程结果合并进去。
		// 需要深拷贝 local，因为 local 作为本地配置不应该被修改。
		// 如果不复制的话，则结果里面只有远程配置，没有本地配置。如果远程配置返回的的配置项不完整，会导致结果里面缺少本地配置。
		// 之前是依赖在 Update 中，本地配置优先时，使用本地配置再次覆盖。
		// 由于 ocp 当前返回的远程配置版本一定大于本地配置，导致本地配置覆盖分支无法生效。
		localConfigCopy, ok := proto.Clone(o.local).(*model.GetConfigResponse)
		if ok {
			*rsp = *localConfigCopy
		}
	}
	err := retry(
		func() error {
			return invoke(
				url, req, timeout, rsp, map[string]string{model.APIKeyHeaderKey: o.apiKey},
			)
		}, totalTimeout, 5,
	)
	if err != nil {
		log.Errorf("[galileo]getOcpConfig|req=%v,err=%v", req, err)
		return nil, err
	}
	if rsp.Code != retCodeOK {
		log.Errorf("[galileo]getOcpConfig|req=%v,rsp.Code=%v,msg=%v", req, rsp.Code, rsp.Msg)
		return nil, fmt.Errorf("code=%v,msg=%v", rsp.Code, rsp.Msg)
	}
	return rsp, nil
}
