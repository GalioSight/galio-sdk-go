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
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	collectorlogpb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	"google.golang.org/protobuf/proto"
)

type logClient struct {
	url     string
	headers map[string]string
}

func newLogClient(addr string, headers map[string]string) *logClient {
	return &logClient{
		url:     fmt.Sprintf("%s/v1/logs", GetFullDomain(addr)),
		headers: headers,
	}
}

// GetFullDomain 返回给定地址的完整域名。
// 如果地址已经以 http:// 或 https:// 开头，则直接返回该地址。
// 如果地址是内网域名（包含），则添加 http:// 前缀。
// 如果地址是外网域名（不包含），则添加 https:// 前缀。
// 这是因为伽利略的内网域名接入不支持 https，而外网接入只支持 https。
func GetFullDomain(addr string) string {
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return addr
	}
	protocol := "https://"
	return protocol + addr
}

func (l *logClient) Export(
	ctx context.Context, in *collectorlogpb.ExportLogsServiceRequest,
) (*collectorlogpb.ExportLogsServiceResponse, error) {
	buf, err := proto.Marshal(in)
	if err != nil {
		return nil, err
	}
	// 压缩数据
	var b bytes.Buffer
	zw := gzip.NewWriter(&b)
	if _, err = zw.Write(buf); err != nil {
		return nil, err
	}
	if err = zw.Close(); err != nil {
		return nil, err
	}

	// 设置 HTTP 请求
	req, err := http.NewRequestWithContext(
		ctx, "POST", l.url, &b,
	)
	if err != nil {
		return nil, err
	}

	// 设置 HTTP 头部
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Content-Encoding", "gzip") // 设置压缩头部
	for k, v := range l.headers {
		req.Header.Set(k, v)
	}
	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: time.Second * 5, // 超时时间 5 秒，且不允许用户配置
	}

	// 发送 HTTP 请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	return &collectorlogpb.ExportLogsServiceResponse{}, nil
}
