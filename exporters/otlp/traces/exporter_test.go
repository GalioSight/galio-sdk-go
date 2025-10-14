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
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"galiosight.ai/galio-sdk-go/configs/ocp"
	traceconf "galiosight.ai/galio-sdk-go/configs/traces"
	"galiosight.ai/galio-sdk-go/exporters/otlp/traces/tracestate"
	"galiosight.ai/galio-sdk-go/model"
)

func testResource() *model.Resource {
	platform := "Galileo-Dial" // 资源所在平台，如 PCG-123、STKE
	app := "galileo"
	server := "SDK"
	resource := model.NewResource(
		platform, app, server,
		"DemoService",
		model.Production,           // 物理环境，只能在 model.Production 和 model.Development 枚举，正式环境必须是 model.Production
		"formal",                   // 用户环境，一般是 formal 和 test (或形如 3c170118 等自定义), 正式环境必须是 formal
		"set.sz.1",                 // set 名，可以为空
		"sz",                       // 城市，可以为空
		"127.0.0.1",                // 实例 IP，可以为空
		"test.galileo.SDK.sz10010", // 容器名，可以为空
	)
	return resource
}

func TestNewExporter(t *testing.T) {
	c := traceconf.NewConfig(testResource())
	tracesExporter, err := NewExporter(c)
	assert.Nil(t, err)
	assert.NotNil(t, tracesExporter)
	assert.NotNil(t, tracesExporter.(Tracer).DeferredSampler())
}

func TestNewExporter_WithTrpcNamespace(t *testing.T) {
	platform := "Galileo-Dial" // 资源所在平台，如 PCG-123、STKE
	app := "galileo"
	server := "SDK"
	resource := model.NewResource(
		platform, app, server,
		"DemoService",
		model.Production,           // 物理环境，只能在 model.Production 和 model.Development 枚举，正式环境必须是 model.Production
		"formal",                   // 用户环境，一般是 formal 和 test (或形如 3c170118 等自定义), 正式环境必须是 formal
		"set.sz.1",                 // set 名，可以为空
		"sz",                       // 城市，可以为空
		"127.0.0.1",                // 实例 IP，可以为空
		"test.galileo.SDK.sz10010", // 容器名，可以为空，
	)
	c := traceconf.NewConfig(resource)
	tracesExporter, err := NewExporter(c)
	assert.Nil(t, err)
	assert.NotNil(t, tracesExporter)
	tracesExporter.Watch(
		&ocp.GalileoConfig{
			Config: model.GetConfigResponse{},
		},
	)
	ctx, span := tracesExporter.Start(context.Background(), "test")
	span.End()
	assert.Equal(t, span.SpanContext().SpanID().String(), trace.SpanFromContext(ctx).SpanContext().SpanID().String())
}

type testUserSampler struct {
	old UserSampler
}

func (u testUserSampler) ShouldSample(p *sdktrace.SamplingParameters, state *tracestate.SampleState) sdktrace.SamplingResult {
	for _, a := range p.Attributes {
		if a.Key == "user-sample" {
			v := a.Value.AsString()
			if v == "drop" {
				return drop
			} else if v == "keep" {
				return recordAndSample
			}
		}
	}
	return u.old.ShouldSample(p, state)
}

func (s *suited) TestUserSample() {
	old := s.sampled.UserSampler()
	s.sampled.SetUserSampler(testUserSampler{old})

	_, sp := s.sampled.Start(s.ctx, "test-span", trace.WithAttributes(attribute.String("user-sample", "drop")))
	s.Equal(false, sp.SpanContext().IsSampled())

	_, sp = s.sampled.Start(s.ctx, "test-span")
	s.Equal(true, sp.SpanContext().IsSampled())

	_, sp = s.sampled.Start(s.ctx, "test-span", trace.WithAttributes(attribute.String("user-sample", "keep")))
	s.Equal(true, sp.SpanContext().IsSampled())
	parsed, _ := tracestate.Parse(sp.SpanContext().TraceState().Get(galileoVendor))
	s.Equal(tracestate.StrategyUser, parsed.Sample.SampledStrategy)

	s.sampled.SetUserSampler(old)

}
