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

package metrics

import (
	"math"
	"math/rand"
	"testing"

	"galiosight.ai/galio-sdk-go/model"
	"github.com/stretchr/testify/assert"
)

func Test_sampler_sample(t *testing.T) {
	type fields struct {
		sampleMonitors []model.SampleMonitor
	}
	type args struct {
		monitorName    string
		labelsHashCode uint64
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantFactor float64
		wantSample bool
	}{
		{
			name:       "sampleMonitors is nil",
			fields:     fields{sampleMonitors: nil},
			args:       args{monitorName: "rpc_server", labelsHashCode: uint64(0x3000000000000000)},
			wantFactor: 1,
			wantSample: true,
		},
		{
			name: "not exist monitor",
			fields: fields{
				sampleMonitors: []model.SampleMonitor{
					{
						MonitorName: "rpc_client", SampleType: model.MetricsSampleType_METRICS_SAMPLE_TYPE_RAND,
						Fraction: 0.1,
					},
				},
			},
			args:       args{monitorName: "rpc_server", labelsHashCode: uint64(0x3000000000000000)},
			wantFactor: 1,
			wantSample: true,
		},
		{
			name: "not exist sample type",
			fields: fields{
				sampleMonitors: []model.SampleMonitor{
					{MonitorName: "rpc_server", SampleType: 99, Fraction: 0.1},
				},
			},
			args:       args{monitorName: "rpc_server", labelsHashCode: uint64(0x3000000000000000)},
			wantFactor: 1,
			wantSample: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				s := newSampler(tt.fields.sampleMonitors)
				m, sample := s.sample(tt.args.monitorName, tt.args.labelsHashCode)
				assert.Equalf(t, tt.wantFactor, m, "sample(%v, %v)", tt.args.monitorName, tt.args.labelsHashCode)
				assert.Equalf(t, tt.wantSample, sample, "sample(%v, %v)", tt.args.monitorName, tt.args.labelsHashCode)
			},
		)
	}
}

func Test_sampler_randSample(t *testing.T) {
	s := newSampler(
		[]model.SampleMonitor{
			{
				MonitorName: "rpc_server", SampleType: model.MetricsSampleType_METRICS_SAMPLE_TYPE_RAND, Fraction: 0.2,
			},
		},
	)
	cnt := 10000.0
	sampleCnt := 0.0
	for i := 0.0; i < cnt; i++ {
		if m, ok := s.sample("rpc_server", 0); ok {
			assert.Equal(t, float64(5), m)
			sampleCnt++
		}
	}
	assert.True(t, math.Abs(float64(sampleCnt/cnt)-0.2) < 0.1)
}

func Test_sampler_rowsSample(t *testing.T) {
	s := newSampler(
		[]model.SampleMonitor{
			{
				MonitorName: "rpc_server", SampleType: model.MetricsSampleType_METRICS_SAMPLE_TYPE_ROWS, Fraction: 0.5,
			},
		},
	)
	cnt := 10000.0
	sampleCnt := 0.0
	for i := 0.0; i < cnt; i++ {
		if m, ok := s.sample("rpc_server", rand.Uint64()); ok {
			assert.Equal(t, float64(2), m)
			sampleCnt++
		}
	}
	assert.True(t, math.Abs(float64(sampleCnt/cnt)-0.5) < 0.1)
}
