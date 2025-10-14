// Copyright 2025 Tencent Galileo Authors
//
// Copyright 2025 Tencent OpenTelemetry Oteam
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

// Package model
package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	modelv1 "galiosight.ai/galio-sdk-go/model"
	"galiosight.ai/galio-sdk-go/semconv"
	"galiosight.ai/galio-sdk-go/version"
)

func TestNewResource(t *testing.T) {
	telemetryTarget := "PCG-123.myapp.myserver"
	deploymentNamespace := modelv1.Production
	deploymentEnvironmentName := "myenv"
	hostIP := "127.0.0.1"
	containerName := "mycontainer"
	serviceSetName := "mystage"
	deploymentCity := "beijing"
	serviceVersion := "v1.0.0"
	rpcSystem := "trpc"

	got := NewResource(telemetryTarget, deploymentNamespace, deploymentEnvironmentName, hostIP, containerName, serviceSetName, deploymentCity, serviceVersion, rpcSystem)

	want := &modelv1.Resource{
		Target:        telemetryTarget,
		Platform:      "PCG-123",
		ObjectName:    "myapp.myserver",
		ServiceName:   "myapp.myserver",
		Namespace:     string(deploymentNamespace),
		EnvName:       deploymentEnvironmentName,
		SetName:       serviceSetName,
		Region:        serviceSetName,
		Instance:      hostIP,
		ContainerName: containerName,
		Version:       version.Number,
		City:          deploymentCity,
		FrameCode:     rpcSystem,
		Language:      semconv.TelemetrySDKLanguageGo.Value.AsString(),
		SdkName:       "galileo",
		App:           "myapp",
		Server:        "myserver",
	}

	assert.Equal(t, got, want)
}
