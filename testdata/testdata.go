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

package testdata

import (
	"galiosight.ai/galio-sdk-go/model"
)

// Resource 用于测试的配置数据。
var Resource = model.Resource{
	Target:        "PCG-123.galileo.collector",
	Namespace:     "Production",
	EnvName:       "formal",
	Region:        "sh",
	Instance:      "9.143.128.197",
	Node:          "cls-b6je2c0t-3b00926fe615bc5580e0149d662d1788-155",
	ContainerName: "formal.galileo.collector.sh102312",
	SdkName:       "galileo",
	City:          "sh",
}

// NormalLabels 用于测试的标签集。
var NormalLabels = model.NormalLabels{
	Fields: []model.NormalLabels_Field{
		{
			Name:  model.NormalLabels_target,
			Value: "PCG-123.galileo.collector",
		},
		{
			Name:  model.NormalLabels_namespace,
			Value: "Production",
		},
		{
			Name:  model.NormalLabels_env_name,
			Value: "formal",
		},
		{
			Name:  model.NormalLabels_region,
			Value: "sh",
		},
		{
			Name:  model.NormalLabels_instance,
			Value: "9.143.128.197",
		},
		{
			Name:  model.NormalLabels_node,
			Value: "cls-b6je2c0t-3b00926fe615bc5580e0149d662d1788-155",
		},
		{
			Name:  model.NormalLabels_container_name,
			Value: "formal.galileo.collector.sh102312",
		},
		{
			Name:  model.NormalLabels_city,
			Value: "sh",
		},
	},
}
