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
	"testing"

	"galiosight.ai/galio-sdk-go/version"

	"github.com/stretchr/testify/assert"
)

func TestNewResource(t *testing.T) {
	platform := "PCG-123"
	app := "myapp"
	server := "server1"
	service := "myservice"
	namespace := Production
	env := "formal"
	set := "set1"
	city := "sh"
	instance := "2001:db8:2de::e88"
	containerName := "container1"

	resource := NewResource(platform, app, server, service, namespace, env, set, city, instance, containerName)

	// Assert that the created resource has the correct values
	assert.Equal(t, platform+"."+app+"."+server, resource.Target)
	assert.Equal(t, platform, resource.Platform)
	assert.Equal(t, app+"."+server, resource.ObjectName)
	assert.Equal(t, app, resource.App)
	assert.Equal(t, server, resource.Server)
	assert.Equal(t, app+"."+server, resource.ServiceName)
	assert.Equal(t, namespace, Namespace(resource.Namespace))
	assert.Equal(t, env, resource.EnvName)
	assert.Equal(t, set, resource.SetName)
	assert.Equal(t, set, resource.Region)
	assert.Equal(t, city, resource.City)
	assert.Equal(t, instance, resource.Instance)
	assert.Equal(t, containerName, resource.ContainerName)
	assert.Equal(t, version.Number, resource.Version)
	assert.Equal(t, "trpc", resource.FrameCode)
	assert.Equal(t, "go", resource.Language)
	assert.Equal(t, "galileo", resource.SdkName)
}

func TestFixNamespace(t *testing.T) {
	testCases := []struct {
		input    Namespace
		expected Namespace
	}{
		{Production, Production},
		{Development, Development},
		{"Custom", Production},
		{"", Production},
	}

	for _, tc := range testCases {
		t.Run(
			"", func(t *testing.T) {
				r := Resource{
					Namespace: string(tc.input),
				}
				r.FixNamespace()
				assert.Equal(t, tc.expected, Namespace(r.Namespace))
			},
		)
	}
}
