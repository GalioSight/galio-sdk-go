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

package configs

import (
	"bytes"
	"sort"

	"galiosight.ai/galio-sdk-go/lib/strings"
)

// Bucket 递增直方图分桶。
type Bucket struct {
	// Key 分桶唯一键，用来检测分桶变化，由 Seconds 生成。
	Key string
	// Values 递增分桶值。
	Values []float64
}

// NewBucket 构造递增直方图分桶。
func NewBucket(values []float64) *Bucket {
	b := &Bucket{Values: values}
	b.sortAndSetKey()
	return b
}

// sortAndSetKey 桶排序且设置 key。
func (b *Bucket) sortAndSetKey() {
	sort.Float64s(b.Values) // 配置桶必须递增序，查询要二分。
	// 如果第一个桶大于0，则增加一个0桶，这样 histogram 的统计会更准确。
	if len(b.Values) > 0 && b.Values[0] > 0 {
		b.Values = append([]float64{0}, b.Values...)
	}
	var buffer bytes.Buffer
	for i := range b.Values {
		buffer.WriteString(strings.VMRangeFloatToString(b.Values[i]))
		buffer.WriteString("_")
	}
	b.Key = buffer.String()
}

// ConvBuckets 转换分桶配置到 map，方便使用。
func (m *Metrics) ConvBuckets() {
	m.HistogramBuckets = make(map[string]*Bucket, len(m.Processor.HistogramBuckets))
	for _, histogramBucket := range m.Processor.HistogramBuckets {
		m.HistogramBuckets[histogramBucket.Name] = NewBucket(histogramBucket.Buckets)
	}
}
