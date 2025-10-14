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
	"context"
	"os"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	v11 "go.opentelemetry.io/proto/otlp/common/v1"
	logsproto "go.opentelemetry.io/proto/otlp/logs/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func Test_debugger_DebugLogsInvalidUTF8(t *testing.T) {
	_ = os.Setenv("GALILEO_SDK_DEBUG", "utf8")
	invalidUTF8String := string([]byte{255, 255})
	_, invalidUTF8Error := proto.Marshal(&logsproto.LogRecord{SeverityText: invalidUTF8String})
	exportError := status.Error(codes.Internal, invalidUTF8Error.Error())
	d := NewUTF8Debugger()
	tp := sdktrace.NewTracerProvider(sdktrace.WithResource(resource.Empty()))
	_, span := tp.Tracer("").Start(context.Background(), "invalid UTF-8")
	span.SetAttributes(attribute.String("xxKey", invalidUTF8String))
	span.AddEvent(invalidUTF8String, trace.WithAttributes(
		attribute.String("eventKey", invalidUTF8String)))
	span.End()
	d.DebugSpansInvalidUTF8(exportError, []sdktrace.ReadOnlySpan{span.(sdktrace.ReadOnlySpan)})
	d.DebugLogsInvalidUTF8(exportError, []*logsproto.ResourceLogs{
		{
			ScopeLogs: []*logsproto.ScopeLogs{
				{
					LogRecords: []*logsproto.LogRecord{
						{
							Attributes: []*v11.KeyValue{
								{
									Key: "invalidKey",
									Value: &v11.AnyValue{
										Value: &v11.AnyValue_StringValue{
											StringValue: invalidUTF8String,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})
}
