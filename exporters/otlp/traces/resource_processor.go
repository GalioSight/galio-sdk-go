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

	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	semconv "galiosight.ai/galio-sdk-go/semconv/v1.0.0"
)

// ResourceProcessor 添加 Resource 通用属性到 span 的 tags 属性 (当前只是上报了环境量)
type ResourceProcessor struct {
	next      sdktrace.SpanProcessor
	o         ResourceSpanProcessorOptions
	nameSpace string
	envName   string
}

// NewResourceProcessor 创建一个 ResourceProcessor
func NewResourceProcessor(next sdktrace.SpanProcessor, options ...ResourceSpanProcessorOption) *ResourceProcessor {
	o := ResourceSpanProcessorOptions{}
	for _, opt := range options {
		opt(&o)
	}
	return &ResourceProcessor{
		next:      next,
		nameSpace: o.NameSpace,
		envName:   o.EnvName,
	}
}

// ResourceSpanProcessorOption ResourceSpanProcessor Option helper
type ResourceSpanProcessorOption func(o *ResourceSpanProcessorOptions)

// ResourceSpanProcessorOptions ResourceSpanProcessor 控制项
type ResourceSpanProcessorOptions struct {
	NameSpace string
	EnvName   string
}

// WithNameSpace 设置
func WithNameSpace(ns string) ResourceSpanProcessorOption {
	return func(o *ResourceSpanProcessorOptions) {
		o.NameSpace = ns
	}
}

// WithEnvName 设置
func WithEnvName(env string) ResourceSpanProcessorOption {
	return func(o *ResourceSpanProcessorOptions) {
		o.EnvName = env
	}
}

// OnStart 在 Span 启动时被调用
func (p *ResourceProcessor) OnStart(ctx context.Context, s sdktrace.ReadWriteSpan) {
	s.SetAttributes(
		semconv.TrpcNamespaceKey.String(p.nameSpace),
		semconv.TrpcEnvnameKey.String(p.envName),
	)
}

// OnEnd 在 Span 结束时被调用
func (p *ResourceProcessor) OnEnd(s sdktrace.ReadOnlySpan) {
	if p.next != nil {
		p.next.OnEnd(s)
	}
}

// Shutdown 关闭
func (p *ResourceProcessor) Shutdown(ctx context.Context) error {
	if p.next != nil {
		return p.next.Shutdown(ctx)
	}
	return nil
}

// ForceFlush 强制刷新
func (p *ResourceProcessor) ForceFlush(ctx context.Context) error {
	if p.next != nil {
		return p.next.ForceFlush(ctx)
	}
	return nil
}
