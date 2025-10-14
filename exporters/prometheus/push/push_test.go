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

package push

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"galiosight.ai/galio-sdk-go/model"
	omp3 "galiosight.ai/galio-sdk-go/semconv"
	semconv "galiosight.ai/galio-sdk-go/semconv/v1.0.0"
)

func TestPrometheusPush(t *testing.T) {
	// 创建一个测试服务器
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				// 打印请求的详细信息
				t.Logf("Request URL: %s", r.URL.String())
				t.Logf("Request Method: %s", r.Method)
				t.Logf("Request Headers: %v", r.Header)
				t.Logf("Request Body: %s", readBody(r))

				// 这里可以检查请求的内容，例如头部信息和主体
				author := r.Header.Get("Authorization")
				if author != basicAuth("username", "password") {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				schema := r.Header.Get(model.SchemaURLHeaderKey)
				if schema != semconv.SchemaURL && schema != omp3.SchemaURL {
					http.Error(w, "Invalid value for header: "+model.SchemaURLHeaderKey, http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer ts.Close()

	res := model.Resource{
		Target:        "PCG-123.example.greeter",
		Namespace:     "Production",
		EnvName:       "formal",
		Instance:      "127.0.0.1",
		ContainerName: "formal.ContainerName.0",
		TenantId:      "galileo",
	}

	cfg := model.PrometheusPushConfig{
		Enable:       true,
		Url:          ts.URL,
		Job:          "testjob",
		Interval:     1,
		UseBasicAuth: true,
		Username:     "username",
		Password:     "password",
		Grouping: map[string]string{
			"foo": "bar",
		},
		HttpHeaders: nil,
	}

	t.Run(
		"normal", func(t *testing.T) {
			cancel, err := PrometheusPush(res, cfg)
			assert.NoError(t, err)
			// 等一下，看定时上报是否正常
			time.Sleep(time.Second * 2)
			cancel()
			// 等一下，看定时上报是否取消
			time.Sleep(time.Millisecond * 200)
		},
	)

	t.Run(
		"Interval", func(t *testing.T) {
			cfg2 := cfg
			cfg2.Interval = 0
			cancel, err := PrometheusPush(res, cfg2)
			assert.Equal(t, ErrCfgInterval, err)
			assert.Nil(t, cancel)
		},
	)

	t.Run(
		"authorization", func(t *testing.T) {
			cfg2 := cfg
			cfg2.Password = "abc"
			cancel, err := PrometheusPush(res, cfg2)
			assert.Error(t, err)
			assert.True(t, strings.Contains(err.Error(), strconv.Itoa(http.StatusUnauthorized)))
			assert.Nil(t, cancel)
		},
	)

	t.Run(
		"Enabled", func(t *testing.T) {
			cfg2 := cfg
			cfg2.Enable = false
			cancel, err := PrometheusPush(res, cfg2)
			assert.Equal(t, ErrCfgEnabled, err)
			assert.Nil(t, cancel)
		},
	)

	t.Run("omp-v3", func(t *testing.T) {
		cancel, err := PrometheusPush(res, cfg, WithSchemaURL(omp3.SchemaURL))
		assert.NoError(t, err)
		// 等一下，看定时上报是否正常
		time.Sleep(time.Second * 2)
		cancel()
		// 等一下，看定时上报是否取消
		time.Sleep(time.Millisecond * 200)
	})
}

// basicAuth 创建一个基本的 HTTP 认证字符串
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

// readBody 读取并返回请求的主体内容
func readBody(r *http.Request) string {
	bodyBytes, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // 重置请求体，以便后续处理
	return string(bodyBytes)
}
