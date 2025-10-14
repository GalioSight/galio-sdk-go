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

package point

import (
	"testing"

	"galiosight.ai/galio-sdk-go/configs"
	"galiosight.ai/galio-sdk-go/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type otpMetric struct {
	name        string
	ranges      []string
	counts      []int64
	value       float64
	count       int64
	aggregation model.Aggregation
}

func newOTPMetric() *otpMetric {
	return &otpMetric{}
}

func (o *otpMetric) SetName(_ int, name string) {
	o.name = name
}

func (o *otpMetric) SetAggregation(_ int, a model.Aggregation) {
	o.aggregation = a
}

func (o *otpMetric) SetHistogram(_ int, sum float64, count int64, ranges []string, counts []int64) {
	o.value = sum
	o.count = count
	o.ranges = ranges
	o.counts = counts
}

func (o *otpMetric) SetAvg(_ int, sum float64, count int64) {
	o.value = sum
	o.count = count
}

func (o *otpMetric) SetCount(_ int, c int64) {
	o.count = c
}

func (o *otpMetric) SetValue(_ int, v float64) {
	o.value = v
}

type testPointParameter struct {
	name              string
	firstUpdate       []float64
	secondUpdate      []float64
	buckets           []float64
	firstOTP          otpMetric
	secondOTP         otpMetric
	secondExportCount int
	firstExportCount  int
	aggregation       model.Aggregation
	factor            float64
}

var testPointParameters = []testPointParameter{
	{
		name:        "test_set",
		aggregation: model.Aggregation_AGGREGATION_SET,
		firstUpdate: []float64{1, 3, 5, 2},
		firstOTP: otpMetric{
			name:        "test_set",
			aggregation: model.Aggregation_AGGREGATION_SET,
			value:       2,
		},
		firstExportCount: 1,
		secondUpdate:     []float64{4, 3, 8},
		secondOTP: otpMetric{
			name:        "test_set",
			aggregation: model.Aggregation_AGGREGATION_SET,
			value:       8,
		},
		secondExportCount: 1,
	},
	{
		name:        "test_sum",
		aggregation: model.Aggregation_AGGREGATION_SUM,
		firstUpdate: []float64{1, 3, 5, 2},
		firstOTP: otpMetric{
			name:        "test_sum",
			aggregation: model.Aggregation_AGGREGATION_SUM,
			value:       11,
		},
		firstExportCount: 1,
		secondUpdate:     []float64{4, 3, 8},
		secondOTP: otpMetric{
			name:        "test_sum",
			aggregation: model.Aggregation_AGGREGATION_SUM,
			value:       15,
		},
		secondExportCount: 1,
	},
	{
		name:        "test_avg",
		aggregation: model.Aggregation_AGGREGATION_AVG,
		firstUpdate: []float64{1, 3, 5},
		firstOTP: otpMetric{
			name:        "test_avg",
			aggregation: model.Aggregation_AGGREGATION_AVG,
			value:       9,
			count:       3,
		},
		firstExportCount: 2,
		secondUpdate:     []float64{4, 7},
		secondOTP: otpMetric{
			name:        "test_avg",
			aggregation: model.Aggregation_AGGREGATION_AVG,
			value:       11,
			count:       2,
		},
		secondExportCount: 2,
	},
	{
		name:        "test_max",
		aggregation: model.Aggregation_AGGREGATION_MAX,
		firstUpdate: []float64{1, 3, 5, 2},
		firstOTP: otpMetric{
			name:        "test_max",
			aggregation: model.Aggregation_AGGREGATION_MAX,
			value:       5,
		},
		firstExportCount: 1,
		secondUpdate:     []float64{4, 3, 8},
		secondOTP: otpMetric{
			name:        "test_max",
			aggregation: model.Aggregation_AGGREGATION_MAX,
			value:       8,
		},
		secondExportCount: 1,
	},
	{
		name:        "test_min",
		aggregation: model.Aggregation_AGGREGATION_MIN,
		firstUpdate: []float64{1, 3, 5, 2},
		firstOTP: otpMetric{
			name:        "test_min",
			aggregation: model.Aggregation_AGGREGATION_MIN,
			value:       1,
		},
		firstExportCount: 1,
		secondUpdate:     []float64{4, 3, 8},
		secondOTP: otpMetric{
			name:        "test_min",
			aggregation: model.Aggregation_AGGREGATION_MIN,
			value:       3,
		},
		secondExportCount: 1,
	},
	{
		name:        "test_histogram",
		aggregation: model.Aggregation_AGGREGATION_HISTOGRAM,
		firstUpdate: []float64{0.1, 0.3, 0.5, 1},
		firstOTP: otpMetric{
			name:        "test_histogram",
			aggregation: model.Aggregation_AGGREGATION_HISTOGRAM,
			value:       1.9,
			count:       4,
			ranges: []string{
				"1.000e-01...2.000e-01", "2.000e-01...5.000e-01",
				"5.000e-01...1.000e+00", "1.000e+00...+Inf",
			},
			counts: []int64{1, 1, 1, 1},
		},
		firstExportCount: 1 + 1 + 4,
		secondUpdate:     []float64{0.001, 0.005, 0.01, 0.3},
		secondOTP: otpMetric{
			name:        "test_histogram",
			aggregation: model.Aggregation_AGGREGATION_HISTOGRAM,
			value:       0.316,
			count:       4,
			ranges: []string{
				"0.000e+00...1.000e-01",
				"2.000e-01...5.000e-01",
			},
			counts: []int64{3, 1},
		},
		secondExportCount: 1 + 1 + 2,
		buckets:           []float64{0.1, 0.2, 0.5, 1},
	},
	{
		name:        "test_counter",
		aggregation: model.Aggregation_AGGREGATION_COUNTER,
		firstUpdate: []float64{1, 2, 3},
		firstOTP: otpMetric{
			name:        "test_counter",
			aggregation: model.Aggregation_AGGREGATION_COUNTER,
			count:       6,
		},
		firstExportCount: 1,
		secondUpdate:     []float64{4, 5, 6},
		secondOTP: otpMetric{
			name:        "test_counter",
			aggregation: model.Aggregation_AGGREGATION_COUNTER,
			count:       15,
		},
		secondExportCount: 1,
	},
}

func Test_Point(t *testing.T) {
	for _, tt := range testPointParameters {
		t.Run(
			tt.aggregation.String(), func(t *testing.T) {
				// 构造点。
				point := Get(tt.aggregation, tt.name)
				if tt.aggregation == model.Aggregation_AGGREGATION_HISTOGRAM {
					point.SetBucket(func() *configs.Bucket { return configs.NewBucket(tt.buckets) })
				}
				defer Put(point)
				assert.NotNil(t, point)
				// 第一次更新数据。
				for i := range tt.firstUpdate {
					point.Update(tt.firstUpdate[i])
				}
				otp := newOTPMetric()
				// 第一次导出。
				firstExportCount, err := point.ToOTP(otp, 0)
				assert.NoError(t, err)
				assert.Equal(t, tt.firstExportCount, firstExportCount)
				// 第一次断言。
				assert.Equal(t, tt.firstOTP, *otp)
				// 第二次更新数据。
				for i := range tt.secondUpdate {
					point.Update(tt.secondUpdate[i])
				}
				// 第二次导出。
				otp = newOTPMetric()
				secondExportCount, err := point.ToOTP(otp, 0)
				assert.NoError(t, err)
				assert.Equal(t, tt.secondExportCount, secondExportCount)
				// 第二次断言。
				assert.Equal(t, tt.secondOTP, *otp)
			},
		)
	}
}

var testSamplePointParameters = []testPointParameter{
	{
		name:        "test_set",
		aggregation: model.Aggregation_AGGREGATION_SET,
		firstUpdate: []float64{1, 3, 5, 2},
		firstOTP: otpMetric{
			name:        "test_set",
			aggregation: model.Aggregation_AGGREGATION_SET,
			value:       2,
		},
		firstExportCount: 1,
		secondUpdate:     []float64{4, 3, 8},
		secondOTP: otpMetric{
			name:        "test_set",
			aggregation: model.Aggregation_AGGREGATION_SET,
			value:       8,
		},
		secondExportCount: 1,
		factor:            3,
	},
	{
		name:        "test_sum",
		aggregation: model.Aggregation_AGGREGATION_SUM,
		firstUpdate: []float64{1, 3, 5, 2},
		firstOTP: otpMetric{
			name:        "test_sum",
			aggregation: model.Aggregation_AGGREGATION_SUM,
			value:       33,
		},
		firstExportCount: 1,
		secondUpdate:     []float64{4, 3, 8},
		secondOTP: otpMetric{
			name:        "test_sum",
			aggregation: model.Aggregation_AGGREGATION_SUM,
			value:       45,
		},
		secondExportCount: 1,
		factor:            3,
	},
	{
		name:        "test_avg",
		aggregation: model.Aggregation_AGGREGATION_AVG,
		firstUpdate: []float64{1, 3, 5},
		firstOTP: otpMetric{
			name:        "test_avg",
			aggregation: model.Aggregation_AGGREGATION_AVG,
			value:       27,
			count:       9,
		},
		firstExportCount: 2,
		secondUpdate:     []float64{4, 7},
		secondOTP: otpMetric{
			name:        "test_avg",
			aggregation: model.Aggregation_AGGREGATION_AVG,
			value:       33,
			count:       6,
		},
		secondExportCount: 2,
		factor:            3,
	},
	{
		name:        "test_max",
		aggregation: model.Aggregation_AGGREGATION_MAX,
		firstUpdate: []float64{1, 3, 5, 2},
		firstOTP: otpMetric{
			name:        "test_max",
			aggregation: model.Aggregation_AGGREGATION_MAX,
			value:       5,
		},
		firstExportCount: 1,
		secondUpdate:     []float64{4, 3, 8},
		secondOTP: otpMetric{
			name:        "test_max",
			aggregation: model.Aggregation_AGGREGATION_MAX,
			value:       8,
		},
		secondExportCount: 1,
		factor:            3,
	},
	{
		name:        "test_min",
		aggregation: model.Aggregation_AGGREGATION_MIN,
		firstUpdate: []float64{1, 3, 5, 2},
		firstOTP: otpMetric{
			name:        "test_min",
			aggregation: model.Aggregation_AGGREGATION_MIN,
			value:       1,
		},
		firstExportCount: 1,
		secondUpdate:     []float64{4, 3, 8},
		secondOTP: otpMetric{
			name:        "test_min",
			aggregation: model.Aggregation_AGGREGATION_MIN,
			value:       3,
		},
		secondExportCount: 1,
		factor:            3,
	},
	{
		name:        "test_histogram",
		aggregation: model.Aggregation_AGGREGATION_HISTOGRAM,
		firstUpdate: []float64{0.15, 0.3, 0.5, 1},
		firstOTP: otpMetric{
			name:        "test_histogram",
			aggregation: model.Aggregation_AGGREGATION_HISTOGRAM,
			value:       5.85,
			count:       12,
			ranges: []string{
				"1.000e-01...2.000e-01", "2.000e-01...5.000e-01",
				"5.000e-01...1.000e+00", "1.000e+00...+Inf",
			},
			counts: []int64{3, 3, 3, 3},
		},
		firstExportCount: 1 + 1 + 4,
		secondUpdate:     []float64{0.001, 0.005, 0.01},
		secondOTP: otpMetric{
			name:        "test_histogram",
			aggregation: model.Aggregation_AGGREGATION_HISTOGRAM,
			value:       0.048,
			count:       9,
			ranges: []string{
				"0.000e+00...1.000e-01",
			},
			counts: []int64{9},
		},
		secondExportCount: 1 + 1 + 1,
		buckets:           []float64{0.1, 0.2, 0.5, 1},
		factor:            3,
	},
	{
		name:        "test_counter",
		aggregation: model.Aggregation_AGGREGATION_COUNTER,
		firstUpdate: []float64{1, 2, 3},
		firstOTP: otpMetric{
			name:        "test_counter",
			aggregation: model.Aggregation_AGGREGATION_COUNTER,
			count:       18,
		},
		firstExportCount: 1,
		secondUpdate:     []float64{4, 5, 6},
		secondOTP: otpMetric{
			name:        "test_counter",
			aggregation: model.Aggregation_AGGREGATION_COUNTER,
			count:       45,
		},
		secondExportCount: 1,
		factor:            3,
	},
}

func Test_Point_Sample(t *testing.T) {
	for _, tt := range testSamplePointParameters {
		t.Run(
			tt.aggregation.String(), func(t *testing.T) {
				// 构造点。
				point := Get(tt.aggregation, tt.name)
				if tt.aggregation == model.Aggregation_AGGREGATION_HISTOGRAM {
					point.SetBucket(func() *configs.Bucket { return configs.NewBucket(tt.buckets) })
				}
				defer Put(point)
				assert.NotNil(t, point)
				// 第一次更新数据。
				for i := range tt.firstUpdate {
					point.Update(tt.firstUpdate[i])
				}
				otp := newOTPMetric()
				// 第一次导出。
				err := point.Change(tt.factor)
				assert.NoError(t, err)
				firstExportCount, err := point.ToOTP(otp, 0)
				assert.NoError(t, err)
				assert.Equal(t, tt.firstExportCount, firstExportCount)
				// 第一次断言。
				assert.Equal(t, tt.firstOTP, *otp)
				// 第二次更新数据。
				for i := range tt.secondUpdate {
					point.Update(tt.secondUpdate[i])
				}
				// 第二次导出。
				otp = newOTPMetric()
				err = point.Change(tt.factor)
				assert.NoError(t, err)
				secondExportCount, err := point.ToOTP(otp, 0)
				assert.NoError(t, err)
				assert.Equal(t, tt.secondExportCount, secondExportCount)
				// 第二次断言。
				assert.Equal(t, tt.secondOTP, *otp)
			},
		)
	}
}

func Test_roundInt64(t *testing.T) {
	type args struct {
		v float64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{name: "1", args: args{v: 1}, want: 1},
		{name: "1.1", args: args{v: 1.1}, want: 1},
		{name: "1.5", args: args{v: 1.5}, want: 2},
		{name: "1.9", args: args{v: 1.9}, want: 2},
		{name: "-1", args: args{v: -1}, want: 0},
		{name: "-1.1", args: args{v: -1.1}, want: 0},
		{name: "-1.5", args: args{v: -1.5}, want: -1},
		{name: "-1.9", args: args{v: -1.9}, want: -1},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				assert.Equalf(t, tt.want, roundInt64(tt.args.v), "roundInt64(%v)", tt.args.v)
			},
		)
	}
}

func TestPoint_Update(t *testing.T) {
	defer func() {
		require.Nil(t, recover())
	}()
	var p *Point
	p.Update(1)
	p = &Point{}
	p.Update(2)
}
