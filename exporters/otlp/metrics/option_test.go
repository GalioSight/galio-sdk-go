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

package metrics

import (
	"testing"
)

func TestOptions(t *testing.T) {
	tests := []struct {
		name     string
		opts     []option
		validate func(*options)
	}{
		{
			name: "default empty options",
			opts: nil,
			validate: func(o *options) {
				if o.schemaURL != "" || o.apiKey != "" {
					t.Error("Expected empty options when no opts applied")
				}
			},
		},
		{
			name: "with schema only",
			opts: []option{WithSchemaURL("https://schema.example")},
			validate: func(o *options) {
				if o.schemaURL != "https://schema.example" {
					t.Errorf("Unexpected schemaURL: %s", o.schemaURL)
				}
				if o.apiKey != "" {
					t.Error("apiKey should be empty")
				}
			},
		},
		{
			name: "with api key only",
			opts: []option{WithAPIKey("secret-key")},
			validate: func(o *options) {
				if o.apiKey != "secret-key" {
					t.Errorf("Unexpected apiKey: %s", o.apiKey)
				}
				if o.schemaURL != "" {
					t.Error("schemaURL should be empty")
				}
			},
		},
		{
			name: "multiple options",
			opts: []option{
				WithSchemaURL("https://prod.schema"),
				WithAPIKey("prod-key"),
			},
			validate: func(o *options) {
				if o.schemaURL != "https://prod.schema" {
					t.Errorf("Schema mismatch: %s", o.schemaURL)
				}
				if o.apiKey != "prod-key" {
					t.Errorf("API key mismatch: %s", o.apiKey)
				}
			},
		},
		{
			name: "empty string values",
			opts: []option{
				WithSchemaURL(""),
				WithAPIKey(""),
			},
			validate: func(o *options) {
				if o.schemaURL != "" {
					t.Error("schemaURL should accept empty string")
				}
				if o.apiKey != "" {
					t.Error("apiKey should accept empty string")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				o := &options{}
				for _, opt := range tt.opts {
					opt(o)
				}
				tt.validate(o)
			},
		)
	}
}
