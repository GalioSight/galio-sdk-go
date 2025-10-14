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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	logpb "go.opentelemetry.io/proto/otlp/logs/v1"
)

// MockSpanExporter is a mock implementation of sdktrace.SpanExporter
type MockSpanExporter struct {
	shouldError bool
}

func (m MockSpanExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	if m.shouldError {
		return errors.New("export error")
	}
	return nil
}

func (m MockSpanExporter) Shutdown(ctx context.Context) error {
	return nil
}

func TestExportBatch(t *testing.T) {
	// Create a mock SpanExporter
	exporter := &MockSpanExporter{}

	bsp := &batchSpanProcessor{
		e:     exporter,
		batch: []sdktrace.ReadOnlySpan{},
	}
	bsp.o.exportToFile = true

	err := bsp.exportBatch(context.Background())

	assert.NoError(t, err)
}

func TestExportBatchError(t *testing.T) {
	// Create a mock SpanExporter that returns an error
	exporter := &MockSpanExporter{shouldError: true}

	bsp := &batchSpanProcessor{
		e:     exporter,
		batch: []sdktrace.ReadOnlySpan{},
	}
	bsp.o.exportToFile = false

	err := bsp.exportBatch(context.Background())

	assert.Error(t, err)
}

func TestExportBatchNil(t *testing.T) {
	bsp := &batchSpanProcessor{
		e:     nil,
		batch: []sdktrace.ReadOnlySpan{},
	}
	bsp.o.exportToFile = true
	err := bsp.exportBatch(context.Background())

	assert.NoError(t, err)
}

// MockDebugger 是一个模拟的 Debugger
type MockDebugger struct {
	enabled bool
	spans   int
}

func (m *MockDebugger) DebugLogsInvalidUTF8(exportErr error, batch []*logpb.ResourceLogs) {
}

func (m *MockDebugger) Enabled() bool {
	m.enabled = true
	return m.enabled
}

func (m *MockDebugger) DebugSpansInvalidUTF8(err error, spans []sdktrace.ReadOnlySpan) {
	m.spans++
}

func TestHandleError(t *testing.T) {
	debugger := &MockDebugger{}
	bsp := &batchSpanProcessor{
		debugger: debugger,
	}

	err := errors.New("string field contains invalid UTF-8")

	bsp.handleError(err)
	assert.True(t, debugger.enabled)
	assert.GreaterOrEqual(t, 1, debugger.spans)
}
