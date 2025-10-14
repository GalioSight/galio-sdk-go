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

// Package internal 直接操控底层的 "go.opentelemetry.io/otel/trace.TraceState"，性能更好
package internal

import (
	"fmt"
	"strings"
	"unsafe"

	"go.opentelemetry.io/otel/trace"
)

type member struct {
	Key   string
	Value string
}

type traceState struct {
	list []member
}

// Convert to internal.traceState
func Convert(ts *trace.TraceState) *traceState {
	return (*traceState)(unsafe.Pointer(ts))
}

const (
	maxListMembers  = 32
	listDelimiter   = ','
	listDelimiters  = ","
	memberDelimiter = "="

	errDuplicate     errorConst = "duplicate list-member in tracestate"
	errMemberNumber  errorConst = "too many list-members in tracestate"
	errInvalidMember errorConst = "invalid tracestate list-member"
)

type errorConst string

func (e errorConst) Error() string {
	return string(e)
}

func checkValueChar(v byte) bool {
	return v >= '\x20' && v <= '\x7e' && v != '\x2c' && v != '\x3d'
}

func checkValueLast(v byte) bool {
	return v >= '\x21' && v <= '\x7e' && v != '\x2c' && v != '\x3d'
}

func checkValue(val string) bool {
	n := len(val)
	if n == 0 || n > 256 {
		return false
	}
	// valueFormat         = `[\x20-\x2b\x2d-\x3c\x3e-\x7e]{0,255}[\x21-\x2b\x2d-\x3c\x3e-\x7e]`
	for i := 0; i < n-1; i++ {
		if !checkValueChar(val[i]) {
			return false
		}
	}
	return checkValueLast(val[n-1])
}

func checkKeyLeft(key string) bool {
	for _, v := range key {
		if (v >= '0' && v <= '9') || (v >= 'a' && v <= 'z') {
			continue
		}
		switch v {
		case '_', '-', '*', '/':
			continue
		}
		return false
	}
	return true
}

func checkKeyPart(key string, n int, tenant bool) bool {
	if len(key) == 0 {
		return false
	}
	a := key[0]
	ret := len(key[1:]) <= n
	if tenant {
		ret = ret && ((a >= 'a' && a <= 'z') || (a >= '0' && a <= '9'))
	} else {
		ret = ret && a >= 'a' && a <= 'z'
	}
	return ret && checkKeyLeft(key[1:])
}

func checkKey(key string) bool {
	// noTenantKeyFormat   = `[a-z][_0-9a-z\-\*\/]{0,255}`
	// withTenantKeyFormat = `[a-z0-9][_0-9a-z\-\*\/]{0,240}@[a-z][_0-9a-z\-\*\/]{0,13}`
	tenant, system, ok := strings.Cut(key, "@")
	if !ok {
		return checkKeyPart(key, 255, false)
	}
	return checkKeyPart(tenant, 240, true) && checkKeyPart(system, 13, false)
}

func parseMember(m string) (member, error) {
	key, val, ok := strings.Cut(m, memberDelimiter)
	bad := member{}
	if !ok {
		return bad, fmt.Errorf("%w: %s", errInvalidMember, m)
	}
	key = strings.TrimLeft(key, " \t")
	if !checkKey(key) {
		return bad, fmt.Errorf("%w: %s", errInvalidMember, m)
	}
	val = strings.TrimRight(val, " \t")
	if !checkValue(val) {
		return bad, fmt.Errorf("%w: %s", errInvalidMember, m)
	}
	return member{Key: key, Value: val}, nil
}

// ParseTraceState a really quick version
func ParseTraceState(ts string) (traceState, error) {
	bad := traceState{}
	wrapErr := func(err error) error {
		return fmt.Errorf("failed to parse tracestate: %w", err)
	}
	found := make(map[string]struct{})
	var members []member
	for ts != "" {
		var member string
		member, ts, _ = strings.Cut(ts, listDelimiters)
		if member == "" {
			continue
		}
		m, err := parseMember(member)
		if err != nil {
			return bad, wrapErr(err)
		}
		if _, ok := found[m.Key]; ok {
			return bad, wrapErr(errDuplicate)
		}
		found[m.Key] = struct{}{}

		members = append(members, m)
		if n := len(members); n > maxListMembers {
			return bad, wrapErr(errMemberNumber)
		}
	}
	return traceState{list: members}, nil
}

// Insert a really quick version
func (ts *traceState) Insert(key, value string) traceState {
	m := member{key, value} // just trust string, because internal.TraceState can only used in SDK code
	n := len(ts.list)
	found := n
	for i := range ts.list {
		if ts.list[i].Key == key {
			found = i
		}
	}
	ret := traceState{}
	if found == n && n < maxListMembers {
		ret.list = make([]member, n+1)
	} else {
		ret.list = make([]member, n)
	}
	ret.list[0] = m
	copy(ret.list[1:], ts.list[0:found])
	if found < n {
		copy(ret.list[1+found:], ts.list[found+1:])
	}
	return ret
}

// String a really quick version
func (ts *traceState) String() string {
	if len(ts.list) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(ts.list[0].Key)
	sb.WriteByte('=')
	sb.WriteString(ts.list[0].Value)
	for i := 1; i < len(ts.list); i++ {
		sb.WriteByte(listDelimiter)
		sb.WriteString(ts.list[i].Key)
		sb.WriteByte('=')
		sb.WriteString(ts.list[i].Value)
	}
	return sb.String()
}

// Convert back to trace.TraceState
func (ts *traceState) Convert() *trace.TraceState {
	return (*trace.TraceState)(unsafe.Pointer(ts))
}
