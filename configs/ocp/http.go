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

package ocp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	httputil "galiosight.ai/galio-sdk-go/lib/http"
	"galiosight.ai/galio-sdk-go/self/log"
)

// invoke 通过 post 方式调用 HTTP 接口，不使用 trpc-go 框架的调用方式，避免依赖 trpc-go。
func invoke(url string, req interface{}, timeout time.Duration, rsp interface{}, headers map[string]string) error {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		log.Errorf("[galileo]invoke|Marshal|req=%v,err=%v", req, err)
		return fmt.Errorf("marshal: %w", err)
	}
	var data []byte
	data, err = post(url, reqBytes, timeout, headers)
	if err != nil {
		log.Errorf("[galileo]invoke|post|url=%v,req=%v,err=%v", url, string(reqBytes), err)
		return fmt.Errorf("post: %w", err)
	}
	err = json.Unmarshal(data, rsp)
	if err != nil {
		log.Errorf("[galileo]invoke|Unmarshal|rsp=%v,err=%v", string(data), err)
		return fmt.Errorf("unmarshal: %w", err)
	}
	return nil
}

// post 封装了将一个 JSON post 到一个 HTTP 服务器的操作。
// url HTTP 服务器，
// JSON 要发送的数据。
// maxDataSize 接收的最大数据量。
// timeoutMs 超时时间。
func post(url string, json []byte, timeout time.Duration, headers map[string]string) (data []byte, err error) {
	var client = &http.Client{
		Timeout:   timeout,
		Transport: &http.Transport{DisableKeepAlives: true},
	}
	return httputil.Post(client, url, json, httputil.WithHeaders(headers))
}
