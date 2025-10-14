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

// Package errs 定义一些常见错误
package errs

import (
	"errors"
	"fmt"
)

var (
	// ErrFactoriesEmpty 工厂总表为空。
	ErrFactoriesEmpty = errors.New("factories empty")
	// ErrFactoryEmpty 工厂为空。
	ErrFactoryEmpty = errors.New("factory empty")
	// ErrCreateMetricsProcessor 创建监控处理器错误。
	ErrCreateMetricsProcessor = errors.New("create metrics processor error")
	// ErrCreateTracesProcessor 创建追踪处理器错误。
	ErrCreateTracesProcessor = errors.New("create traces processor error")
	// ErrCreateLogsProcessor 创建日志处理器错误。
	ErrCreateLogsProcessor = errors.New("create logs processor error")
	// ErrCreateProfilesProcessor 创建监控导出器错误。
	ErrCreateProfilesProcessor = errors.New("create profiles processor error")
	// ErrCreateMetricsExporter 创建监控导出器错误。
	ErrCreateMetricsExporter = errors.New("create metrics exporter error")
	// ErrCreateTracesExporter 创建追踪导出器错误。
	ErrCreateTracesExporter = errors.New("create traces exporter error")
	// ErrCreateLogsExporter 创建日志导出器错误。
	ErrCreateLogsExporter = errors.New("create logs exporter error")
	// ErrCreateProfilesExporter 创建日志导出器错误。
	ErrCreateProfilesExporter = errors.New("create profiles exporter error")
	// ErrGroupInvalid group 命名错误。
	ErrGroupInvalid = errors.New("group invalid, must match regex [a-zA-Z0-9_]*")
	// ErrNameInvalid name 命名错误。
	ErrNameInvalid = errors.New("name invalid, must match regex [a-zA-Z0-9_]*")
	// ErrSDKNotInit SDK 未初始化错误。
	ErrSDKNotInit = errors.New("base SDK not initialized")
	// ErrCustomName 不是合法的 CustomName
	ErrCustomName = errors.New("not CustomName")
	// ErrOcpInvalid 错误的 ocp 地址
	ErrOcpInvalid = errors.New("ocp addr invalid")
	// ErrResourceAlreadyRegistered 表示资源已经注册过的错误
	ErrResourceAlreadyRegistered = errors.New("resource already registered")
	// ErrTargetNotExist 表示 target 不存在的错误
	ErrTargetNotExist = errors.New("target not exist")
	// ErrTimeout 超时
	ErrTimeout = errors.New("timeout")
)

// otlp logs exporter 错误码汇总。
var (
	ErrOTLPLogsExporterAlreadyStarted  = errors.New("otlp logs exporter already started")
	ErrOTLPLogsExporterNotStarted      = errors.New("otlp logs exporter not started")
	ErrOTLPLogsExporterDisconnected    = errors.New("otlp logs exporter disconnected")
	ErrOTLPLogsExporterStopped         = errors.New("otlp logs exporter stopped")
	ErrOTLPLogsExporterContextCanceled = errors.New("otlp logs exporter context canceled")
)

var (
	// ErrReadProcSelfStatEmpty 进程监控错误
	ErrReadProcSelfStatEmpty = errors.New("read /proc/self/stat empty")
)

// NoTenantError 没有租户 id 的 ERROR. 此错误是非常严重的，会导致插件加载失败。
func NoTenantError(target string) error {
	return fmt.Errorf("取不到租户，可能是服务 %s 未允许接入伽利略，请联系 @伽利略小助手", target)
}
