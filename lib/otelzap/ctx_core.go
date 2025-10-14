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

// Package otelzap ...
package otelzap

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ctxWithFunc func(zapcore.Core, context.Context) zapcore.Core

func newCtxCore(core zapcore.Core, fn ctxWithFunc) zapcore.Core {
	if cc, ok := core.(*ctxCore); ok {
		return &ctxCore{ctx: fn, Core: cc.Core}
	}
	return &ctxCore{ctx: fn, Core: core}
}

// ContextWith 低级设置函数，支持用户自行实现扩展
func ContextWith(fn ctxWithFunc) zap.Option {
	return zap.WrapCore(func(core zapcore.Core) zapcore.Core { // 创建 core 的时候执行
		return newCtxCore(core, fn)
	})
}

// ctxCore 支持 设置 traceid 时同时根据 ctx 派生出具体的 log 实例
type ctxCore struct {
	zapcore.Core
	ctx ctxWithFunc
}

// 防止 JSON 序列化打印 context 内容
type ctxValue struct {
	ctx context.Context
}

var _ fmt.Stringer = (*ctxValue)(nil)

func (cv *ctxValue) String() string {
	// just print nothing, we want to ignore it
	return ""
}

const fieldCtx = "_ctx"

// Context like zap.String 输出一个合法的 ctx 透传对象
func Context(ctx context.Context) zap.Field {
	// 这里返回任何 Type 都是没有意义的，trpc 的 log.With 会拆开 Key，Value，重新用 zap.Any 组装。限制死了
	return zap.Stringer(fieldCtx, &ctxValue{ctx: ctx})
}

func (c *ctxCore) With(fields []zapcore.Field) zapcore.Core {
	var ctx context.Context
	for i, f := range fields {
		if ctx = takeCtx(f); ctx != nil {
			fields[i] = zap.Skip() // skip ctx field
			break
		}
	}
	ret := &ctxCore{ctx: c.ctx, Core: c.Core}
	n := len(fields)
	if ctx != nil {
		n -= 1
	}
	if n > 0 { // 除了 ctx 外还有 n 个 field，需要 clone
		ret.Core = ret.Core.With(fields)
	}
	if ctx != nil && c.ctx != nil {
		ret.Core = c.ctx(ret.Core, ctx)
	}
	return ret
}

func takeCtx(v zapcore.Field) context.Context {
	if v.Key != fieldCtx || v.Interface == nil {
		return nil
	}
	val, _ := v.Interface.(*ctxValue)
	if val == nil {
		return nil
	}
	return val.ctx
}
