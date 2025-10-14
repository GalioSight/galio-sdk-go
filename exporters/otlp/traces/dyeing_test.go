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
	"sync/atomic"
	"testing"

	"galiosight.ai/galio-sdk-go/lib/bloom"
	"galiosight.ai/galio-sdk-go/model"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func Test_dyeingSampler_UpdateConfigAndShouldSample(t *testing.T) {
	type fields struct {
		meta atomic.Value
	}
	type args struct {
		data []model.Dyeing
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			"nil",
			fields{},
			args{},
		},
		{
			"uin_1",
			fields{},
			args{
				[]model.Dyeing{
					{
						Key: "uin",
						Values: []string{
							"1",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				d := &dyeingSampler{
					dyeingRules: tt.fields.meta,
				}
				d.UpdateConfig(true, tt.args.data)
				meta, _ := d.dyeingRules.Load().(map[string]map[string]bool)
				assert.Equal(t, len(tt.args.data), len(meta))
				for _, v := range tt.args.data {
					for _, vv := range v.Values {
						assert.True(t, meta[v.Key][vv])
						parameters := sdktrace.SamplingParameters{
							Attributes: []attribute.KeyValue{
								attribute.Key(v.Key).String(vv),
							},
						}
						assert.Equal(t, true, d.ShouldSample(&parameters))
					}
				}
				assert.Equal(
					t, false,
					d.ShouldSample(
						&sdktrace.SamplingParameters{
							Attributes: []attribute.KeyValue{attribute.Key("a").String("b")},
						},
					),
				)
			},
		)
	}
}

func Test_bloomDyeingSampler_UpdateConfigAndShouldSample(t *testing.T) {
	type fields struct {
		meta atomic.Value
	}
	type args struct {
		dyeingKey    string
		dyeingValues []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			"nil",
			fields{},
			args{},
		},
		{
			"uin_1",
			fields{},
			args{
				dyeingKey: "uin",
				dyeingValues: []string{
					"aaa",
					"bbb",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				d := &bloomDyeingSampler{
					bloomDyeingRules: tt.fields.meta,
				}
				m, k := bloom.EstimateParameters(int32(len(tt.args.dyeingValues)), 0.0001)
				bloomFilter := bloom.New(m, k)
				for _, value := range tt.args.dyeingValues {
					bloomFilter.Add(value)
				}
				d.UpdateConfig(
					true, []model.BloomDyeing{
						{
							Key:        tt.args.dyeingKey,
							BitSize:    bloomFilter.Cap(),
							HashNumber: bloomFilter.K(),
							Bitmap:     bloomFilter.Bitmap(),
						},
					},
				)
				for _, v := range tt.args.dyeingValues {
					parameters := sdktrace.SamplingParameters{
						Attributes: []attribute.KeyValue{
							attribute.Key(tt.args.dyeingKey).String(v),
						},
					}
					assert.Equal(t, true, d.ShouldSample(&parameters))
				}
				assert.Equal(
					t, false,
					d.ShouldSample(
						&sdktrace.SamplingParameters{
							Attributes: []attribute.KeyValue{attribute.Key("a").String("b")},
						},
					),
				)
			},
		)
	}
}

func TestShouldSample(t *testing.T) {
	sampler := &dyeingSampler{}
	sampler.dyeingRules.Store(make(map[string]map[string]bool))

	// 创建一个包含一个属性的 SamplingParameters
	params := &sdktrace.SamplingParameters{
		Attributes: []attribute.KeyValue{
			attribute.String("nonexistentKey", "value"),
		},
	}

	// 调用 ShouldSample，期望返回 false 而不是引发 panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("The code panicked: %v", r)
		}
	}()

	result := sampler.ShouldSample(params)
	if result {
		t.Errorf("ShouldSample returned true, expected false")
	}
}

func TestMapNil(t *testing.T) {
	nilString := string([]byte(nil))
	nilString = string([]byte{0, 1, 2})
	var x map[string]map[string]bool = nil
	assert.Equal(t, false, x["abcd"]["xyz"])
	assert.Equal(t, false, x[""][""])
	assert.Equal(t, false, x[nilString][nilString])
	var y map[string]bool = nil
	assert.Equal(t, false, y["abcd"])
	assert.Equal(t, false, y[""])
	assert.Equal(t, false, y[nilString])
	var z map[string]bool = map[string]bool{
		"": false,
	}
	assert.Equal(t, false, z["abcd"])
	assert.Equal(t, false, z[""])
	assert.Equal(t, false, z[nilString])

}
