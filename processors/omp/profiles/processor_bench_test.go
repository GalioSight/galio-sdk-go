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

package profiles

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"sync"
	"syscall"
	"testing"
	"time"

	"galiosight.ai/galio-sdk-go/components"
	profileconf "galiosight.ai/galio-sdk-go/configs/profiles"
	traceconf "galiosight.ai/galio-sdk-go/configs/traces"
	"galiosight.ai/galio-sdk-go/exporters/otlp/traces"
	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/version"
	pprofile "github.com/google/pprof/profile"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func BenchmarkTraceProfile(b *testing.B) {
	benchFunc := func(b *testing.B, req *testReq) {
		enableProfile := os.Getenv("ENABLE_PROFILE") == "true"
		profileExporter := newFakeExporter()
		app, err := newTestApp(enableProfile, profileExporter)
		if err != nil {
			b.Errorf("faild to create a new app, err: %+v", err)
			return
		}
		app.start(b)
		defer app.stop(b)

		b.ResetTimer()
		var (
			stopWait     sync.WaitGroup
			concurrency  = runtime.GOMAXPROCS(0)
			cpuTimeStart = getRusage(b)
		)
		for i := 0; i < concurrency; i++ {
			stopWait.Add(1)
			go func() {
				defer stopWait.Done()
				for j := 0; j < b.N; j++ {
					app.SendReq(b, req)
				}
			}()
		}
		stopWait.Wait()
		cpuTime := getRusage(b) - cpuTimeStart
		b.StopTimer()

		app.stop(b)
		prof, profBytes := getFirstProf(profileExporter.GetProfileData())

		b.ReportMetric(float64(numOfSamples(prof))/float64(b.N*concurrency), "pprof-samples/op")
		b.ReportMetric(float64(len(profBytes))/float64(b.N*concurrency), "pprof-B/op")
		b.ReportMetric(float64(cpuTime)/float64(b.N*concurrency), "cpu-ns/op")
	}

	b.Run(
		"Do nothing", func(b *testing.B) {
			benchFunc(
				b, &testReq{
					CPUDuration: int64(0 * time.Millisecond),
					SQLDuration: int64(0 * time.Millisecond),
				},
			)
		},
	)

	b.Run(
		"cpu-bound", func(b *testing.B) {
			benchFunc(
				b, &testReq{
					CPUDuration: int64(90 * time.Millisecond),
					SQLDuration: int64(10 * time.Millisecond),
				},
			)
		},
	)

	b.Run(
		"io-bound", func(b *testing.B) {
			benchFunc(
				b, &testReq{
					CPUDuration: int64(10 * time.Millisecond),
					SQLDuration: int64(90 * time.Millisecond),
				},
			)
		},
	)
}

type testApp struct {
	httpServer    *httptest.Server
	tracer        components.TracesExporter
	profiler      components.ProfilesProcessor
	enableProfile bool
}

type testReq struct {
	CPUDuration int64 `json:"cpu_duration"`
	SQLDuration int64 `json:"sql_duration"`
}

// 资源描述，见文档 https://galiosight.ai/eco/blob/go/sdk/base/v0.3.20/semantic_conventions/resource/service.yaml
var resource = &model.Resource{
	Target:        "Galileo-Dial.galileo.SDK", // 观测对象（模块）的唯一标识 ID，需要全局唯一，避免与其他的数据混淆。
	Namespace:     "Development",              // 物理环境，只能是 Production 和 Development。
	EnvName:       "test",                     // 用户环境
	SetName:       "set1.sz.1",
	Region:        "sz",                   // 地域
	Instance:      "aaa.bbb.ccc.ddd",      // 实例 ip
	Node:          "cls-as9z3nec-2",       // 节点
	ContainerName: "test.galileo.SDK.sz1", // 容器
	Version:       version.Number,         // SDK 版本号
	Platform:      "Galileo-Dial",         // 平台
	App:           "galileo",
	Server:        "SDK",
	ServiceName:   "trpc.galileo.SDK.Demo",
	ObjectName:    "galileo.SDK", // 对象名称
	FrameCode:     "trpc",
	Language:      "go",
	SdkName:       "galileo",
}

func newTestApp(enableProfile bool, exporter components.ProfilesExporter) (*testApp, error) {
	a := &testApp{
		enableProfile: enableProfile,
	}
	tracer, err := newTestTracer()
	if err != nil {
		return nil, err
	}
	a.tracer = tracer
	profiler, err := newTestProfiler(a.enableProfile, exporter)
	if err != nil {
		return nil, err
	}
	a.profiler = profiler
	return a, nil
}

func (a *testApp) start(tb testing.TB) {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", a.testHandler)
	a.httpServer = httptest.NewServer(mux)
	if a.enableProfile {
		a.profiler.Start()
	}
}

func (a *testApp) stop(tb testing.TB) {
	a.profiler.Shutdown()
	a.httpServer.Close()
}

func (a *testApp) testHandler(w http.ResponseWriter, req *http.Request) {
	reqBody := &testReq{}
	if err := json.NewDecoder(req.Body).Decode(reqBody); err != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
	}
	parentCtx := context.Background()
	ctx, sp := a.tracer.Start(parentCtx, "test", trace.WithAttributes(attribute.String("foo", "bar")))
	defer sp.End()
	// 模拟 io-bound 场景
	a.fakeSQLQuery(ctx, "SELECT * FROM foo", time.Duration(reqBody.SQLDuration))
	// 模拟 cpu-bound 场景
	a.cpuHog(time.Duration(reqBody.CPUDuration))
}

func (a *testApp) SendReq(tb testing.TB, req *testReq) {
	body, err := json.Marshal(req)
	require.NoError(tb, err)
	url := a.httpServer.URL + "/test"
	resp, err := http.Post(url, "text/plain", bytes.NewReader(body))
	require.NoError(tb, err)
	defer resp.Body.Close()
}

func newTestTracer() (components.TracesExporter, error) {
	cfg := traceconf.NewConfig(resource)
	cfg.Processor.Sampler.Fraction = 0.1
	cfg.Processor.EnableProfile = true
	cfg.Log.SetLevel(logs.LevelNone)
	return traces.NewExporter(cfg)
}

func newTestProfiler(
	enableProfile bool,
	exporter components.ProfilesExporter,
) (components.ProfilesProcessor, error) {
	cfg := profileconf.NewConfig(resource)
	cfg.Log.SetLevel(logs.LevelNone)
	cfg.Enable = enableProfile
	cfg.Processor.ProfileTypes = []string{"cpu"}
	processor, err := NewProcessor(cfg, exporter)
	if err != nil {
		return nil, err
	}
	return processor, nil
}

func (a *testApp) cpuHog(d time.Duration) {
	stop := time.After(d)
	i := 0
	for {
		select {
		case <-stop:
			return
		default:
			// 消耗 cpu
			i++
			fmt.Fprintf(io.Discard, "%d", i)
		}
	}
}

func (a *testApp) fakeSQLQuery(ctx context.Context, sql string, d time.Duration) {
	_, span := a.tracer.Start(
		ctx, "sql.query", trace.WithAttributes(
			attribute.String("sql", sql),
		),
	)
	defer span.End()
	time.Sleep(d)
}

func getRusage(tb testing.TB) time.Duration {
	var rusage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &rusage); err != nil {
		tb.Fatal(err)
	}
	return time.Duration(rusage.Stime.Nano()) + time.Duration(rusage.Utime.Nano())
}

func getFirstProf(batch []*model.ProfilesBatch) (*pprofile.Profile, []byte) {
	if len(batch) == 0 {
		return nil, nil
	}
	firstBatch := batch[0]
	if len(firstBatch.Profiles) == 0 {
		return nil, nil
	}
	firstProf := firstBatch.Profiles[0]
	prof, err := pprofile.ParseData(firstProf.Data)
	if err != nil {
		return nil, nil
	}
	return prof, firstProf.Data
}

func numOfSamples(prof *pprofile.Profile) int {
	if prof == nil {
		return 0
	}
	return len(prof.Sample)
}
