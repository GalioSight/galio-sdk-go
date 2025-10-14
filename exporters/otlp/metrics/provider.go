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

// Package metrics 用于将 OpenTelemetry 指标通过 push 方式上报到 OpenTelemetry collector
package metrics

import (
	"context"
	"errors"
	"sort"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	expres "galiosight.ai/galio-sdk-go/internal/resource"
	"galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/semconv"
)

// ErrCfgEnabled 表示 OpenTelemetry 推送未启用。
// 当 cfg.Enabled = false 时返回该错误。
var ErrCfgEnabled = errors.New("cfg.Enabled = false, no OpenTelemetry push enabled")

// NewMeterProvider 根据 Resource 和 OpenTelemetryPushConfig 构造 OpenTelemetry MeterProvider。
// 此函数用于在伽利略 SDK 中使用 OpenTelemetry 协议上报指标数据。
func NewMeterProvider(res model.Resource, cfg model.OpenTelemetryPushConfig, opt ...option) (
	*sdkmetric.MeterProvider, error,
) {
	if !cfg.Enable {
		return nil, ErrCfgEnabled
	}
	var o options
	// omp v1.0.0 metric schema 比较混乱，默认的 v1.0.0 的 metric schemaURL 是 prom-v1.0.0 即 prometheus 类型的 _target_
	// 这里直接默认只支持 v3.0.0
	o.schemaURL = semconv.SchemaURL
	for _, f := range opt {
		f(&o)
	}

	ctx := context.Background()
	tenant := res.TenantId // 租户信息。
	exporter, err := otlpmetrichttp.New(
		ctx, // 创建导出器。
		otlpmetrichttp.WithTemporalitySelector(getTemporalitySelector()),    // 差值。
		otlpmetrichttp.WithAggregationSelector(getAggregationSelector(nil)), // 分桶配置。
		otlpmetrichttp.WithEndpoint(cfg.Url),
		otlpmetrichttp.WithInsecure(),
		otlpmetrichttp.WithHeaders(
			map[string]string{
				model.TenantHeaderKey: tenant,
				model.TargetHeaderKey: res.Target,
				model.APIKeyHeaderKey: o.apiKey,
			},
		), // 租户信息
	)
	if err != nil {
		return nil, err
	}
	provider := sdkmetric.NewMeterProvider(
		// 构造 meter provider。
		sdkmetric.WithResource(expres.GenResource(o.schemaURL, &res, expres.SchemaTypeMetric)), // 构造来源属性。
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				exporter,
				sdkmetric.WithInterval(time.Second*20),
			), // push 间隔，一般设置 20 秒。
		),
	)
	return provider, nil
}

func getTemporalitySelector() sdkmetric.TemporalitySelector {
	return deltaSelector // 聚合窗口时间性，差值。
}

func deltaSelector(sdkmetric.InstrumentKind) metricdata.Temporality {
	return metricdata.DeltaTemporality
}

func getAggregationSelector(boundaries []float64) sdkmetric.AggregationSelector {
	if len(boundaries) == 0 {
		boundaries = []float64{0, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 5} // 默认分桶。
	}
	sort.Float64s(boundaries)
	return func(ik sdkmetric.InstrumentKind) sdkmetric.Aggregation {
		switch ik {
		case sdkmetric.InstrumentKindCounter,
			sdkmetric.InstrumentKindUpDownCounter,
			sdkmetric.InstrumentKindObservableCounter,
			sdkmetric.InstrumentKindObservableUpDownCounter:
			return sdkmetric.AggregationSum{}
		case sdkmetric.InstrumentKindGauge,
			sdkmetric.InstrumentKindObservableGauge:
			return sdkmetric.AggregationLastValue{}
		case sdkmetric.InstrumentKindHistogram:
			return sdkmetric.AggregationExplicitBucketHistogram{
				Boundaries: boundaries,
				NoMinMax:   false,
			}
		}
		return sdkmetric.AggregationSum{}
	}
}
