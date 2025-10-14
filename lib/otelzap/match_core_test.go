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
	"log"
	"regexp"
	"strings"
	"testing"

	"galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const sampled = "sampled"

func withContextSampleLevel(s coreStrategy) zap.Option {
	return ContextWith(
		func(c zapcore.Core, ctx context.Context) zapcore.Core {
			if m, _ := c.(*MatchCore); m != nil {
				c = m.Core
			}
			ret := &MatchCore{Core: c, strategy: s}
			if v := ctx.Value(sampled); v != nil {
				log.Println(v.(bool))
				ret.matched = v.(bool)
			}
			return ret
		},
	)
}

func TestMatchCore(t *testing.T) {
	tests := []struct {
		cfg  model.LogsProcessor
		want string
	}{
		{
			model.LogsProcessor{}, `{"level":"info","msg":"info","tag2":"value2"}` + "\n" +
				`{"level":"warn","msg":"warn","sampled":"true","tag3":"value3"}` + "\n",
		},
		{
			model.LogsProcessor{OnlyTraceLog: true},
			`{"level":"warn","msg":"warn","sampled":"true","tag3":"value3"}` + "\n",
		},
		{
			model.LogsProcessor{MustLogTraced: true},
			`{"level":"debug","msg":"sampled","testtag":"testvalue","tag1":"value1"}` + "\n" +
				`{"level":"info","msg":"info","tag2":"value2"}` + "\n" +
				`{"level":"warn","msg":"warn","sampled":"true","tag3":"value3"}` + "\n",
		},
		{
			model.LogsProcessor{OnlyTraceLog: true, MustLogTraced: true},
			`{"level":"debug","msg":"sampled","testtag":"testvalue","tag1":"value1"}` + "\n" +
				`{"level":"warn","msg":"warn","sampled":"true","tag3":"value3"}` + "\n",
		},
		{
			model.LogsProcessor{OnlyTraceLog: true, MustLogTraced: true, LogTracedType: string(model.LogTracedDyeing)},
			`{"level":"debug","msg":"dyeing","sampled":"true","tag4":"value4"}` + "\n",
		},
	}

	ctx := context.Background()
	tsre := regexp.MustCompile(`"ts":[^,]+,`)
	a := assert.New(t)
	for _, test := range tests {
		t.Run(
			"", func(t *testing.T) {
				sink := &strings.Builder{}
				logger := zap.New(
					newCore(zapcore.AddSync(sink), zap.NewAtomicLevelAt(zap.InfoLevel)),
					toZapOptions(&configs.Logs{Processor: test.cfg})...,
				).With(Context(ctx))
				logger.Debug("debug", zap.String("tag1", "value1"))
				logger.With(Context(sampleContext(true)), zap.String("testtag", "testvalue")).Debug(
					"sampled", zap.String("tag1", "value1"),
				)
				logger.Info("info", zap.String("tag2", "value2"))
				logger.With(Context(sampleContext(true)), zap.String("sampled", "true")).Warn(
					"warn", zap.String("tag3", "value3"),
				)
				if test.cfg.LogTracedType == string(model.LogTracedDyeing) {
					logger.With(Context(dyeingContext(true)), zap.String("sampled", "true")).Debug(
						"dyeing", zap.String("tag4", "value4"),
					)
				}
				a.Equal(test.want, tsre.ReplaceAllString(sink.String(), ""))
			},
		)
	}
}
