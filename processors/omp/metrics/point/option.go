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
	"galiosight.ai/galio-sdk-go/model"
)

// option 数据点构造选项。
type option func(p *Point)

type config struct {
	options []option
}

var pointConfigs = newPointConfigs()

func newPointConfigs() []config {
	configs := make([]config, model.Aggregation_MAX_AGGREGATION)
	configs[model.Aggregation_AGGREGATION_SET] = config{
		options: []option{initSet},
	}
	configs[model.Aggregation_AGGREGATION_SUM] = config{
		options: []option{initSum},
	}
	configs[model.Aggregation_AGGREGATION_AVG] = config{
		options: []option{initAvg},
	}
	configs[model.Aggregation_AGGREGATION_MAX] = config{
		options: []option{initMax},
	}
	configs[model.Aggregation_AGGREGATION_MIN] = config{
		options: []option{initMin},
	}
	configs[model.Aggregation_AGGREGATION_HISTOGRAM] = config{
		options: []option{initHistogram},
	}
	configs[model.Aggregation_AGGREGATION_COUNTER] = config{
		options: []option{initCounter},
	}
	return configs
}

func getDefaultOptions(a model.Aggregation) []option {
	if a <= 0 || int(a) >= len(pointConfigs) {
		return nil
	}
	return pointConfigs[a].options
}
