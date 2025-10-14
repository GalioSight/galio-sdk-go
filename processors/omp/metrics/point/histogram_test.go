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
	"galiosight.ai/galio-sdk-go/lib/strings"
	"galiosight.ai/galio-sdk-go/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_updateHistogram(t *testing.T) {
	tests := []struct {
		want    model.NormalMetricOTP
		name    string
		buckets []float64
		v       float64
	}{
		{
			name:    "Test_updateHistogram_不配置桶",
			buckets: []float64{},
			v:       1,
			want: model.NormalMetricOTP{
				Metric: &model.MetricOTP{
					Name:        "Test_updateHistogram_不配置桶",
					V:           model.NewOTPHistogram(1, 1, []int64{}, []string{}),
					Aggregation: model.Aggregation_AGGREGATION_HISTOGRAM,
				},
			},
		},
		{
			name:    "Test_updateHistogram_中间取左",
			buckets: []float64{0, 0.01, 0.02, 0.05, 0.1},
			v:       0.04,
			want: model.NormalMetricOTP{
				Metric: &model.MetricOTP{
					Name: "Test_updateHistogram_中间取左",
					V: model.NewOTPHistogram(
						0.04, 1,
						[]int64{1},
						[]string{
							"2.000e-02...5.000e-02",
						},
					),
					Aggregation: model.Aggregation_AGGREGATION_HISTOGRAM,
				},
			},
		},
		{
			name:    "Test_updateHistogram_过大取最大",
			buckets: []float64{0, 0.01, 0.02, 0.05, 0.1},
			v:       0.11,
			want: model.NormalMetricOTP{
				Metric: &model.MetricOTP{
					Name: "Test_updateHistogram_过大取最大",
					V: model.NewOTPHistogram(
						0.11, 1,
						[]int64{1},
						[]string{
							"1.000e-01...+Inf",
						},
					),
					Aggregation: model.Aggregation_AGGREGATION_HISTOGRAM,
				},
			},
		},
		{
			name:    "Test_updateHistogram_过小取最小",
			buckets: []float64{0.01, 0.02, 0.05, 0.1},
			v:       0.001,
			want: model.NormalMetricOTP{
				Metric: &model.MetricOTP{
					Name: "Test_updateHistogram_过小取最小",
					V: model.NewOTPHistogram(
						0.001, 1,
						[]int64{1},
						[]string{
							"0.000e+00...1.000e-02",
						},
					),
					Aggregation: model.Aggregation_AGGREGATION_HISTOGRAM,
				},
			},
		}, {
			name:    "Test_updateHistogram_等于边界取右_左闭右开区间",
			buckets: []float64{0.01, 0.02, 0.05, 0.1},
			v:       0.02,
			want: model.NormalMetricOTP{
				Metric: &model.MetricOTP{
					Name: "Test_updateHistogram_等于边界取右_左闭右开区间",
					V: model.NewOTPHistogram(
						0.02, 1,
						[]int64{1},
						[]string{
							"2.000e-02...5.000e-02",
						},
					),
					Aggregation: model.Aggregation_AGGREGATION_HISTOGRAM,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				p := get(tt.name, model.Aggregation_AGGREGATION_HISTOGRAM, []option{initHistogram})
				p.SetBucket(func() *configs.Bucket { return configs.NewBucket(tt.buckets) })
				updateHistogram(p, tt.v)
				otp := model.NewNormalMetricsOTP()
				_, err := p.ToOTP(otp, 0)
				require.NoError(t, err)
				assert.Equal(t, tt.want, *otp)
			},
		)
	}
}

func Test_lowerBound(t *testing.T) {
	type args struct {
		array  []float64
		target float64
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "test0",
			args: args{array: []float64{1, 10, 10, 10, 20, 30, 40, 50, 60}, target: 10},
			want: 1,
		},
		{
			name: "test1",
			args: args{array: []float64{1, 10, 20, 30, 40, 50, 60}, target: 1},
			want: 0,
		},
		{
			name: "test2",
			args: args{array: []float64{1, 10, 20, 30, 40, 50, 60}, target: 49},
			want: 5,
		}, {
			name: "test2",
			args: args{array: []float64{1, 10, 20, 30, 40, 50, 60}, target: 50},
			want: 5,
		}, {
			name: "test33",
			args: args{array: []float64{1, 10, 20, 30, 40, 50, 60}, target: 51},
			want: 6,
		},
		{
			name: "test3",
			args: args{array: []float64{1, 10, 20, 30, 40, 50, 60}, target: 60},
			want: 6,
		},
		{
			name: "test4",
			args: args{array: []float64{1, 10, 20, 30, 40, 50, 60}, target: 61},
			want: 7,
		},
		{
			name: "test5",
			args: args{array: []float64{1, 10, 20, 30, 40, 50, 60}, target: 2},
			want: 1,
		},
		{
			name: "test6",
			args: args{array: []float64{1, 10, 20, 30, 40, 50, 60}, target: 59},
			want: 6,
		},
		{
			name: "test7",
			args: args{array: []float64{1, 10, 20, 30, 40, 50, 60, 60, 60}, target: 60},
			want: 6,
		},
		{
			name: "test8",
			args: args{array: []float64{}, target: 1},
			want: 0,
		},
		{
			name: "test9",
			args: args{array: []float64{-5, -2, -2, -2}, target: -4},
			want: 1,
		},
		{
			name: "test10",
			args: args{array: []float64{1, 5, 5, 5, 5, 7, 7, 7, 7, 9}, target: 0},
			want: 0,
		},
		{
			name: "test11",
			args: args{array: []float64{1, 2, 5, 8, 10}, target: 11},
			want: 5,
		},
		{
			name: "test12",
			args: args{array: []float64{1, 2, 2, 2, 3, 3, 10}, target: 3},
			want: 4,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := lowerBound(tt.args.array, tt.args.target); got != tt.want {
					t.Errorf("lowerBound() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_handleBucketChange(t *testing.T) {
	buckets := configs.NewBucket([]float64{0, 1, 2, 5, 10}) // 初始桶配置 A，5 个桶。
	p := get("point_name", model.Aggregation_AGGREGATION_HISTOGRAM, []option{initHistogram})
	p.SetBucket(func() *configs.Bucket { return buckets })
	p.Update(1.5) // 使用桶配置 A。

	buckets = configs.NewBucket([]float64{0, 1, 2, 5}) // 更新桶配置 B，少 1 个桶。

	otp := model.NewNormalMetricsOTP()
	count, err := p.ToOTP(otp, 0) // 导出之后，分桶变化检测。
	require.NoError(t, err)
	require.Equal(t, 3, count) // 1sum 1count 5 分桶，其中 1 个桶有数据

	p.Update(1.5) // 使用桶配置 B。

	buckets = configs.NewBucket([]float64{}) // 更新桶配置 C，删除所有桶。

	otp = model.NewNormalMetricsOTP()
	count, err = p.ToOTP(otp, 0)
	require.NoError(t, err)
	require.Equal(t, 3, count) // 1sum 1count 4 分桶，其中 1 个桶有数据

	p.Update(1.5) // 使用桶配置 C。

	otp = model.NewNormalMetricsOTP()
	count, err = p.ToOTP(otp, 0)
	require.NoError(t, err)
	require.Equal(t, 2, count) // 1sum 1count 0 分桶
}

func Test_changeHistogram(t *testing.T) {
	type args struct {
		buckets []float64
		values  []float64
		factor  float64
	}
	tests := []struct {
		name string
		args args
		want model.NormalMetricOTP
	}{
		{
			name: "不配置桶",
			args: args{
				buckets: nil,
				values:  []float64{1, 2, 3},
				factor:  3,
			},
			want: model.NormalMetricOTP{
				Metric: &model.MetricOTP{
					Name:        "不配置桶",
					V:           model.NewOTPHistogram(18, 9, []int64{}, []string{}),
					Aggregation: model.Aggregation_AGGREGATION_HISTOGRAM,
				},
			},
		},
		{
			name: "采样复原",
			args: args{
				buckets: []float64{0, 1, 2, 3, 5},
				values:  []float64{0.5, 1, 1.5, 2.5, 3.5, 5, 7},
				factor:  3,
			},
			want: model.NormalMetricOTP{
				Metric: &model.MetricOTP{
					Name: "采样复原",
					V: model.NewOTPHistogram(
						63, 21,
						[]int64{3, 6, 3, 3, 6},
						[]string{
							strings.VMRangeFloatToString(
								0,
							) + strings.VMRangeSeparator + strings.VMRangeFloatToString(
								1,
							),
							strings.VMRangeFloatToString(
								1,
							) + strings.VMRangeSeparator + strings.VMRangeFloatToString(
								2,
							),
							strings.VMRangeFloatToString(
								2,
							) + strings.VMRangeSeparator + strings.VMRangeFloatToString(
								3,
							),
							strings.VMRangeFloatToString(
								3,
							) + strings.VMRangeSeparator + strings.VMRangeFloatToString(
								5,
							),
							strings.VMRangeFloatToString(5) + strings.VMRangeSeparator + strings.VMRangeMax,
						},
					),
					Aggregation: model.Aggregation_AGGREGATION_HISTOGRAM,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				p := get(tt.name, model.Aggregation_AGGREGATION_HISTOGRAM, []option{initHistogram})
				p.SetBucket(func() *configs.Bucket { return configs.NewBucket(tt.args.buckets) })
				for _, v := range tt.args.values {
					updateHistogram(p, v)
				}
				changeHistogram(p, tt.args.factor)
				otp := model.NewNormalMetricsOTP()
				_, err := p.ToOTP(otp, 0)
				require.NoError(t, err)
				assert.Equal(t, tt.want, *otp)
			},
		)
	}
}
func TestRemoveEmptyBuckets(t *testing.T) {
	p := &Point{
		ranges: []string{"0-10", "10-20", "20-30", "30-40"},
		counts: []int64{0, 5, 0, 10},
	}

	expectedRanges := []string{"10-20", "30-40"}
	expectedCounts := []int64{5, 10}

	ranges, counts := p.getAndClearBucket()

	if len(ranges) != len(expectedRanges) || len(counts) != len(expectedCounts) {
		t.Fatalf(
			"Expected %d ranges and counts, got %d ranges and %d counts", len(expectedRanges), len(ranges), len(counts),
		)
	}

	for i := range ranges {
		if ranges[i] != expectedRanges[i] || counts[i] != expectedCounts[i] {
			t.Errorf(
				"Expected range %s with count %d, got range %s with count %d", expectedRanges[i], expectedCounts[i],
				ranges[i], counts[i],
			)
		}
	}
}
