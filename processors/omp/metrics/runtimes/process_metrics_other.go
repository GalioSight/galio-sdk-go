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

//go:build !linux
// +build !linux

package runtimes

import (
	"time"

	"galiosight.ai/galio-sdk-go/model"
)

func writeProcessMetrics(metrics *model.Metrics) {
	// TODO 待实现
}

// ProcStat see http://man7.org/linux/man-pages/man5/proc.5.html
type ProcStat struct {
	State       byte
	Ppid        int
	Pgrp        int
	Session     int
	TtyNr       int
	Tpgid       int
	Flags       uint
	Minflt      uint
	Cminflt     uint
	Majflt      uint
	Cmajflt     uint
	Utime       uint
	Stime       uint
	Cutime      int
	Cstime      int
	Priority    int
	Nice        int
	NumThreads  int
	ItrealValue int
	Starttime   uint64
	Vsize       uint
	Rss         int
}

var startTimeSeconds = time.Now().Unix()

// UserHZ see https://github.com/prometheus/procfs/blob/a4ac0826abceb44c40fc71daed2b301db498b93e/proc_stat.go#L40 .
const UserHZ = 100

// ReadProcStat 读取进程统计。
func ReadProcStat(fileName string) (*ProcStat, error) {
	// TODO 待实现
	return nil, nil
}

// MemStat see https://man7.org/linux/man-pages/man5/procfs.5.html
type MemStat struct {
	VMPeak   uint64
	RssPeak  uint64
	RssAnon  uint64
	RssFile  uint64
	RssShmem uint64
}

// ReadMemStat 读取内存统计。
func ReadMemStat(fileName string) (*MemStat, error) {
	// TODO 待实现
	return nil, nil
}

// IOStat 磁盘统计。
type IOStat struct {
	Rchar      int64
	Wchar      int64
	Syscr      int64
	Syscw      int64
	ReadBytes  int64
	WriteBytes int64
}

// ReadIOStat 读 IO 统计。
func ReadIOStat(fileName string) (*IOStat, error) {
	// TODO 待实现
	return nil, nil
}

// K8SStat k8s 统计。
type K8SStat struct {
	MemoryUsageInBytes int64 // /sys/fs/cgroup/memory/memory.usage_in_bytes
	MemoryLimitInBytes int64 // /sys/fs/cgroup/memory/memory.limit_in_bytes
}

// ReadK8SStat 读 k8s 统计。
func ReadK8SStat(memoryUsageFileName, memoryLimitFileName string) (*K8SStat, error) {
	// TODO 待实现
	return nil, nil
}
