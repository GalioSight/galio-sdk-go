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

package traces

import (
	omp3 "galiosight.ai/galio-sdk-go/semconv"
	semconv "galiosight.ai/galio-sdk-go/semconv/v1.0.0"
)

const (
	// 兼容外部编译
	NamespaceKey     = semconv.TrpcNamespaceKey
	EnvNameKey       = semconv.TrpcEnvnameKey
	StatusCode       = semconv.TrpcStatusCodeKey
	StatusMsg        = semconv.TrpcStatusMsgKey
	StatusType       = semconv.TrpcStatusTypeKey
	ProtocolKey      = semconv.TrpcProtocolKey
	CallerServiceKey = semconv.TrpcCallerServiceKey
	CallerMethodKey  = semconv.TrpcCallerMethodKey
	CallerServerKey  = semconv.TrpcCallerServerKey
	CalleeServiceKey = semconv.TrpcCalleeServiceKey
	CalleeMethodKey  = semconv.TrpcCalleeMethodKey
	CalleeServerKey  = semconv.TrpcCalleeServerKey
	ForceSamplerKey  = omp3.TraceForceSampleKey
)
