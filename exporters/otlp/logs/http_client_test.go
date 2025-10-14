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
	"bytes"
	"compress/gzip"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"galiosight.ai/galio-sdk-go/model"
	"github.com/stretchr/testify/assert"
	collectorlogpb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	common "go.opentelemetry.io/proto/otlp/common/v1"
	logpb "go.opentelemetry.io/proto/otlp/logs/v1"
	resource "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/protobuf/proto"
)

func TestLogClientExportWithResourceLogs(t *testing.T) {
	// 创建一个模拟的 HTTP 服务器
	server := httptest.NewUnstartedServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				// 检查请求头
				assert.Equal(t, "application/x-protobuf", r.Header.Get("Content-Type"))
				assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

				// 读取请求体
				buf, err := ioutil.ReadAll(r.Body)
				assert.NoError(t, err)

				// 解压缩请求体
				reader, err := gzip.NewReader(bytes.NewReader(buf))
				assert.NoError(t, err)
				defer reader.Close()

				// 反序列化请求体
				var req collectorlogpb.ExportLogsServiceRequest
				err = proto.Unmarshal(reader.Extra, &req)
				assert.NoError(t, err)

				// 构造响应
				res := &collectorlogpb.ExportLogsServiceResponse{}

				// 序列化响应
				responseData, err := proto.Marshal(res)
				assert.NoError(t, err)

				// 设置响应头
				w.Header().Set("Content-Type", "application/x-protobuf")
				w.WriteHeader(http.StatusOK)

				// 写入响应体
				_, err = w.Write(responseData)
				assert.NoError(t, err)
			},
		),
	)

	server.Start()
	defer server.Close()

	urls := []string{
		server.URL,
	}
	for _, url := range urls {
		// 创建测试数据
		req := &collectorlogpb.ExportLogsServiceRequest{
			ResourceLogs: []*logpb.ResourceLogs{
				{
					Resource: &resource.Resource{
						Attributes: []*common.KeyValue{
							{
								Key:   "containerName",
								Value: &common.AnyValue{Value: &common.AnyValue_StringValue{StringValue: "unit_test_containerName"}},
							},
							{
								Key:   "env",
								Value: &common.AnyValue{Value: &common.AnyValue_StringValue{StringValue: "unit_test"}},
							},
							{
								Key: "instance",
								Value: &common.AnyValue{
									Value: &common.
										AnyValue_StringValue{StringValue: "unit_test_instance"},
								},
							},
							{
								Key:   "namespace",
								Value: &common.AnyValue{Value: &common.AnyValue_StringValue{StringValue: "Development"}},
							},
							{
								Key:   "server",
								Value: &common.AnyValue{Value: &common.AnyValue_StringValue{StringValue: "example.greeter"}},
							},
							{
								Key: "target",
								Value: &common.AnyValue{
									Value: &common.AnyValue_StringValue{
										StringValue: "PCG-123.example.greeter",
									},
								},
							},
						},
						DroppedAttributesCount: 0,
					},
					ScopeLogs: []*logpb.ScopeLogs{
						{
							Scope: nil,
							LogRecords: []*logpb.LogRecord{
								{
									TimeUnixNano:         uint64(time.Now().UnixNano()),
									ObservedTimeUnixNano: uint64(time.Now().UnixNano()),
									SeverityNumber:       0,
									SeverityText:         "debug",
									Body: &common.AnyValue{
										Value: &common.
											AnyValue_StringValue{StringValue: "msg: test, url=" + url},
									},
									Attributes: []*common.KeyValue{
										{
											Key:   "sampled",
											Value: &common.AnyValue{Value: &common.AnyValue_BoolValue{BoolValue: true}},
										},
										{
											Key:   "abcde",
											Value: &common.AnyValue{Value: &common.AnyValue_StringValue{StringValue: "efg"}},
										},
										{
											Key:   "line",
											Value: &common.AnyValue{Value: &common.AnyValue_StringValue{StringValue: "unit_test_line"}},
										},
									},
									DroppedAttributesCount: 0,
									Flags:                  0,
									TraceId:                nil,
									SpanId:                 nil,
								},
							},
							SchemaUrl: "",
						},
					},
					SchemaUrl: "",
				},
			},
		}
		t.Run(
			url, func(t *testing.T) {
				// 创建 logClient 实例
				client := newLogClient(url, map[string]string{model.TenantHeaderKey: "galileo"})
				resp, err := client.Export(context.Background(), req)
				// 检查结果
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			},
		)
	}
}
