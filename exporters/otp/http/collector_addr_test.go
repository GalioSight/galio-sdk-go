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

package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCollectorAddr(t *testing.T) {
	t.Run(
		"有效的URL，无IP - 连接失败", func(t *testing.T) {
			fullURL := "http://a.b.c/d/e/fg"

			// 模拟所有连接失败
			old := isConnectionSuccessful
			isConnectionSuccessful = func(ipPort string) bool {
				return false
			}
			defer func() { isConnectionSuccessful = old }()

			c := NewCollectorAddr(fullURL, nil)
			c.initEnableDirectIP()
			assert.Equal(t, fullURL, c.FullURL)
			assert.ElementsMatch(t, []string{}, c.directIPs)
			assert.Equal(t, false, c.enableDirectIP)
		},
	)

	t.Run(
		"有效的URL，有IP - 所有连接失败", func(t *testing.T) {
			fullURL := "http://a.b.c/d/e/fg"
			directIPPorts := []string{"0.0.0.0:0", "0.0.0.0:1"}

			// 模拟所有连接失败
			old := isConnectionSuccessful
			isConnectionSuccessful = func(ipPort string) bool {
				return false
			}
			defer func() { isConnectionSuccessful = old }()

			c := NewCollectorAddr(fullURL, directIPPorts)
			c.initEnableDirectIP()
			assert.Equal(t, fullURL, c.FullURL)
			assert.ElementsMatch(t, []string{"http://0.0.0.0:0/d/e/fg", "http://0.0.0.0:1/d/e/fg"}, c.directIPs)
			assert.Equal(t, false, c.enableDirectIP)
		},
	)

	t.Run(
		"有效的URL，有IP - 一个连接成功", func(t *testing.T) {
			fullURL := "http://a.b.c/d/e/fg"
			directIPPorts := []string{"127.0.0.1:8080", "0.0.0.0:1"}

			// 模拟只有一个 IP 成功
			old := isConnectionSuccessful
			isConnectionSuccessful = func(ipPort string) bool {
				return strings.Contains(ipPort, "127.0.0.1:8080")
			}
			defer func() { isConnectionSuccessful = old }()

			c := NewCollectorAddr(fullURL, directIPPorts)
			c.initEnableDirectIP()
			assert.Equal(t, fullURL, c.FullURL)
			assert.ElementsMatch(t, []string{"http://127.0.0.1:8080/d/e/fg", "http://0.0.0.0:1/d/e/fg"}, c.directIPs)
			assert.Equal(t, true, c.enableDirectIP)
		},
	)
}

func TestCollectorAddr_UpdateIPPorts(t *testing.T) {
	tests := []struct {
		name           string
		initialFullURL string
		initialIPs     []string
		newFullURL     string
		newIPs         []string
		wantFullURL    string
		wantDirectIPs  []string
	}{
		{
			name:           "Update with new IPs",
			initialFullURL: "http://a.b.c/d/e/fg",
			initialIPs:     []string{"0.0.0.0:0"},
			newFullURL:     "http://a.b.c/d/e/fg",
			newIPs:         []string{"1.1.1.1:1"},
			wantFullURL:    "http://a.b.c/d/e/fg",
			wantDirectIPs:  []string{"http://1.1.1.1:1/d/e/fg"},
		},
		{
			name:           "Update with empty IPs",
			initialFullURL: "http://a.b.c/d/e/fg",
			initialIPs:     []string{"0.0.0.0:0"},
			newFullURL:     "http://a.b.c/d/e/fg",
			newIPs:         []string{},
			wantFullURL:    "http://a.b.c/d/e/fg",
			wantDirectIPs:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := &CollectorAddr{
					FullURL:   tt.initialFullURL,
					directIPs: tt.initialIPs,
				}
				c.UpdateIPPorts(tt.newFullURL, tt.newIPs)
				assert.Equal(t, tt.wantFullURL, c.FullURL)
				assert.ElementsMatch(t, tt.wantDirectIPs, c.directIPs)
			},
		)
	}
}

func TestCollectorAddr_GetAddr(t *testing.T) {
	tests := []struct {
		name           string
		fullURL        string
		directIPs      []string
		idx            int
		retryIdx       int
		want           string
		enableDirectIP bool
	}{
		{"0 个 IP, enableDirectIP: false", "http://a.b.c/d/e/fg", nil, 0, 0, "http://a.b.c/d/e/fg", false},
		{"0 个 IP, enableDirectIP: true", "http://a.b.c/d/e/fg", nil, 0, 0, "http://a.b.c/d/e/fg", true},
		{
			"3 个 IP-0, enableDirectIP: false", "http://a.b.c/d/e/fg", []string{"0.0.0.0:0", "0.0.0.0:1", "0.0.0.0:2"},
			0, 0,
			"http://a.b.c/d/e/fg", false,
		}, {
			"3 个 IP-0, enableDirectIP: false", "http://a.b.c/d/e/fg", []string{"0.0.0.0:0", "0.0.0.0:1", "0.0.0.0:2"},
			0, 0,
			"http://0.0.0.0:1/d/e/fg", true,
		},
		{
			"3 个 IP-1, enableDirectIP: true", "http://a.b.c/d/e/fg", []string{"0.0.0.0:0", "0.0.0.0:1", "0.0.0.0:2"},
			1, 0,
			"http://0.0.0.0:2/d/e/fg", true,
		},
		{
			"3 个 IP-1, enableDirectIP: true", "http://a.b.c/d/e/fg", []string{"0.0.0.0:0", "0.0.0.0:1", "0.0.0.0:2"},
			1, 1, "http://a.b.c/d/e/fg", true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := &CollectorAddr{
					FullURL:        tt.fullURL,
					directIPs:      nil,
					enableDirectIP: tt.enableDirectIP,
				}
				c.UpdateIPPorts(tt.fullURL, tt.directIPs)
				c.idx.Store(int32(tt.idx))
				assert.Equal(t, tt.want, c.GetAddr(tt.retryIdx))
			},
		)
	}
}

func TestIsConnectionSuccessful(t *testing.T) {
	t.Run(
		"successful connection", func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
						_, _ = w.Write([]byte("body"))
					},
				),
			)
			defer server.Close()

			assert.True(t, isConnectionSuccessful(server.URL))
		},
	)

	t.Run(
		"404 status code", func(t *testing.T) {
			notFoundServer := httptest.NewServer(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusNotFound)
					},
				),
			)
			defer notFoundServer.Close()

			assert.False(t, isConnectionSuccessful(notFoundServer.URL))
		},
	)

	t.Run(
		"server closed", func(t *testing.T) {
			assert.False(t, isConnectionSuccessful("http://localhost:12345/abcde"))
		},
	)

	t.Run(
		"500 status code", func(t *testing.T) {
			errorServer := httptest.NewServer(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusInternalServerError)
					},
				),
			)
			defer errorServer.Close()

			assert.True(t, isConnectionSuccessful(errorServer.URL))
		},
	)

	t.Run(
		"200 empty response body", func(t *testing.T) {
			errorReadServer := httptest.NewServer(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					},
				),
			)
			defer errorReadServer.Close()

			assert.True(t, isConnectionSuccessful(errorReadServer.URL))
		},
	)
}
