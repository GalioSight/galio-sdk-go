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
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"go.uber.org/atomic"
)

// CollectorAddr 收集器地址，orderDirectURLs 会在运行中进行热更新。此对象是线程安全的。
type CollectorAddr struct {
	// 上报地址，完整的 url,
	FullURL string
	// 排序后的直连地址列表，通常按健康度排序，这样优先选择前面的地址进行上报。
	// 当前是随机排序，未来再进行算法上的优化。
	directIPs []string
	// enableDirectIP，用于判断是否开启 IP 直连，在某些网络环境下，IP 无法直连的，但是如果运行期去尝试连接，会浪费 CPU，浪费服务端连接资源。
	// 所以在 CollectorAddr 对象创建时，对所有 IP 进行一次判断，如果所有 IP 都连不成功，则说明网络无法直连 IP。
	enableDirectIP bool
	mu             sync.RWMutex
	idx            atomic.Int32
}

// NewCollectorAddr 创建 CollectorAddr 对象。
func NewCollectorAddr(fullURL string, directIPPorts []string) *CollectorAddr {
	c := &CollectorAddr{
		FullURL:        fullURL,
		directIPs:      nil,
		enableDirectIP: false,
	}
	c.UpdateIPPorts(fullURL, directIPPorts)
	go c.initEnableDirectIP()
	return c
}

// initEnableDirectIP 有可能耗时很长，放在异步协程中执行，避免阻塞进程初始化。
func (c *CollectorAddr) initEnableDirectIP() {
	success := checkAnyURLConnectionSuccess(c.directIPs)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enableDirectIP = success
}

// UpdateIPPorts 更新配置
func (c *CollectorAddr) UpdateIPPorts(fullURL string, directIPPorts []string) {
	u, err := url.Parse(fullURL)
	// 在配置正确的情况下，此 err 是不可能发生的。
	if err != nil {
		return
	}
	addrStatuses := buildOrderedAddrStatus(u, directIPPorts)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.FullURL = fullURL
	c.directIPs = addrStatuses
}

func buildOrderedAddrStatus(u *url.URL, directIPPorts []string) []string {
	urls := make([]string, 0, len(directIPPorts))
	for _, v := range directIPPorts {
		urls = append(
			urls, directURL(u, v),
		)
	}
	return urls
}

// checkAnyURLConnectionSuccess 检查给定的 URL 列表中是否至少有一个 URL 能够成功连接。
// 此方法受网络环境影响，有可能会耗时较长。
func checkAnyURLConnectionSuccess(urls []string) bool {
	for _, u := range urls {
		if isConnectionSuccessful(u) {
			return true
		}
	}
	return false
}

// 检测 URL 是否可以访问
var isConnectionSuccessful = func(url string) bool {
	client := &http.Client{
		Timeout: 2 * time.Second, // 设置请求超时
	}

	// 空消息压缩后是 0x0
	data := []byte{0x0}
	resp, err := client.Post(url, "application/octet-stream", bytes.NewReader(data))
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	// 如果收到服务端的消息，说明网络是通的。
	if len(body) != 0 {
		return true
	}
	// 服务端会返回 500 状态码。
	if resp.StatusCode == http.StatusInternalServerError {
		return true
	}
	// 假如服务端逻辑变了，返回 200，成功
	if resp.StatusCode == http.StatusOK {
		return true
	}
	// 其他状态码认为是失败
	return false
}

func directURL(u *url.URL, directIPPort string) string {
	u.Host = directIPPort
	return u.String()
}

// GetAddr 获取 OTP 服务地址，retryIdx: 第几次重试，第一次重试使用直连地址，第二次开始使用域名。
func (c *CollectorAddr) GetAddr(retryIdx int) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.enableDirectIP {
		return c.FullURL
	}
	if retryIdx > 0 {
		return c.FullURL
	}
	n := len(c.directIPs)
	if n == 0 {
		return c.FullURL
	}
	idx := int(c.idx.Inc())
	return c.directIPs[idx%n]
}
