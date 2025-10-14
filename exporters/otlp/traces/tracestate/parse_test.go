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

// Package main ...
package tracestate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func match(w int, s Strategy) func(s *State) bool {
	return func(st *State) bool {
		return st.Workflow.Result == WorkflowResult(w) && st.Sample.SampledStrategy == s
	}
}

func path(p, pp string) func(s *State) bool {
	return func(st *State) bool {
		return st.Workflow.Path == p && st.Workflow.ParentPath == pp
	}
}

func match2(r Strategy) func(s *State) bool {
	return func(st *State) bool {
		return st.Sample.RootStrategy == r
	}
}

func TestParse(t *testing.T) {
	const same = "<same>"
	m := match
	m2 := match2
	p := path
	tests := []struct {
		ts     string
		format string
		check  func(s *State) bool
	}{
		{"", "", m(0, -1)},
		{"w;s", "w:2;s", m(2, 1)},
		{"w:;s:;", "w:2;s", m(2, 1)},
		{"w:alpha;s:2:a", "w:2;s:2:a", m(2, StrategyDyeing)},
		{"w:alpha;s:bad", "w:2;s", m(2, StrategyMatch)},
		{"w;u;s", "u;w:2;s", m(2, 1)},
		{"w;u:a1;u2:a2;s;u3:a3;", "u:a1;u2:a2;u3:a3;w:2;s", m(2, 1)},
		{"w:1:abcd:2:3:4", same, p("abcd", "2")},
		{"w:3", "w:3", m(3, 0)},
		{"w:2", "w:2", m(2, 0)},
		{"w:1", "w:1", m(1, 0)},
		{"w:4", "w:4", m(4, 0)},
		{"s;r", "s;r", m2(1)},
		{"", "", m2(-1)},
		{"r:3", same, m2(StrategyMinCount)},
	}

	a := assert.New(t)
	for _, test := range tests {
		t.Run(
			"", func(t *testing.T) {
				s, _ := Parse(test.ts)
				if test.format == same {
					a.Equal(test.ts, s.String())
				} else {
					a.Equal(test.format, s.String(), "%#v", s)
				}
				a.True(test.check(&s))
			},
		)
	}
}

func TestModify(t *testing.T) {
	m := func(w int, s Strategy) func(p *State) {
		return func(p *State) {
			if w == 1 {
				p.Workflow.Result = WorkflowSample
			} else {
				p.Workflow.Result = WorkflowDrop
			}
			p.Sample.SampledStrategy = s
		}
	}
	p := func(w int, path string) func(p *State) {
		return func(p *State) {
			p.Workflow.Result = WorkflowResult(w)
			p.Workflow.Path = path
		}
	}
	r := func(r int) func(p *State) {
		return func(p *State) {
			p.Sample.RootStrategy = Strategy(r)
		}
	}
	tests := []struct {
		ts  string
		e   string
		mod func(p *State)
	}{
		{"", "w:2;s:2", m(1, StrategyDyeing)},
		{"w;s:2", "w:1;s", m(0, StrategyMatch)},
		{"w:alpha;s:1:a:b", "w:2;s:2:a:b", m(1, StrategyDyeing)},
		{"w:alpha;s:1:a:b", "w:2:abcd;s:1:a:b", p(2, "abcd")},
		{"w:alpha;s:1:a:b", "s:1:a:b", p(0, "abcd")},
		{"w;s:2", "w:2;s:2;r", r(1)},
		{"r:1:2", "r:3:2", r(3)},
	}

	a := assert.New(t)
	for _, test := range tests {
		t.Run(
			"", func(t *testing.T) {
				s, e := Parse(test.ts)
				a.NoError(e)
				test.mod(&s)
				a.Equal(test.e, s.String())
			},
		)
	}
}

func TestSampled(t *testing.T) {
	tests := []struct {
		ts    string
		match bool
	}{
		{"", false},
		{"w:0", false},
		{"w:1", false},
		{"w", true},
		{"w:2", true},
		{"w:3", true},
	}

	a := assert.New(t)
	var niled *WorkflowState
	a.False(niled.Sampled())
	for _, test := range tests {
		t.Run(
			"", func(t *testing.T) {
				s, _ := Parse(test.ts)
				a.Equal(test.match, s.Workflow.Sampled())
			},
		)
	}
}
