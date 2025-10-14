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

package strings

import (
	"testing"
)

func TestVMRangeFloatToString(t *testing.T) {
	tests := []struct {
		name string
		f    float64
		want string
	}{
		{
			name: "0.050000",
			f:    0.050000,
			want: "5.000e-02",
		},
		{
			name: "0.05000000074505806",
			f:    0.05000000074505806,
			want: "5.000e-02",
		},
		{
			name: "0.060000",
			f:    0.060000,
			want: "6.000e-02",
		},
		{
			name: "0.05999999865889549",
			f:    0.05999999865889549,
			want: "6.000e-02",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := VMRangeFloatToString(tt.f); got != tt.want {
					t.Errorf("VMRangeFloatToString() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
