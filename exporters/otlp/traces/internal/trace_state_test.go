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

// Package internal ...
package internal

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

func n32() string {
	var lst []string
	for i := 0; i < 32; i++ {
		lst = append(lst, fmt.Sprintf("a%d=1", i))
	}
	return strings.Join(lst, ",")
}

func TestInsert(t *testing.T) {
	tests := []struct {
		ts         string
		key, value string
		want       string
	}{
		{"", "g", "1", "g=1"},
		{"g=1", "g", "2", "g=2"},
		{"ot=1", "g", "2", "g=2,ot=1"},
		{"ot=1,g=1", "g", "2", "g=2,ot=1"},
		{
			n32(), "a32", "2",
			"a32=2,a0=1,a1=1,a2=1,a3=1,a4=1,a5=1,a6=1,a7=1,a8=1,a9=1,a10=1,a11=1,a12=1,a13=1,a14=1,a15=1,a16=1,a17=1,a18=1,a19=1,a20=1,a21=1,a22=1,a23=1,a24=1,a25=1,a26=1,a27=1,a28=1,a29=1,a30=1",
		},
		{
			n32(), "a31", "2",
			"a31=2,a0=1,a1=1,a2=1,a3=1,a4=1,a5=1,a6=1,a7=1,a8=1,a9=1,a10=1,a11=1,a12=1,a13=1,a14=1,a15=1,a16=1,a17=1,a18=1,a19=1,a20=1,a21=1,a22=1,a23=1,a24=1,a25=1,a26=1,a27=1,a28=1,a29=1,a30=1",
		},
	}

	a := assert.New(t)
	for _, test := range tests {
		t.Run(
			"", func(t *testing.T) {
				ts, err := trace.ParseTraceState(test.ts)
				a.NoError(err)
				ts2 := Convert(&ts).Insert(test.key, test.value)
				a.Equal(test.want, ts2.Convert().String())
				a.Equal(test.want, ts2.String())
			},
		)
	}
}

type TraceState = traceState

// Taken from the W3C tests:
// https://github.com/w3c/trace-context/blob/dcd3ad9b7d6ac36f70ff3739874b73c11b0302a1/test/test_data.json
var testcases = []struct {
	name       string
	in         string
	tracestate TraceState
	out        string
	err        error
}{
	{
		name: "duplicate with the same value",
		in:   "foo=1,foo=1",
		err:  errDuplicate,
	},
	{
		name: "duplicate with different values",
		in:   "foo=1,foo=2",
		err:  errDuplicate,
	},
	{
		name: "improperly formatted key/value pair",
		in:   "foo =1",
		err:  errInvalidMember,
	},
	{
		name: "upper case key",
		in:   "FOO=1",
		err:  errInvalidMember,
	},
	{
		name: "key with invalid character",
		in:   "foo.bar=1",
		err:  errInvalidMember,
	},
	{
		name: "multiple keys, one with empty tenant key",
		in:   "foo@=1,bar=2",
		err:  errInvalidMember,
	},
	{
		name: "multiple keys, one with only tenant",
		in:   "@foo=1,bar=2",
		err:  errInvalidMember,
	},
	{
		name: "multiple keys, one with double tenant separator",
		in:   "foo@@bar=1,bar=2",
		err:  errInvalidMember,
	},
	{
		name: "multiple keys, one with multiple tenants",
		in:   "foo@bar@baz=1,bar=2",
		err:  errInvalidMember,
	},
	{
		name: "key too long",
		in:   "foo=1,zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz=1",
		err:  errInvalidMember,
	},
	{
		name: "key too long, with tenant",
		in:   "foo=1,tttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttt@v=1",
		err:  errInvalidMember,
	},
	{
		name: "tenant too long",
		in:   "foo=1,t@vvvvvvvvvvvvvvv=1",
		err:  errInvalidMember,
	},
	{
		name: "multiple values for a single key",
		in:   "foo=bar=baz",
		err:  errInvalidMember,
	},
	{
		name: "no value",
		in:   "foo=,bar=3",
		err:  errInvalidMember,
	},
	{
		name: "too many members",
		in:   "bar01=01,bar02=02,bar03=03,bar04=04,bar05=05,bar06=06,bar07=07,bar08=08,bar09=09,bar10=10,bar11=11,bar12=12,bar13=13,bar14=14,bar15=15,bar16=16,bar17=17,bar18=18,bar19=19,bar20=20,bar21=21,bar22=22,bar23=23,bar24=24,bar25=25,bar26=26,bar27=27,bar28=28,bar29=29,bar30=30,bar31=31,bar32=32,bar33=33",
		err:  errMemberNumber,
	},
	{
		name: "valid key/value list",
		in:   "abcdefghijklmnopqrstuvwxyz0123456789_-*/= !\"#$%&'()*+-./0123456789:;<>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~",
		out:  "abcdefghijklmnopqrstuvwxyz0123456789_-*/= !\"#$%&'()*+-./0123456789:;<>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~",
		tracestate: TraceState{
			list: []member{
				{
					Key:   "abcdefghijklmnopqrstuvwxyz0123456789_-*/",
					Value: " !\"#$%&'()*+-./0123456789:;<>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~",
				},
			},
		},
	},
	{
		name: "valid key/value list with tenant",
		in:   "abcdefghijklmnopqrstuvwxyz0123456789_-*/@a-z0-9_-*/= !\"#$%&'()*+-./0123456789:;<>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~",
		out:  "abcdefghijklmnopqrstuvwxyz0123456789_-*/@a-z0-9_-*/= !\"#$%&'()*+-./0123456789:;<>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~",
		tracestate: TraceState{
			list: []member{
				{
					Key:   "abcdefghijklmnopqrstuvwxyz0123456789_-*/@a-z0-9_-*/",
					Value: " !\"#$%&'()*+-./0123456789:;<>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~",
				},
			},
		},
	},
	{
		name: "empty input",
		// Empty input should result in no error and a zero value
		// TraceState being returned, that TraceState should be encoded as an
		// empty string.
	},
	{
		name: "single key and value",
		in:   "foo=1",
		out:  "foo=1",
		tracestate: TraceState{
			list: []member{
				{Key: "foo", Value: "1"},
			},
		},
	},
	{
		name: "single key and value with empty separator",
		in:   "foo=1,",
		out:  "foo=1",
		tracestate: TraceState{
			list: []member{
				{Key: "foo", Value: "1"},
			},
		},
	},
	{
		name: "multiple keys and values",
		in:   "foo=1,bar=2",
		out:  "foo=1,bar=2",
		tracestate: TraceState{
			list: []member{
				{Key: "foo", Value: "1"},
				{Key: "bar", Value: "2"},
			},
		},
	},
	{
		name: "with a key at maximum length",
		in:   "foo=1,zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz=1",
		out:  "foo=1,zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz=1",
		tracestate: TraceState{
			list: []member{
				{
					Key:   "foo",
					Value: "1",
				},
				{
					Key:   "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
					Value: "1",
				},
			},
		},
	},
	{
		name: "with a key and tenant at maximum length",
		in:   "foo=1,ttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttt@vvvvvvvvvvvvvv=1",
		out:  "foo=1,ttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttt@vvvvvvvvvvvvvv=1",
		tracestate: TraceState{
			list: []member{
				{
					Key:   "foo",
					Value: "1",
				},
				{
					Key:   "ttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttt@vvvvvvvvvvvvvv",
					Value: "1",
				},
			},
		},
	},
	{
		name: "with maximum members",
		in:   "bar01=01,bar02=02,bar03=03,bar04=04,bar05=05,bar06=06,bar07=07,bar08=08,bar09=09,bar10=10,bar11=11,bar12=12,bar13=13,bar14=14,bar15=15,bar16=16,bar17=17,bar18=18,bar19=19,bar20=20,bar21=21,bar22=22,bar23=23,bar24=24,bar25=25,bar26=26,bar27=27,bar28=28,bar29=29,bar30=30,bar31=31,bar32=32",
		out:  "bar01=01,bar02=02,bar03=03,bar04=04,bar05=05,bar06=06,bar07=07,bar08=08,bar09=09,bar10=10,bar11=11,bar12=12,bar13=13,bar14=14,bar15=15,bar16=16,bar17=17,bar18=18,bar19=19,bar20=20,bar21=21,bar22=22,bar23=23,bar24=24,bar25=25,bar26=26,bar27=27,bar28=28,bar29=29,bar30=30,bar31=31,bar32=32",
		tracestate: TraceState{
			list: []member{
				{Key: "bar01", Value: "01"},
				{Key: "bar02", Value: "02"},
				{Key: "bar03", Value: "03"},
				{Key: "bar04", Value: "04"},
				{Key: "bar05", Value: "05"},
				{Key: "bar06", Value: "06"},
				{Key: "bar07", Value: "07"},
				{Key: "bar08", Value: "08"},
				{Key: "bar09", Value: "09"},
				{Key: "bar10", Value: "10"},
				{Key: "bar11", Value: "11"},
				{Key: "bar12", Value: "12"},
				{Key: "bar13", Value: "13"},
				{Key: "bar14", Value: "14"},
				{Key: "bar15", Value: "15"},
				{Key: "bar16", Value: "16"},
				{Key: "bar17", Value: "17"},
				{Key: "bar18", Value: "18"},
				{Key: "bar19", Value: "19"},
				{Key: "bar20", Value: "20"},
				{Key: "bar21", Value: "21"},
				{Key: "bar22", Value: "22"},
				{Key: "bar23", Value: "23"},
				{Key: "bar24", Value: "24"},
				{Key: "bar25", Value: "25"},
				{Key: "bar26", Value: "26"},
				{Key: "bar27", Value: "27"},
				{Key: "bar28", Value: "28"},
				{Key: "bar29", Value: "29"},
				{Key: "bar30", Value: "30"},
				{Key: "bar31", Value: "31"},
				{Key: "bar32", Value: "32"},
			},
		},
	},
	{
		name: "with several members",
		in:   "foo=1,bar=2,rojo=1,congo=2,baz=3",
		out:  "foo=1,bar=2,rojo=1,congo=2,baz=3",
		tracestate: TraceState{
			list: []member{
				{Key: "foo", Value: "1"},
				{Key: "bar", Value: "2"},
				{Key: "rojo", Value: "1"},
				{Key: "congo", Value: "2"},
				{Key: "baz", Value: "3"},
			},
		},
	},
	{
		name: "with tabs between members",
		in:   "foo=1 \t , \t bar=2, \t baz=3",
		out:  "foo=1,bar=2,baz=3",
		tracestate: TraceState{
			list: []member{
				{Key: "foo", Value: "1"},
				{Key: "bar", Value: "2"},
				{Key: "baz", Value: "3"},
			},
		},
	},
	{
		name: "with multiple tabs between members",
		in:   "foo=1\t \t,\t \tbar=2,\t \tbaz=3",
		out:  "foo=1,bar=2,baz=3",
		tracestate: TraceState{
			list: []member{
				{Key: "foo", Value: "1"},
				{Key: "bar", Value: "2"},
				{Key: "baz", Value: "3"},
			},
		},
	},
	{
		name: "with space at the end of the member",
		in:   "foo=1 ",
		out:  "foo=1",
		tracestate: TraceState{
			list: []member{
				{Key: "foo", Value: "1"},
			},
		},
	},
	{
		name: "with tab at the end of the member",
		in:   "foo=1\t",
		out:  "foo=1",
		tracestate: TraceState{
			list: []member{
				{Key: "foo", Value: "1"},
			},
		},
	},
	{
		name: "with tab and space at the end of the member",
		in:   "foo=1 \t",
		out:  "foo=1",
		tracestate: TraceState{
			list: []member{
				{Key: "foo", Value: "1"},
			},
		},
	},
}

func TestParseTraceState(t *testing.T) {
	for _, tc := range testcases {
		t.Run(
			tc.name, func(t *testing.T) {
				got, err := ParseTraceState(tc.in)
				assert.Equal(t, tc.tracestate, got)
				if tc.err != nil {
					assert.ErrorIs(t, err, tc.err, tc.in)
				} else {
					assert.NoError(t, err, tc.in)
				}
			},
		)
	}
}

var maxMembers = func() TraceState {
	members := make([]member, maxListMembers)
	for i := 0; i < maxListMembers; i++ {
		members[i] = member{
			Key:   fmt.Sprintf("key%d", i+1),
			Value: fmt.Sprintf("value%d", i+1),
		}
	}
	return TraceState{list: members}
}()

func TestTraceStateInsert(t *testing.T) {
	ts := TraceState{
		list: []member{
			{Key: "key1", Value: "val1"},
			{Key: "key2", Value: "val2"},
			{Key: "key3", Value: "val3"},
		},
	}

	testCases := []struct {
		name       string
		tracestate TraceState
		key, value string
		expected   TraceState
		err        error
	}{
		{
			name:       "add new",
			tracestate: ts,
			key:        "key4@vendor",
			value:      "val4",
			expected: TraceState{
				list: []member{
					{Key: "key4@vendor", Value: "val4"},
					{Key: "key1", Value: "val1"},
					{Key: "key2", Value: "val2"},
					{Key: "key3", Value: "val3"},
				},
			},
		},
		{
			name:       "replace",
			tracestate: ts,
			key:        "key2",
			value:      "valX",
			expected: TraceState{
				list: []member{
					{Key: "key2", Value: "valX"},
					{Key: "key1", Value: "val1"},
					{Key: "key3", Value: "val3"},
				},
			},
		},
		{
			name:       "drop the right-most member(oldest) in queue",
			tracestate: maxMembers,
			key:        "keyx",
			value:      "valx",
			expected: func() TraceState {
				// Prepend the new element and remove the oldest one, which is over capacity.
				return TraceState{
					list: append(
						[]member{{Key: "keyx", Value: "valx"}},
						maxMembers.list[:len(maxMembers.list)-1]...,
					),
				}
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(
			tc.name, func(t *testing.T) {
				actual := tc.tracestate.Insert(tc.key, tc.value)
				assert.Equal(t, tc.expected, actual)
			},
		)
	}
}
