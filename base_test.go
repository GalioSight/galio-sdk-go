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

package galio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"

	"galiosight.ai/galio-sdk-go/components"
	"galiosight.ai/galio-sdk-go/configs"
	logsconfig "galiosight.ai/galio-sdk-go/configs/logs"
	traceconf "galiosight.ai/galio-sdk-go/configs/traces"
	"galiosight.ai/galio-sdk-go/exporters/otlp/traces"
	"galiosight.ai/galio-sdk-go/exporters/otlp/traces/tracestate"
	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/testdata"
	"galiosight.ai/galio-sdk-go/version"
)

func setUp(collectorAddr string) error {
	metricsCfg := configs.Metrics{
		SelfMonitor: model.SelfMonitor{
			Collector: model.Collector{Addr: collectorAddr},
		},
		Log:      logs.NopWrapper(),
		Resource: testdata.Resource,
		Processor: model.MetricsProcessor{
			Protocol:       "omp",
			WindowSeconds:  1,
			ClearSeconds:   100,
			ExpiresSeconds: 100,
			PointLimit:     10000,
		},
		Exporter: model.MetricsExporter{
			Protocol:      "otp",
			Collector:     model.Collector{Addr: collectorAddr},
			ThreadCount:   10,
			BufferSize:    10000,
			WindowSeconds: 1,
			PageSize:      1000,
		},
	}
	metricsProcessor, err := NewMetricsProcessor(&metricsCfg)
	if err != nil {
		return err
	}
	SetDefaultMetricsProcessor(metricsProcessor)
	return err
}

func TestClientMetrics(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"code":0,"msg":"success"}`))
			},
		),
	)
	defer ts.Close()
	err := setUp(ts.URL)
	assert.Nil(t, err)
	type args struct {
		clientMetrics *model.ClientMetrics
	}
	tests := []struct {
		args args
		name string
	}{
		{
			name: "test",
			args: args{
				clientMetrics: &model.ClientMetrics{
					Metrics: []model.ClientMetrics_Metric{
						{
							Name:        model.ClientMetrics_rpc_client_started_total,
							Value:       1,
							Aggregation: model.Aggregation_AGGREGATION_COUNTER,
						},
						{
							Name:        model.ClientMetrics_rpc_client_handled_total,
							Value:       1,
							Aggregation: model.Aggregation_AGGREGATION_COUNTER,
						},
						{
							Name:        model.ClientMetrics_rpc_client_handled_seconds,
							Value:       0.5,
							Aggregation: model.Aggregation_AGGREGATION_HISTOGRAM,
						},
					},
					RpcLabels: model.RPCLabels{
						Fields: []model.RPCLabels_Field{
							{
								Name:  model.RPCLabels_caller_service,
								Value: "abc",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				GetDefaultMetricsProcessor().ProcessClientMetrics(tt.args.clientMetrics)
				// 等 5 秒，让数据异步发送完成。
				time.Sleep(time.Second * 5)
				stats := GetDefaultMetricsProcessor().GetStats()
				assert.Equal(t, int64(0), stats.ReportErrorTotal.Load())
				assert.Equal(t, true, stats.ReportHandledTotal.Load() >= int64(1))
			},
		)
	}
}

func TestLogger(t *testing.T) {
	// 如果是使用 trpc-go-galileo 插件进行上报，会自动初始化，不需要调用 initLog。
	initLog(t)
	log := GetLogger()
	for i := 0; i < 10; i++ {
		idx := strconv.Itoa(i)
		log.Error("unit test msg:"+idx, zap.String("k0", "v0"), zap.String("k1", "v1"))
		log.Sync()
	}
	// 注意，伽利略远程日志是异步的，即使调用 Sync，也不会立即上报，需要等几秒才会上报。
	time.Sleep(time.Second * 1)
}

func TestReportEvent(t *testing.T) {
	// 注意，event 也是通过日志组件进行上报的，所以也需要初始化。
	// 如果是使用 trpc-go-galileo 插件进行上报，会自动初始化，不需要调用 initEventLog。
	initEventLog(t)
	for i := 0; i < 10; i++ {
		idx := strconv.Itoa(i)
		ReportEvent("galileo SDK event:"+idx, "andyning", "test", "demo", zap.String("event.foo", "bar"))
	}
	// 注意，伽利略远程日志是异步的，即使调用 Sync，也不会立即上报，需要等几秒才会上报。
	time.Sleep(time.Second * 1)
}

func TestSpanFromContext(t *testing.T) {
	a := assert.New(t)
	ts := SpanFromContext(context.Background()).TraceState()
	a.Equal("", ts.String())
}

func initLog(t *testing.T) {
	resource := defaultResource()
	conf := logsconfig.NewConfig(&resource)
	conf.Processor.Level = "debug"
	logger, err := NewLogger(conf)
	assert.Nil(t, err)
	assert.NotNil(t, logger)
	SetLogger(logger)
}

func initEventLog(t *testing.T) {
	resource := defaultResource()
	conf := logsconfig.NewConfig(&resource)
	conf.Processor.Level = "debug"
	logger, err := NewLogger(conf)
	assert.Nil(t, err)
	assert.NotNil(t, logger)
	SetEventLogger(logger)
}

func defaultResource() model.Resource {
	return model.Resource{
		Target:        "PCG-123.example.greeter",
		Namespace:     "Development",
		EnvName:       "test",
		Region:        "sh",
		Instance:      "127.0.0.1",
		ContainerName: "test.example.greeter.sh12345",
		Version:       version.Number,
		Platform:      "PCG-123",
		ObjectName:    "example.greeter",
		App:           "example",
		Server:        "greeter",
		SetName:       "set1.sh.1",
		FrameCode:     "trpc",
		ServiceName:   "trpc.example.greeter.Greeter",
		TenantId:      "galileo",
		Language:      "go",
		SdkName:       "galileo",
		City:          "sh",
	}
}

type suited struct {
	suite.Suite
	ctrl   *gomock.Controller
	ctx    context.Context
	any    gomock.Matcher
	tracer components.TracesExporter
}

type defered struct {
	traces.DeferredSampler
	st tracestate.Strategy
}

type tracerHook struct {
	components.TracesExporter
	deferr *defered
}

func (th *tracerHook) DeferredSampler() traces.DeferredSampler {
	return th.deferr
}

func (d *defered) DeferSample(span traces.Span) tracestate.Strategy {
	d.st = d.DeferredSampler.DeferSample(span)
	return d.st
}

func (s *suited) SetupSuite() {
	s.ctx = context.Background()
	s.ctrl = gomock.NewController(s.T())
	s.any = gomock.Any()
	res := defaultResource()
	c := traceconf.NewConfig(&res)
	c.Processor.EnableDeferredSample = true
	c.Processor.DeferredSampleError = true
	c.Processor.Sampler.EnableMinSample = false
	c.Processor.Sampler.ErrorFraction = 1.0
	tracer, err := traces.NewExporter(c)
	s.tracer = tracer
	s.NoError(err)
	SetDefaultTracesExporter(tracer)
}

func (s *suited) TearDownSuite() {
	s.ctrl.Finish()
}

func (s *suited) TestWithSpan() {
	tracer := &tracerHook{TracesExporter: s.tracer}
	tracer.deferr = &defered{DeferredSampler: tracer.TracesExporter.(traces.Tracer).DeferredSampler()}
	SetDefaultTracesExporter(tracer)

	WithSpan(
		s.ctx, "", func(ctx context.Context) error {
			SpanFromContext(ctx).SetStatus(codes.Error, "期望命中后置采样")
			return nil
		},
	)
	s.Equal(tracestate.StrategyError, tracer.deferr.st)
	SetDefaultTracesExporter(s.tracer)
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(suited))
}
