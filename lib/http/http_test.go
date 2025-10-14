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
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"galiosight.ai/galio-sdk-go/model"
	"github.com/golang/snappy"
	"github.com/stretchr/testify/assert"
)

func TestPost(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"code":0,"msg":"success"}`))
			},
		),
	)
	defer ts.Close()
	b := serializeTss()
	buf, err := Post(ts.Client(), ts.URL, b)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("sendByTrpc err: %v", err)
	}
	assert.Equal(t, string(buf), `{"code":0,"msg":"success"}`)
	_, err = Post(ts.Client(), "error", b)
	assert.NotNil(t, err)
}

func TestPostURLError(t *testing.T) {
	b := serializeTss()
	buf, err := Post(&http.Client{}, "a", b)
	assert.Error(t, err)
	assert.Equal(t, []byte(nil), buf)
}

func TestPostURLEmptyError(t *testing.T) {
	b := serializeTss()
	buf, err := Post(&http.Client{}, "", b)
	assert.Error(t, err)
	assert.Equal(t, []byte(nil), buf)
}

func TestPostNetError(t *testing.T) {
	b := serializeTss()
	buf, err := Post(&http.Client{}, "http://127.0.0.1:12701/test", b)
	assert.Error(t, err)
	assert.Equal(t, []byte(nil), buf)
	assert.True(t, IsNetOpError(err))
}

func TestPostTimeoutError(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(time.Second * 10)
				_, _ = w.Write([]byte(`{"code":0,"msg":"success"}`))
			},
		),
	)
	defer ts.Close()
	b := serializeTss()
	buf, err := Post(&http.Client{Timeout: time.Millisecond * 10}, ts.URL, b)
	assert.Error(t, err)
	assert.Equal(t, []byte(nil), buf)
	assert.False(t, IsNetOpError(err))
	err = errors.Unwrap(err)
	err = errors.Unwrap(err)
	assert.Equal(t, "context deadline exceeded (Client.Timeout exceeded while awaiting headers)", err.Error())
}

func TestPostHttp500Error(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("test 500"))
			},
		),
	)
	defer ts.Close()
	b := serializeTss()
	buf, err := Post(&http.Client{Timeout: time.Millisecond * 10}, ts.URL, b)
	assert.Error(t, err)
	assert.Equal(t, "test 500", string(buf))
	assert.Equal(t, "httpStatusCodeErr: 500, rsp=test 500", err.Error())
}

func serializeTss() []byte {
	req := &model.Metrics{
		ClientMetrics: []*model.ClientMetricsOTP{
			{
				RpcClientStartedTotal: 102,
				RpcClientHandledTotal: 203,
				RpcClientHandledSeconds: &model.Histogram{
					Sum:   101.2,
					Count: 3,
					Buckets: []*model.Bucket{
						{
							Range: "23",
							Count: 19,
						},
						{
							Range: "367",
							Count: 10,
						},
					},
				},
				RpcLabels: &model.RPCLabels{
					Fields: []model.RPCLabels_Field{
						{
							Name:  model.RPCLabels_callee_container,
							Value: "container10123",
						},
						{
							Name:  model.RPCLabels_callee_ip,
							Value: "127.0.0.1:123",
						},
					},
				},
			},
		},
	}
	b, err := req.Marshal()
	if err != nil {
		return nil
	}
	dstBuf := snappy.Encode(nil, b)
	return dstBuf
}

func TestName(t *testing.T) {
	t.Logf("%v", time.Unix(0, 5707039364956332032))
}
