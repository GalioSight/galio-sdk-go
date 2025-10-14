// Copyright 2025 Tencent Galileo Authors
//
// Copyright 2025 Tencent OpenTelemetry Oteam
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

// Package model ...
package model

import (
	modelv1 "galiosight.ai/galio-sdk-go/model"
)

// RPCLabelsField 将 v3 的 labels 转换成 v1 的 labels
func RPCLabelsField(name RPCLabels_FieldName, value string) modelv1.RPCLabels_Field {
	return modelv1.RPCLabels_Field{Name: modelv1.RPCLabels_FieldName(name), Value: value}
}
