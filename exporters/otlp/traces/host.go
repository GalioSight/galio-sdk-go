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

// Package traces traces 子系统
package traces

import (
	"net"
	"os"
	"sync"

	"go.opentelemetry.io/otel/attribute"

	semconv "galiosight.ai/galio-sdk-go/semconv/v1.0.0"
)

var (
	hostname     string
	hostnameOnce sync.Once
)

const localhost = "127.0.0.1"

// PeerInfo 对方节点信息。
func PeerInfo(addr net.Addr) []attribute.KeyValue {
	if addr == nil {
		return nil
	}
	host, port, err := net.SplitHostPort(addr.String())

	if err != nil {
		return []attribute.KeyValue{}
	}

	if host == "" {
		host = localhost
	}

	return []attribute.KeyValue{
		semconv.NetPeerIPKey.String(host),
		semconv.NetPeerPortKey.String(port),
	}
}

// HostInfo 主机节点信息。
func HostInfo(addr net.Addr) []attribute.KeyValue {
	if addr == nil {
		return []attribute.KeyValue{
			semconv.NetHostNameKey.String(getHostname()),
		}
	}
	host, port, err := net.SplitHostPort(addr.String())

	if err != nil {
		return []attribute.KeyValue{
			semconv.NetHostNameKey.String(getHostname()),
		}
	}

	if host == "" {
		host = localhost
	}

	return []attribute.KeyValue{
		semconv.NetHostIPKey.String(host),
		semconv.NetHostPortKey.String(port),
		semconv.NetHostNameKey.String(getHostname()),
	}
}

// getHostname 主机名。
func getHostname() string {
	hostnameOnce.Do(
		func() {
			hostname, _ = os.Hostname()
		},
	)
	return hostname
}
