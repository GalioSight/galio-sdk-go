// Copyright 2024 Tencent Galileo Authors

// Package attribute ...
package attribute

import (
	"go.opentelemetry.io/otel/attribute"

	semconv "galiosight.ai/galio-sdk-go/semconv/v1.0.0"
)

const iNotFound = -1

// idxfn 输入 key，输出写入目标数组的下标
// takeKey 快速获取 attribute 指定字段
func takeKey(to []*string, idxfn func(key attribute.Key) int, kv []attribute.KeyValue) {
	n := 0
	for i := range kv {
		if x := idxfn(kv[i].Key); x > iNotFound {
			*to[x] = kv[i].Value.AsString()
			n++
			if n == len(to) {
				break
			}
		}
	}
}

// SDK 上报只上报了 service，在 collector 解出来的 server
var keyCmp = []attribute.Key{
	semconv.TrpcCalleeMethodKey, semconv.TrpcCalleeServiceKey,
	semconv.TrpcCallerMethodKey, semconv.TrpcCallerServiceKey,
}

const iCallex = len("trpc.caller") - 1    // 直接取 calle[er]_method 处的字符
const iCallerY = len("trpc.caller_m") - 1 // 直接取 caller_[m]ethod, caller_[s]erver 处的字符

const (
	calleeMethodKey int = iota
	calleeServiceKey
	callerMethodKey
	callerServiceKey
)

// RPCKeys 获取 rpc 字段
type RPCKeys struct {
	CalleeMethod  string
	CalleeService string
	CallerMethod  string
	CallerService string
}

// NewRPCKeys ...
func NewRPCKeys(kv []attribute.KeyValue) RPCKeys {
	var r RPCKeys
	to := [4]*string{
		&r.CalleeMethod, &r.CalleeService, &r.CallerMethod, &r.CallerService,
	}
	takeKey(to[:], rpcIndex, kv)
	return r
}

// RPCIndex 返回 caller，callee 字段下标
func rpcIndex(key attribute.Key) int {
	i := 0
	switch len(key) {
	case len(semconv.TrpcCallerMethodKey):
		if key[iCallerY] != 'm' {
			return iNotFound
		}
		switch key[iCallex] {
		case 'e':
			i = calleeMethodKey
		case 'r':
			i = callerMethodKey
		default:
			return iNotFound
		}
	case len(semconv.TrpcCallerServiceKey):
		if key[iCallerY] != 's' {
			return iNotFound
		}
		switch key[iCallex] {
		case 'e':
			i = calleeServiceKey
		case 'r':
			i = callerServiceKey
		default:
			return iNotFound
		}
	}
	if key == keyCmp[i] {
		return int(i)
	}
	return iNotFound
}

// CalleeKeys ...
type CalleeKeys struct {
	CalleeMethod  string
	CalleeService string
}

// NewCalleeKeys ...
func NewCalleeKeys(kv []attribute.KeyValue) CalleeKeys {
	var k CalleeKeys
	to := [2]*string{
		&k.CalleeMethod, &k.CalleeService,
	}
	takeKey(to[:], calleeIndex, kv)
	return k
}

// CalleeIndex 返回被调下标
func calleeIndex(key attribute.Key) int {
	i := 0
	switch len(key) {
	case len(semconv.TrpcCallerMethodKey):
		if key[iCallerY] != 'm' {
			return iNotFound
		}
		switch key[iCallex] {
		case 'e':
			i = calleeMethodKey
		default:
			return iNotFound
		}
	case len(semconv.TrpcCallerServiceKey):
		if key[iCallerY] != 's' {
			return iNotFound
		}
		switch key[iCallex] {
		case 'e':
			i = calleeServiceKey
		default:
			return iNotFound
		}
	}
	if key == keyCmp[i] {
		return int(i)
	}
	return iNotFound
}
