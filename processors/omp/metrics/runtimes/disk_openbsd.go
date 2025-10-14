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

//go:build openbsd
// +build openbsd

package runtimes

import (
	"syscall"

	"galiosight.ai/galio-sdk-go/model"
)

// writeDiskUsage 写 disk 监控到 Metrics。
// openbsd 的 syscall.Statfs_t 结构与其他 unix 不一致，这里单独处理。
func writeDiskUsage(metrics *model.Metrics, path string) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return
	}
	diskAll := float64(fs.F_blocks * uint64(fs.F_bsize))
	diskFree := float64(fs.F_bfree * uint64(fs.F_bsize))
	diskUsed := diskAll - diskFree
	diskUsedFraction := diskUsed / diskAll

	metrics.AddNormalMetric("process_disk_free", model.Aggregation_AGGREGATION_SET, diskFree/float64(GB))
	metrics.AddNormalMetric("process_disk_used", model.Aggregation_AGGREGATION_SET, diskUsed/float64(GB))
	metrics.AddNormalMetric("process_disk_used_fraction", model.Aggregation_AGGREGATION_SET, diskUsedFraction)
}
