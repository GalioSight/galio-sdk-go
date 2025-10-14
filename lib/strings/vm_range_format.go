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
	"fmt"
)

// VMRangeFloatToString VictoriaMetrics range 分桶浮点数转字符串
// https://github.com/VictoriaMetrics/metrics/blob/master/histogram.go#L186
func VMRangeFloatToString(f float64) string {
	return fmt.Sprintf("%.3e", f)
}

const (
	// VMRangeSeparator vm range 分隔符，开始...结束。
	VMRangeSeparator = "..."
	// VMRangeMax vm range 最大。
	VMRangeMax = "+Inf"
	// VMRangeMin vm range 最小。
	VMRangeMin = "-Inf"
)
