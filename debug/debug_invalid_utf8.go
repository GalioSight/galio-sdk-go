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
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package debug

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	selflog "galiosight.ai/galio-sdk-go/self/log"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	logsproto "go.opentelemetry.io/proto/otlp/logs/v1"
	"google.golang.org/grpc/status"
)

// UTF8Debugger SDK 调试器。
type UTF8Debugger interface {
	Enabled() bool
	DebugSpansInvalidUTF8(exportErr error, batch []sdktrace.ReadOnlySpan)
	DebugLogsInvalidUTF8(exportErr error, batch []*logsproto.ResourceLogs)
}

// NewUTF8Debugger 创建一个调试器。
func NewUTF8Debugger() UTF8Debugger {
	d := &debugger{}
	// export GALILEO_SDK_DEBUG=utf8
	if sdkDebugEnv := os.Getenv("GALILEO_SDK_DEBUG"); strings.Contains(sdkDebugEnv, "utf8") {
		selflog.Infof("galileo: env GALILEO_SDK_DEBUG:%s", sdkDebugEnv)
		d.enabled = true
	}
	return d
}

type debugger struct {
	enabled bool
}

// Enabled 启用开关
func (d *debugger) Enabled() bool {
	return d.enabled
}

// DebugSpansInvalidUTF8 调试导出 spans 时 invalid utf8 错误。
func (d *debugger) DebugSpansInvalidUTF8(exportErr error, batch []sdktrace.ReadOnlySpan) {
	s, ok := status.FromError(exportErr)
	if !ok {
		return
	}
	if !strings.Contains(s.String(), "invalid UTF-8") {
		return
	}
	for _, v := range batch {
		d.debugUTF8(telemetrySpan, "Name", v.Name())
		d.debugUTF8(telemetrySpan, "Status.Description", v.Status().Description)
		for _, attr := range v.Resource().Attributes() {
			d.debugUTF8(telemetrySpan, fmt.Sprintf("Resource.Attributes.Key.%s", attr.Key), string(attr.Key))
			d.debugUTF8(telemetrySpan, fmt.Sprintf("Resource.Attributes.%s", attr.Key), attr.Value.Emit())
		}
		for _, attr := range v.Attributes() {
			d.debugUTF8(telemetrySpan, fmt.Sprintf("Attributes.Key.%s", attr.Key), string(attr.Key))
			d.debugUTF8(telemetrySpan, fmt.Sprintf("Attributes.%s", attr.Key), attr.Value.Emit())
		}
		for i, event := range v.Events() {
			d.debugUTF8(telemetrySpan, fmt.Sprintf("Events.%d.Name", i), event.Name)
			for _, attr := range event.Attributes {
				d.debugUTF8(
					telemetrySpan,
					fmt.Sprintf("Events.%d.Attributes.Key.%s", i, attr.Key), string(attr.Key),
				)
				d.debugUTF8(
					telemetrySpan, fmt.Sprintf(
						"Events.%d.Attributes.%s",
						i, attr.Key,
					), attr.Value.Emit(),
				)
			}
		}
	}
}

// DebugLogsInvalidUTF8 调试导出 logs 时 invalid utf8 错误。
func (d *debugger) DebugLogsInvalidUTF8(exportErr error, batch []*logsproto.ResourceLogs) {
	s, ok := status.FromError(exportErr)
	if !ok {
		return
	}
	if !strings.Contains(s.String(), "invalid UTF-8") {
		return
	}
	for _, v := range batch {
		for _, attr := range v.GetResource().GetAttributes() {
			d.debugUTF8(
				telemetryLog,
				fmt.Sprintf("Attributes.Attributes.Key.%s", attr.Key),
				attr.Key,
			)
			d.debugUTF8(
				telemetryLog,
				fmt.Sprintf("Resource.Attributes.%s", attr.Key),
				attr.Value.GetStringValue(),
			)
		}
		for _, vv := range v.GetScopeLogs() {
			for _, vvv := range vv.GetLogRecords() {
				for _, attr := range vvv.GetAttributes() {
					d.debugUTF8(
						telemetryLog,
						fmt.Sprintf("Attributes.Key.%s", attr.Key),
						attr.Key,
					)
					d.debugUTF8(
						telemetryLog,
						fmt.Sprintf("Attributes.%s", attr.Key),
						attr.Value.GetStringValue(),
					)
				}
				d.debugUTF8(telemetryLog, "Message", vvv.GetBody().GetStringValue())
			}
		}
	}
}

func (d *debugger) debugUTF8(telemetry string, field string, value string) {
	if !utf8.ValidString(value) {
		selflog.Infof(
			"galileo: %v.%v is not a valid UTF-8 string, value:%s",
			telemetry, field, value,
		)
	}
}

const (
	telemetrySpan = "span"
	telemetryLog  = "log"
)
