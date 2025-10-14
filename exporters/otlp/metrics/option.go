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

// Package metrics ...
package metrics

type options struct {
	schemaURL string
	apiKey    string
}

type option func(*options)

// WithSchemaURL 设置 schema URL
func WithSchemaURL(schema string) option {
	return func(o *options) {
		o.schemaURL = schema
	}
}

// WithAPIKey 设置 apiKey
func WithAPIKey(apiKey string) option {
	return func(o *options) {
		o.apiKey = apiKey
	}
}
