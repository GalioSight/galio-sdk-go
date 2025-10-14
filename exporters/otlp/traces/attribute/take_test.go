// Copyright 2024 Tencent Galileo Authors
//
// Package attribute ...
package attribute

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/attribute"

	semconv "galiosight.ai/galio-sdk-go/semconv/v1.0.0"
)

type suited struct {
	suite.Suite
	ctx context.Context
}

func (s *suited) SetupSuite() {
	s.ctx = context.Background()
}

func (s *suited) TearDownSuite() {
}

func (s *suited) SetupTest() {
}

func (s *suited) TearDownTest() {
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(suited))
}

func (s *suited) TestTakeKey() {
	in := []attribute.KeyValue{
		semconv.TrpcCallerServiceKey.String("A"),
		semconv.TrpcProtocolKey.String("B"),
		semconv.TrpcCalleeMethodKey.String("C"),
		semconv.TrpcCalleeServiceKey.String("D"),
		semconv.TrpcCallerMethodKey.String("E"),
	}
	keys := NewRPCKeys(in)
	s.Equal(keys.CallerService, "A")
	s.Equal(keys.CallerMethod, "E")
	s.Equal(keys.CalleeService, "D")
	s.Equal(keys.CalleeMethod, "C")
}
