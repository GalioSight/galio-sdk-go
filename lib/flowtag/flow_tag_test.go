// Copyright 2024 Tencent Galileo Authors
//
// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// template file: https://github.com/open-telemetry/opentelemetry-go/blob/main/semconv/template.j2

package flowtag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlowTagString(t *testing.T) {
	tests := []struct {
		tag      FlowTag
		expected string
	}{
		{0, ""},                                            // 0: no tags set
		{Gray, "Gray"},                                     // 1: Gray
		{Downgrade, "Downgrade"},                           // 2: Downgrade
		{Retry, "Retry"},                                   // 4: Retry
		{Gray | Downgrade, "Gray,Downgrade"},               // 3: Gray + Downgrade
		{Gray | Retry, "Gray,Retry"},                       // 5: Gray + Retry
		{Downgrade | Retry, "Downgrade,Retry"},             // 6: Downgrade + Retry
		{Gray | Downgrade | Retry, "Gray,Downgrade,Retry"}, // 7: Gray + Downgrade + Retry
		{8, ""}, // 8: no tags set
		{9, ""}, // 9: no tags set
	}

	for _, tt := range tests {
		t.Run(
			tt.expected, func(t *testing.T) {
				assert.Equal(t, tt.expected, tt.tag.String(), "Expected string representation to match")
			},
		)
	}
}
