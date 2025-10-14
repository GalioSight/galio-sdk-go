// Copyright 2024 Tencent Galileo Authors

// Package attribute ...
package attribute

import (
	"errors"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel/trace"
)

type affinity struct {
	target string
}

var Affinity affinity

// ParsedAffinity 解析 Affinity 字符串，详见 Parse 函数
type ParsedAffinity struct {
	Host, Peer string
	Kind       trace.SpanKind
}

const sep = "│" // 制表符的竖线，不是|，键盘打不出来的

// Parse 提供尽力解析
// 如果能解析出来 kind，根据 kind 能确定具体主服务是 caller 这边还是 callee 这边
// 格式：
// host-peer-3
// peer-host-2
func (a *affinity) Parse(v string) (ParsedAffinity, error) {
	left, v, _ := strings.Cut(v, sep)
	right, v, _ := strings.Cut(v, sep)
	var kind trace.SpanKind
	if n, err := strconv.Atoi(v); err != nil {
		return ParsedAffinity{}, err
	} else {
		kind = trace.SpanKind(n)
	}
	ret := ParsedAffinity{Kind: kind}
	switch kind {
	case trace.SpanKindServer, trace.SpanKindConsumer:
		ret.Host, ret.Peer = right, left
	case trace.SpanKindClient, trace.SpanKindProducer:
		ret.Host, ret.Peer = left, right
	default:
		return ret, errors.New("unexpected span kind")
	}
	return ret, nil
}

func (a *affinity) String(keys *RPCKeys, kind trace.SpanKind) string {
	skind := strconv.Itoa(int(kind))
	switch kind {
	case trace.SpanKindServer, trace.SpanKindConsumer:
		return keys.CallerService + sep + a.target + sep + skind
	case trace.SpanKindClient, trace.SpanKindProducer:
		return a.target + sep + keys.CalleeService + sep + skind
	default:
		return keys.CallerService + sep + keys.CalleeService + sep + skind
	}
}

func (a *affinity) SetTarget(t string) {
	a.target = t
}
