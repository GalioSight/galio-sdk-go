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

// Package runtimes 运行时指标
package runtimes

import (
	"galiosight.ai/galio-sdk-go/model"
)

// Write 写运行时监控到 Metrics。
func Write(metrics *model.Metrics) {
	writeGoMetrics(metrics)
	writeProcessMetrics(metrics)
	writeFDMetrics(metrics)
	writePidCount(metrics)
	writeDiskUsage(metrics, "/")
}

// WriteGalileoMetric 写 galileo runtime metric 数据
func WriteGalileoMetric(metrics *model.Metrics) {
	writeServerQPS(metrics)
}
