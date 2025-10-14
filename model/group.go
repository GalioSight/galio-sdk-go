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

// MetricGroup 监控分组。0：主调 1：被调 2：属性 3：自定义。
type MetricGroup int

const (
	// ClientGroup 主调监控。
	ClientGroup MetricGroup = iota
	// ServerGroup 被调监控。
	ServerGroup
	// NormalGroup 属性监控。
	NormalGroup
	// CustomGroup 用户自定义监控。
	CustomGroup
	// MaxGroup 最大分组数，用来终止循环。
	MaxGroup
)
