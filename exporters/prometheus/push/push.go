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

// Package push 用于将 prometheus 指标通过 push 方式上报到 pushgateway
package push

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"

	"galiosight.ai/galio-sdk-go/model"
	selflog "galiosight.ai/galio-sdk-go/self/log"
	selfmetric "galiosight.ai/galio-sdk-go/self/metric"
	omp3 "galiosight.ai/galio-sdk-go/semconv"
	semconv "galiosight.ai/galio-sdk-go/semconv/v1.0.0"
)

// PrometheusPush 将 Prometheus 指标通过 push 方式进行上报。
// 此函数主要用于兼容天机阁等使用 Prometheus API 的场景。
// 如果推送成功，将返回一个取消函数和 nil；如果失败，将返回相应的错误。
func PrometheusPush(res model.Resource, cfg model.PrometheusPushConfig, opt ...option) (context.CancelFunc, error) {
	if !cfg.Enable {
		return nil, ErrCfgEnabled
	}

	var o options
	o.schemaURL = semconv.SchemaURL
	for _, f := range opt {
		f(&o)
	}

	pusher := createPusher(cfg)
	if o.schemaURL == omp3.SchemaURL {
		addGroupingV3(pusher, &res)
	} else {
		addGroupingV1(pusher, &res)
	}
	for name, value := range cfg.Grouping {
		pusher.Grouping(name, value)
	}

	cfg.HttpHeaders = addResourceHeaders(cfg.HttpHeaders, res, &o)
	pusher.Client(newHTTPDoerWithHeaders(cfg.HttpHeaders))
	if err := doPush(pusher); err != nil {
		selfmetric.GetSelfMonitor().Stats.PrometheusPushStats.InitErrorTotal.Inc()
		return nil, errors.New("failed to push prometheus metrics: " + err.Error())
	}

	return startTicker(pusher, cfg.Interval)
}

// createPusher 创建一个新的 pusher 实例。
// 该实例用于将指标推送到指定的 Prometheus 服务器。
func createPusher(cfg model.PrometheusPushConfig) *push.Pusher {
	pusher := push.New(cfg.Url, cfg.Job).
		Gatherer(prometheus.DefaultGatherer)

	if cfg.UseBasicAuth {
		pusher.BasicAuth(cfg.Username, cfg.Password)
	}

	return pusher
}

// addResourceHeaders 将资源的相关信息添加到 HTTP 头部，以便在推送时使用。
func addResourceHeaders(headers map[string]string, res model.Resource, o *options) map[string]string {
	if headers == nil {
		headers = make(map[string]string)
	}
	headers[model.TargetHeaderKey] = res.Target
	headers[model.TenantHeaderKey] = res.TenantId
	headers[model.SchemaURLHeaderKey] = o.schemaURL
	return headers
}

// startTicker 启动定时器以定期推送指标。
// 如果配置的间隔有效，返回一个取消函数；否则返回相应的错误。
func startTicker(pusher *push.Pusher, interval int32) (context.CancelFunc, error) {
	if interval <= 0 {
		return nil, ErrCfgInterval
	}

	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(time.Duration(interval) * time.Second)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = doPush(pusher)
			case <-ctx.Done():
				return
			}
		}
	}()

	return cancel, nil
}

func doPush(pusher *push.Pusher) error {
	if err := pusher.Push(); err != nil {
		selfmetric.GetSelfMonitor().Stats.PrometheusPushStats.FailedExportCounter.Inc()
		selflog.Errorf("failed to push prometheus metrics: %v", err)
		return err
	} else {
		selfmetric.GetSelfMonitor().Stats.PrometheusPushStats.SucceededExportCounter.Inc()
	}
	return nil
}

// ErrCfgInterval 表示配置的间隔时间无效。
// 当 cfg.Interval <= 0 时返回该错误。
var ErrCfgInterval = errors.New("cfg.Interval <= 0")

// ErrCfgEnabled 表示 Prometheus 推送未启用。
// 当 cfg.Enabled = false 时返回该错误。
var ErrCfgEnabled = errors.New("cfg.Enabled = false, no prometheus push enabled")

type pushHTTPDoer struct {
	headers map[string]string
	client  *http.Client
}

// Do 执行 HTTP 请求，并在请求中添加指定的头部信息。
func (p *pushHTTPDoer) Do(r *http.Request) (*http.Response, error) {
	for k, v := range p.headers {
		r.Header.Set(k, v)
	}
	return p.client.Do(r)
}

// newHTTPDoerWithHeaders 创建一个新的 HTTP Doer 实例，
// 该实例在执行请求时会添加指定的头部信息。
func newHTTPDoerWithHeaders(headers map[string]string) push.HTTPDoer {
	return &pushHTTPDoer{
		headers: headers,
		client:  &http.Client{},
	}
}
