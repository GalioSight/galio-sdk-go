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
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
)

type httpOption struct {
	headers map[string]string
}

type option func(*httpOption)

// WithHeaders 设置请求头
func WithHeaders(headers map[string]string) option {
	return func(o *httpOption) {
		o.headers = headers
	}
}

// Post 用 post 方式发送 HTTP 请求。
func Post(client *http.Client, url string, data []byte, opts ...option) ([]byte, error) {
	r := bytes.NewReader(data)
	req, err := http.NewRequest("POST", url, r)
	if err != nil {
		return nil, fmt.Errorf("new httputil request err, %w", err)
	}

	httpOpt := httpOption{}
	for _, o := range opts {
		o(&httpOpt)
	}

	// set header
	for k, v := range httpOpt.headers {
		req.Header.Set(k, v)
	}

	rsp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("httputil Do err: %w", err)
	}
	defer func() { _ = rsp.Body.Close() }()
	if rsp.StatusCode != http.StatusOK {
		rspBytes, _ := io.ReadAll(rsp.Body)
		return rspBytes, fmt.Errorf("httpStatusCodeErr: %v, rsp=%v", rsp.StatusCode, string(rspBytes))
	}
	buf, err := io.ReadAll(rsp.Body)
	return buf, err
}

// IsNetOpError 判断一个 err 是否是网络错误。
func IsNetOpError(err error) bool {
	err = errors.Unwrap(err)
	err = errors.Unwrap(err)
	var opError *net.OpError
	return errors.As(err, &opError)
}
