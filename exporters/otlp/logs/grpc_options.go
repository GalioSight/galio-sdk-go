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

package logs

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type grpcOptions struct {
	clientCredentials  credentials.TransportCredentials // grpc 证书对象。
	headers            map[string]string                // grpc 请求头。
	compressor         string                           // grpc 压缩类型。
	serviceConfig      string                           // grpc service 配置。
	addr               string                           // grpc 服务端地址。
	dialOptions        []grpc.DialOption                // grpc 连接参数。
	reconnectionPeriod time.Duration                    // grpc 客户端保活重连间隔。
	dialInsecure       bool                             // grpc 跳过对服务器证书的验证。
	httpEnabled        bool                             // 是否使用 http 协议
}

type grpcOption func(*grpcOptions)

// defaultGRPCServiceConfig is the gRPC service config used if none is
// provided by the user.
//
// For more info on gRPC service configs:
// https://github.com/grpc/proposal/blob/master/A6-client-retries.md
//
// Note: MaxAttempts > 5 are treated as 5. See
// https://github.com/grpc/proposal/blob/master/A6-client-retries.md#validation-of-retrypolicy
// for more details.
const defaultGRPCServiceConfig = `{
	"methodConfig":[{
		"name":[
			{ "service":"opentelemetry.proto.collector.metrics.v1.MetricsService" },
			{ "service":"opentelemetry.proto.collector.trace.v1.TraceService" }
		],
		"retryPolicy":{
			"MaxAttempts":5,
			"InitialBackoff":"0.3s",
			"MaxBackoff":"5s",
			"BackoffMultiplier":2,
			"RetryableStatusCodes":[
				"UNAVAILABLE",
				"CANCELLED",
				"DEADLINE_EXCEEDED",
				"RESOURCE_EXHAUSTED",
				"ABORTED",
				"OUT_OF_RANGE",
				"UNAVAILABLE",
				"DATA_LOSS"
			]
		}
	}]
}`

func newGRPCOptions(opts ...grpcOption) grpcOptions {
	o := grpcOptions{
		serviceConfig: defaultGRPCServiceConfig,
		httpEnabled:   true,
	}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// withInsecure disables client transport security for the exporter's gRPC connection
// just like grpc.withInsecure() https://pkg.go.dev/google.golang.org/grpc#withInsecure
// does. Note, by default, client security is required unless withInsecure is used.
func withInsecure() grpcOption {
	return func(o *grpcOptions) {
		o.dialInsecure = true
	}
}

// withAddress allows one to set the address that the exporter will
// connect to the collector on. If unset, it will instead try to use
// connect to DefaultCollectorHost:DefaultCollectorPort.
func withAddress(addr string) grpcOption {
	return func(o *grpcOptions) {
		o.addr = addr
	}
}

// withCompressor will set the compressor for the gRPC client to use when sending requests.
// It is the responsibility of the caller to ensure that the compressor set has been registered
// with google.golang.org/grpc/encoding. This can be done by encoding.RegisterCompressor. Some
// compressors auto-register on import, such as gzip, which can be registered by calling
// `import _ "google.golang.org/grpc/encoding/gzip"`
func withCompressor(compressor string) grpcOption {
	return func(o *grpcOptions) {
		o.compressor = compressor
	}
}

// withGRPCHeaders will send the provided headers with gRPC requests
func withGRPCHeaders(headers map[string]string) grpcOption {
	return func(o *grpcOptions) {
		o.headers = headers
	}
}

// withHTTPEnabled will send the provided headers with gRPC requests
func withHTTPEnabled(enable bool) grpcOption {
	return func(o *grpcOptions) {
		o.httpEnabled = enable
	}
}

// withTLSCredentials allows the connection to use TLS credentials
// when talking to the server. It takes in grpc.TransportCredentials instead
// of say a Certificate file or a tls.Certificate, because the retrieving
// these credentials can be done in many ways e.g. plain file, in code tls.Config
// or by certificate rotation, so it is up to the caller to decide what to use.
func withTLSCredentials(creds credentials.TransportCredentials) grpcOption {
	return func(o *grpcOptions) {
		o.clientCredentials = creds
	}
}

// withGRPCServiceConfig defines the default gRPC service config used.
func withGRPCServiceConfig(serviceConfig string) grpcOption {
	return func(o *grpcOptions) {
		o.serviceConfig = serviceConfig
	}
}

// withGRPCDialOption opens support to any grpc.DialOption to be used. If it conflicts
// with some other configuration the GRPC specified via the collector the ones here will
// take preference since they are set last.
func withGRPCDialOption(opts ...grpc.DialOption) grpcOption {
	return func(o *grpcOptions) {
		o.dialOptions = opts
	}
}
