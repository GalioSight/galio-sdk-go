// Package semconv ...
//
// Copyright 2024 Tencent Galileo Authors
package semconv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"

	omp "galiosight.ai/galio-sdk-go/semconv/v1.0.0"
)

func TestCommonSemconv(t *testing.T) {
	// 增加几个易错常量的 assert
	tests := []struct {
		key  attribute.Key
		text string
	}{
		{omp.ContainerNameSnakeKey, "container_name"},
		{ContainerNameKey, "container.name"},
		{omp.ConSetidKey, "con_setid"},
		{omp.EnvNameKey, "env_name"},
		{omp.EnvnameKey, "envname"},
		{omp.SetNameKey, "set.name"},
	}

	a := assert.New(t)
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			a.Equal(test.text, string(test.key))
		})
	}
}
