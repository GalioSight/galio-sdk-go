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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
)

func Test_retry(t *testing.T) {
	cnt := atomic.Int64{}
	type args struct {
		f       func() error
		timeout time.Duration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wantCnt int
	}{
		{
			name: "succ",
			args: args{
				f: func() error {
					cnt.Inc()
					return nil
				}, timeout: time.Millisecond,
			},
			wantErr: false,
			wantCnt: 1,
		},
		{
			name: "error",
			args: args{
				f: func() error {
					cnt.Inc()
					time.Sleep(time.Millisecond * 50)
					return errors.New("error")
				}, timeout: time.Millisecond * 100,
			},
			wantErr: true,
			wantCnt: 2,
		},
		{
			name: "timeout",
			args: args{
				f: func() error {
					cnt.Inc()
					time.Sleep(time.Millisecond * 200)
					return errors.New("timeout")
				},
				timeout: time.Millisecond * 100,
			},
			wantErr: true,
			wantCnt: 1,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				cnt.Store(0)
				err := retry(tt.args.f, tt.args.timeout, 100)
				if (err != nil) != tt.wantErr {
					t.Errorf("retry() error = %v, wantErr %v", err, tt.wantErr)
				}
				assert.Equal(t, tt.wantCnt, int(cnt.Load()))
				// 确保协程退出，sleep 一段时间，判定它的执行次数没有增加。
				time.Sleep(time.Second)
				assert.Equal(t, tt.wantCnt, int(cnt.Load()))
			},
		)
	}
}

func TestRetryPanic(t *testing.T) {

	// 模拟一个总是返回 nil 的函数
	f := func() error {
		return nil
	}

	// 调用 retryOld 函数，期望发生 panic
	_ = retry(f, 1*time.Second, -1)
}
