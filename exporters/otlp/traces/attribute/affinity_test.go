// Copyright 2024 Tencent Galileo Authors

// Package attribute ...
package attribute

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

var keys = &RPCKeys{CallerService: "a", CallerMethod: "b", CalleeService: "c", CalleeMethod: "d"}

func init() {
	Affinity.SetTarget("PCG-123.example.greeter")
}

func TestSerialize(t *testing.T) {

	tests := []struct {
		kind trace.SpanKind
		want string
	}{
		{trace.SpanKindClient, "PCG-123.example.greeter│c│3"},
		{trace.SpanKindServer, "a│PCG-123.example.greeter│2"},
	}

	a := assert.New(t)
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			a.Equal(test.want, Affinity.String(keys, test.kind))
			parsed, err := Affinity.Parse(test.want)
			a.NoError(err)
			a.Equal(test.kind, parsed.Kind)
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		text string
		err  string
		host string
	}{
		{"t│c│3", "", "t"},
		{"a│t│2", "", "t"},
		{"a/b-c/d-3", "invalid syntax", ""},
		{"a/b-c/d", "invalid syntax", ""},
		{"PCG-123.example.greeter│trpc.example.greeter.Greeter│3", "", "PCG-123.example.greeter"},
	}

	a := assert.New(t)
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			parsed, err := Affinity.Parse(test.text)
			if test.err == "" {
				a.NoError(err)
			} else {
				a.ErrorContains(err, test.err)
				return
			}
			a.Equal(test.host, parsed.Host)
		})
	}
}
