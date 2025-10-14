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

//go:build linux
// +build linux

package runtimes

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"galiosight.ai/galio-sdk-go/errs"
	"galiosight.ai/galio-sdk-go/model"
)

// writeProcessMetrics 上报逻辑迁移自：https://github.com/VictoriaMetrics/metrics/blob/master/process_metrics_linux.go
func writeProcessMetrics(metrics *model.Metrics) {
	writeProcStat(metrics)
	writeMemStat(metrics)
	writeIOStat(metrics)
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

func writeProcStat(metrics *model.Metrics) {
	stat, err := ReadProcStat("")
	if err != nil {
		return
	}
	utime := float64(stat.Utime) / UserHZ
	stime := float64(stat.Stime) / UserHZ
	metrics.AddNormalMetric("process_cpu_seconds_system_total", model.Aggregation_AGGREGATION_SET, float64(stime))
	metrics.AddNormalMetric("process_cpu_seconds_total", model.Aggregation_AGGREGATION_SET, float64(utime+stime))
	metrics.AddNormalMetric("process_cpu_seconds_user_total", model.Aggregation_AGGREGATION_SET, float64(utime))
	metrics.AddNormalMetric("process_major_pagefaults_total", model.Aggregation_AGGREGATION_SET, float64(stat.Majflt))
	metrics.AddNormalMetric("process_minor_pagefaults_total", model.Aggregation_AGGREGATION_SET, float64(stat.Minflt))
	metrics.AddNormalMetric("process_num_threads", model.Aggregation_AGGREGATION_SET, float64(stat.NumThreads))
	metrics.AddNormalMetric("process_resident_memory_bytes", model.Aggregation_AGGREGATION_SET, float64(stat.Rss*4096))
	metrics.AddNormalMetric("process_start_time_seconds", model.Aggregation_AGGREGATION_SET, float64(startTimeSeconds))
	metrics.AddNormalMetric("process_virtual_memory_bytes", model.Aggregation_AGGREGATION_SET, float64(stat.Vsize))
}

// ReadProcStat 读取进程统计。
func ReadProcStat(fileName string) (*ProcStat, error) {
	if fileName == "" {
		fileName = "/proc/self/stat"
	}
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	return parseProcStat(data)
}

// parseProcStat data：10462 (cat) R 23041 10462 23041 ... ... ...
func parseProcStat(data []byte) (*ProcStat, error) {
	n := bytes.LastIndex(data, []byte(") "))
	if n < 0 {
		return nil, errs.ErrReadProcSelfStatEmpty
	}
	if n+2 >= len(data) {
		return nil, errs.ErrReadProcSelfStatEmpty
	}
	data = data[n+2:]
	var stat ProcStat
	buf := bytes.NewBuffer(data)
	_, err := fmt.Fscanf(
		buf, "%c %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d",
		&stat.State, &stat.Ppid, &stat.Pgrp, &stat.Session, &stat.TtyNr, &stat.Tpgid, &stat.Flags, &stat.Minflt,
		&stat.Cminflt, &stat.Majflt, &stat.Cmajflt, &stat.Utime, &stat.Stime, &stat.Cutime, &stat.Cstime,
		&stat.Priority, &stat.Nice, &stat.NumThreads, &stat.ItrealValue, &stat.Starttime, &stat.Vsize, &stat.Rss,
	)
	if err != nil {
		return nil, err
	}
	return &stat, nil
}

// MemStat see https://man7.org/linux/man-pages/man5/procfs.5.html
type MemStat struct {
	VMPeak   uint64
	RssPeak  uint64
	RssAnon  uint64
	RssFile  uint64
	RssShmem uint64
}

func writeMemStat(metrics *model.Metrics) {
	stat, err := ReadMemStat("")
	if err != nil {
		return
	}
	metrics.AddNormalMetric(
		"process_virtual_memory_peak_bytes",
		model.Aggregation_AGGREGATION_SET, float64(stat.VMPeak),
	)
	metrics.AddNormalMetric(
		"process_resident_memory_peak_bytes",
		model.Aggregation_AGGREGATION_SET, float64(stat.RssPeak),
	)
	metrics.AddNormalMetric(
		"process_resident_memory_anon_bytes",
		model.Aggregation_AGGREGATION_SET, float64(stat.RssAnon),
	)
	metrics.AddNormalMetric(
		"process_resident_memory_file_bytes",
		model.Aggregation_AGGREGATION_SET, float64(stat.RssFile),
	)
	metrics.AddNormalMetric(
		"process_resident_memory_shared_bytes",
		model.Aggregation_AGGREGATION_SET, float64(stat.RssShmem),
	)
}

// ReadMemStat 读取内存统计。
func ReadMemStat(fileName string) (*MemStat, error) {
	if fileName == "" {
		fileName = "/proc/self/status"
	}
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	return parseMemStat(data)
}

func parseMemStat(data []byte) (*MemStat, error) {
	var stat MemStat
	lines := strings.Split(string(data), "\n")
	for _, s := range lines {
		if !strings.HasPrefix(s, "Vm") && !strings.HasPrefix(s, "Rss") {
			continue
		}
		// Extract key value.
		line := strings.Fields(s)
		if len(line) != 3 {
			return nil, fmt.Errorf("unexpected number of fields found in %q; got %d; want %d", s, len(line), 3)
		}
		memStatName := line[0]
		memStatValue := line[1]
		value, err := strconv.ParseUint(memStatValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot parse number from %q: %w", s, err)
		}
		if line[2] != "kB" {
			return nil, fmt.Errorf("expecting kB value in %q; got %q", s, line[2])
		}
		value *= 1024
		switch memStatName {
		case "VmPeak:":
			stat.VMPeak = value
		case "VmHWM:":
			stat.RssPeak = value
		case "RssAnon:":
			stat.RssAnon = value
		case "RssFile:":
			stat.RssFile = value
		case "RssShmem:":
			stat.RssShmem = value
		}
	}
	return &stat, nil
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

func writeIOStat(metrics *model.Metrics) {
	stat, err := ReadIOStat("")
	if err != nil {
		return
	}
	metrics.AddNormalMetric(
		"process_io_read_bytes_total",
		model.Aggregation_AGGREGATION_SET, float64(stat.Rchar),
	)
	metrics.AddNormalMetric(
		"process_io_written_bytes_total",
		model.Aggregation_AGGREGATION_SET, float64(stat.Wchar),
	)
	metrics.AddNormalMetric(
		"process_io_read_syscalls_total",
		model.Aggregation_AGGREGATION_SET, float64(stat.Syscr),
	)
	metrics.AddNormalMetric(
		"process_io_write_syscalls_total",
		model.Aggregation_AGGREGATION_SET, float64(stat.Syscw),
	)
	metrics.AddNormalMetric(
		"process_io_storage_read_bytes_total",
		model.Aggregation_AGGREGATION_SET, float64(stat.ReadBytes),
	)
	metrics.AddNormalMetric(
		"process_io_storage_written_bytes_total",
		model.Aggregation_AGGREGATION_SET, float64(stat.WriteBytes),
	)
}

// ReadIOStat 读 IO 统计。
func ReadIOStat(fileName string) (*IOStat, error) {
	if fileName == "" {
		fileName = "/proc/self/io"
	}
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	return parseIOStat(data)
}

func parseIOStat(data []byte) (*IOStat, error) {
	getInt := func(s string) int64 {
		n := strings.IndexByte(s, ' ')
		if n < 0 {
			return 0
		}
		v, err := strconv.ParseInt(s[n+1:], 10, 64)
		if err != nil {
			return 0
		}
		return v
	}
	var stat IOStat
	lines := strings.Split(string(data), "\n")
	for _, s := range lines {
		s = strings.TrimSpace(s)
		switch {
		case strings.HasPrefix(s, "rchar: "):
			stat.Rchar = getInt(s)
		case strings.HasPrefix(s, "wchar: "):
			stat.Wchar = getInt(s)
		case strings.HasPrefix(s, "syscr: "):
			stat.Syscr = getInt(s)
		case strings.HasPrefix(s, "syscw: "):
			stat.Syscw = getInt(s)
		case strings.HasPrefix(s, "read_bytes: "):
			stat.ReadBytes = getInt(s)
		case strings.HasPrefix(s, "write_bytes: "):
			stat.WriteBytes = getInt(s)
		}
	}
	return &stat, nil
}

// K8SStat k8s 统计。
type K8SStat struct {
	MemoryUsageInBytes int64 // /sys/fs/cgroup/memory/memory.usage_in_bytes
	MemoryLimitInBytes int64 // /sys/fs/cgroup/memory/memory.limit_in_bytes
}

// ReadK8SStat 读 k8s 统计。
func ReadK8SStat(memoryUsageFileName, memoryLimitFileName string) (*K8SStat, error) {
	var stat K8SStat
	// 内存使用。
	if memoryUsageFileName == "" {
		memoryUsageFileName = "/sys/fs/cgroup/memory/memory.usage_in_bytes"
	}
	number, err := readNumber(memoryUsageFileName)
	if err != nil {
		return nil, err
	}
	stat.MemoryUsageInBytes = number
	// 内存限制。
	if memoryLimitFileName == "" {
		memoryLimitFileName = "/sys/fs/cgroup/memory/memory.limit_in_bytes"
	}
	number, err = readNumber(memoryLimitFileName)
	if err != nil {
		return nil, err
	}
	stat.MemoryLimitInBytes = number
	return &stat, nil
}

func readNumber(fileName string) (int64, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return 0, err
	}
	number, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, err
	}
	return number, nil
}
