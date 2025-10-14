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

package model

func (r *RPCLabels) grow(n int) {
	if cap(r.Fields) < n {
		r.Fields = make([]RPCLabels_Field, n)
	}
	r.Fields = r.Fields[:n]
}

// NewNormalLabels 构造常规标签，如：容器信息。
func NewNormalLabels() *NormalLabels {
	labels := &NormalLabels{
		Fields: make([]NormalLabels_Field, NormalLabels_max_field),
	}
	for i := range labels.Fields {
		labels.Fields[i].Name = NormalLabels_FieldName(i)
	}
	return labels
}

// ResourceToLabels 将 Resource 转成 NormalLabels。
func ResourceToLabels(r *Resource) *NormalLabels {
	normalLabels := NewNormalLabels()
	if r == nil {
		return normalLabels
	}
	normalLabels.Fields[NormalLabels_target].Value = r.Target
	normalLabels.Fields[NormalLabels_namespace].Value = r.Namespace
	normalLabels.Fields[NormalLabels_env_name].Value = r.EnvName
	normalLabels.Fields[NormalLabels_region].Value = r.Region
	normalLabels.Fields[NormalLabels_instance].Value = r.Instance
	normalLabels.Fields[NormalLabels_node].Value = r.Node
	normalLabels.Fields[NormalLabels_container_name].Value = r.ContainerName
	normalLabels.Fields[NormalLabels_version].Value = r.Version
	normalLabels.Fields[NormalLabels_city].Value = r.City
	normalLabels.Fields[NormalLabels_sdk_name].Value = r.SdkName
	normalLabels.Fields[NormalLabels_release_version].Value = r.ReleaseVersion
	return normalLabels
}
