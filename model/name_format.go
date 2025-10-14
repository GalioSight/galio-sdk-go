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

import (
	"bytes"
	"strings"
	"sync"
	"unicode"

	"galiosight.ai/galio-sdk-go/errs"
	libstrings "galiosight.ai/galio-sdk-go/lib/strings"
	"github.com/jxskiss/base62"
)

// Base62Prefix base62 编码的名字，以 base62 开头，避免与用户名字混淆。
// 注意没有下划线，因为下划线在这里并不能带来额外的信息。
// 而下划线在我们名字中有分割符的作用，下划线太多，会让名字解析变得复杂。
const Base62Prefix = "base62"

// CustomName 生成自定义监控项名字。
// group: 组名
// name: 指标名
// aggregation: 数据聚合类型
// check: 是否检查名字的合法性
// 如果 group、usage 非空，则完整指标名为：custom_$type_$group_$name_$usage。
// 此指标名可用于将许多指标在视图上做分组展示。
func CustomName(group, name string, aggregation Aggregation) string {
	return GetMetricSchema(group, name, aggregation).MetricName
}

var metricSchemas = struct {
	cache map[MetricKey]*MetricSchema
	mu    sync.RWMutex
}{
	cache: map[MetricKey]*MetricSchema{},
}

// GetMetricSchema 获取元数据对象，当 cache 中没有时，创建对象并放入 cache。
func GetMetricSchema(group, name string, aggregation Aggregation) *MetricSchema {
	metricKey := MetricKey{group, name, aggregation}
	metricSchemas.mu.RLock()
	n, ok := metricSchemas.cache[metricKey]
	metricSchemas.mu.RUnlock()
	if ok {
		return n
	}
	metricSchemas.mu.Lock()
	defer metricSchemas.mu.Unlock()
	n, ok = metricSchemas.cache[metricKey]
	if ok {
		return n
	}
	n = NewMetricSchema(group, name, aggregation)
	metricSchemas.cache[metricKey] = n
	return n
}

// NewMetricSchema 创建 MetricSchema 对象
func NewMetricSchema(group, name string, aggregation Aggregation) *MetricSchema {
	if group == "" {
		group = "default"
	}
	c := &MetricSchema{
		MonitorName: group,
		MetricAlias: name,
	}
	c.MetricType, c.Usage = AggregationToTypeUsage(aggregation)
	c.Aggregation = aggregation.String()
	c.MetricName = c.fullMetricName()
	return c
}

func (m *MetricSchema) fullMetricName() string {
	var b bytes.Buffer
	const prefix = "custom_"
	monitorID := NameToIdentifier(m.MonitorName)
	metricID := NameToIdentifier(m.MetricAlias)
	b.Grow(len(prefix) + len(monitorID) + len(metricID) + len(m.Usage) + 3)
	b.WriteString(prefix)
	b.WriteString(m.MetricType)
	b.WriteByte('_')
	b.WriteString(monitorID)
	b.WriteByte('_')
	b.WriteString(metricID)
	if m.Usage != "" {
		b.WriteByte('_')
		b.WriteString(m.Usage)
	}
	return b.String()
}

// ParseCustomName 解释自定义监控项名，得到元数据。
// customName 是符合 OMP 规范的自定义监控项名。
// 此接口未来会移除，使用元数据接口代替。
func ParseCustomName(customName string) (MetricSchema, error) {
	if !strings.HasPrefix(customName, "custom_") {
		return MetricSchema{}, errs.ErrCustomName
	}
	parts := strings.Split(customName, "_")
	n := len(parts)
	// custom_$type_$group_$name_$usage, 一般是 5 个部分，group 为空时，是 4 个部分。
	// custom_$type_$group_$name_$usage, 一般是 5 个部分，group 为空时，是 4 个部分。
	const minParts = 4
	if n < minParts {
		return MetricSchema{}, errs.ErrCustomName
	}
	group, name := getGroupAndName(parts[2 : n-1])
	return MetricSchema{
		MetricType:  parts[1],
		MonitorName: IdentifierToName(group),
		MetricName:  IdentifierToName(name),
		Aggregation: parts[n-1],
	}, nil
}

// getGroupAndName，解析 $group_$name 部分，group 有可能为空。
// 由于 group 允许为空，不是唯一的。所以解析出来可能会有多种情况。
// 比如 "a_b"，有可能是 "a", "b"拼接出来，也有可能是 "","a_b"拼接出来，也可能是"a_b",""拼接出来的。
// 我们选择用第一种结果。
// 但如果其中有一个是 base62，则可以确定。
// 比如 "a_base62xxxx"，可以确定是 "a", "base62xxxx"拼接出来的。
// 所以 group 允许为空是一种不好的设计，在新版本中，我们会默认给它一个 group。
// 但是对于 "a_b_c_d" 这种，还是难以判定，所以需要有元数据来专门记录。
func getGroupAndName(parts []string) (string, string) {
	n := len(parts)
	if n == 1 {
		return "", parts[0]
	}
	var group, name string
	if IsIdentifier(parts[n-1]) {
		group = IdentifierToName(strings.Join(parts[0:n-1], "_"))
		name = IdentifierToName(parts[n-1])
	} else {
		group = IdentifierToName(parts[0])
		name = IdentifierToName(strings.Join(parts[1:n], "_"))
	}
	return group, name
}

// ValidName 检查标签名字是否合法。
// 注意指标名和标签名必须符合正则表达式： "[a-zA-Z0-9_]*"，只能有字母、数字和下划线。
// name 允许为空。
func ValidName(name string) bool {
	for _, c := range name {
		if !ValidRune(c) {
			return false
		}
	}
	return true
}

// NameToIdentifier 将指标名、标签名转成符合 prometheus 命名规则的标识符。
// 因为 prometheus 合法的名字必须符合正则表达式： "[a-zA-Z0-9_]*"，只能有字母、数字和下划线。
// 但是如果业务上报的名字，含有非法字符（比如中文和特殊符号），我们就需要转换成合法的名字，用于存储。
// 但是在展示页面上，还希望继续保留原始的名字，方便用户查看。
// 我们选择的方式是，将 name 用 base62 编码，在展示的时候，进行 base62 解码，这样就可以兼容存储、查询、展示。
// base62 编码的名字，以 base62 开头，避免与用户名字混淆。
// 对应的解码函数：IdentifierToName 可将编码后的名字解码回来。
func NameToIdentifier(name string) string {
	if ValidName(name) {
		return name
	}
	return Base62Prefix + base62.EncodeToString(libstrings.NoAllocBytes(name))
}

// IdentifierToName  将标识符转换成原始名字。
// 标识符是由 NameToIdentifier 函数生成的。
func IdentifierToName(name string) string {
	if IsIdentifier(name) {
		decode, _ := base62.Decode(libstrings.NoAllocBytes(name)[len(Base62Prefix):])
		return libstrings.NoAllocString(decode)
	}
	return name
}

// IsIdentifier 判断该名字是否经 base62 编码后的 id。
func IsIdentifier(name string) bool {
	return strings.HasPrefix(name, Base62Prefix)
}

// ToValidName 将名字转成合法的名字。
// 注意指标名和标签名必须符合正则表达式： "[a-zA-Z0-9_]*"，只能有字母、数字和下划线。
// 如果名字不合法，则只保留合法字符，其他字符用_代替。
// 因为指标名其实类似于数据库的列名，后面的查询中是要直接使用这个名字的。
// 如果它有括号等字符，会直接破坏 sql/promql 的语义，极大地增加了复杂性。
// 所以 Prometheus 指标规范里面，要求指标名必须是一个标识符。
// 与 go 里面的变量名要求一致。
// 由于自定义指标会在前面加上 custom_ 等前缀，所以此处允许 name 为空。
// 参考 CustomName 实现。
func ToValidName(name string) string {
	n := strings.Builder{}
	for _, c := range name {
		if ValidRune(c) {
			n.WriteRune(c)
		} else {
			n.WriteRune('_')
		}
	}
	return n.String()
}

// ValidRune 标签名里面的合法字符，只有 "[a-zA-Z0-9_]"，即只能有字母、数字和下划线。
// 非 ASCII，返回 false，否则 IsDigit 等不符合预期。
func ValidRune(r int32) bool {
	if r > unicode.MaxASCII {
		return false
	}
	return r == '_' || unicode.IsDigit(r) || unicode.IsLetter(r)
}

// AggregationToTypeUsage 用法：total/sum/count/min/max/set/""
// 规范文档：
// https://galiosight.ai/eco/blob/master/proto/omp.yaml#L358
// histogram 和 avg，落地时会拆成多个指标，由 otp 来追加的，所以返回的 usage 是空。
// 其中，histogram 会转成 _sum,_count,_bucket 三个指标。
// avg 会转成 _sum, _count.
func AggregationToTypeUsage(m Aggregation) (string, string) {
	const (
		gauge     = "gauge"
		counter   = "counter"
		histogram = "histogram"

		set   = "set"
		total = "total"
		max   = "max"
		min   = "min"
	)
	switch m {
	case Aggregation_AGGREGATION_SET:
		return gauge, set
	case Aggregation_AGGREGATION_SUM, Aggregation_AGGREGATION_COUNTER:
		return counter, total
	case Aggregation_AGGREGATION_AVG:
		return counter, ""
	case Aggregation_AGGREGATION_MAX:
		return gauge, max
	case Aggregation_AGGREGATION_MIN:
		return gauge, min
	case Aggregation_AGGREGATION_HISTOGRAM:
		return histogram, ""
	default:
		return gauge, set
	}
}
