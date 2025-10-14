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

//go:build !windows
// +build !windows

package runtimes

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"galiosight.ai/galio-sdk-go/model"
)

// writePidCount 写 pid 监控到 Metrics。
func writePidCount(metrics *model.Metrics) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	shell := "ps -eLF|wc -l"
	out, err := exec.CommandContext(ctx, "bash", "-c", shell).Output()
	if err != nil {
		return
	}
	pidNum, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return
	}
	metrics.AddNormalMetric("process_pid_num", model.Aggregation_AGGREGATION_SET, pidNum)
}
