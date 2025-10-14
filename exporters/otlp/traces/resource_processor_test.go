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

package traces

import (
	"context"
	"reflect"
	"testing"

	semconv "galiosight.ai/galio-sdk-go/semconv/v1.0.0"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// TestSimpleTrpcProcessor 测试 ResourceProcessor
func TestSimpleTrpcProcessor(t *testing.T) {
	// 创建一个 MemoryExporter
	exporter := tracetest.NewInMemoryExporter()
	const strNamespace = "theNamespace"
	const strEnv = "theEnv"
	// 创建 ResourceProcessor
	processor := sdktrace.NewSimpleSpanProcessor(exporter)
	var resSpanProcessorOption []ResourceSpanProcessorOption
	resSpanProcessorOption = append(resSpanProcessorOption, WithNameSpace(strNamespace))
	resSpanProcessorOption = append(resSpanProcessorOption, WithEnvName(strEnv))
	// 创建一个 TracerProvider 并设置 ResourceProcessor
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(
			NewResourceProcessor(processor, resSpanProcessorOption...),
		),
	)
	otel.SetTracerProvider(tp)

	// 创建一个 Tracer
	tracer := tp.Tracer("test")

	// 创建一个测试 Span
	ctx, span := tracer.Start(context.Background(), "test-span")

	// 模拟 OnStart 调用
	processor.OnStart(ctx, span.(sdktrace.ReadWriteSpan))

	// 必须要结束 span
	span.End()
	// 检查导出器中的 Span 是否设置了正确的属性
	if len(exporter.GetSpans()) != 1 {
		t.Fatalf("expected 1 span, got %d", len(exporter.GetSpans()))
	}
	gotAttrs := exporter.GetSpans()[0].Attributes
	wantAttrs := []attribute.KeyValue{
		semconv.TrpcNamespaceKey.String(strNamespace),
		semconv.TrpcEnvnameKey.String(strEnv),
	}

	if !reflect.DeepEqual(gotAttrs, wantAttrs) {
		t.Errorf("got attributes %v, want %v", gotAttrs, wantAttrs)
	}
}
