// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2024 Tencent Galileo Authors

// Code generated from semantic convention specification. DO NOT EDIT.

package semconv // import "go.opentelemetry.io/otel/semconv/"

import "go.opentelemetry.io/otel/attribute"

// Namespace: cmdb
const (

	// cmdb ID
	// Stability: Development
	// Type: string
	// Deprecated: 废弃
	//
	// Examples: ""

	CmdbModuleIDKey = attribute.Key("cmdb.module.id")
)

// Namespace: net
const (

	// Stability: Development
	// Type: string
	//
	// Examples: "10.30.50.70"

	NetHostIPKey = attribute.Key("net.host.ip")
	// Stability: Development
	// Type: string
	//
	// Examples: ""

	NetHostNameKey = attribute.Key("net.host.name")
	// Stability: Development
	// Type: string
	//
	// Examples: "8080"

	NetHostPortKey = attribute.Key("net.host.port")
	// Stability: Development
	// Type: string
	//
	// Examples: "10.30.50.70"

	NetPeerIPKey = attribute.Key("net.peer.ip")
	// Stability: Development
	// Type: string
	//
	// Examples: "8080"

	NetPeerPortKey = attribute.Key("net.peer.port")
)

// Namespace: other
const (

	// 被调 set
	// Stability: Development
	// Type: string
	//
	// Examples: "set.sz.1"

	CalleeConSetidKey = attribute.Key("callee_con_setid")
	// 被调容器
	// Stability: Development
	// Type: string
	//
	// Examples: "test.galileo.metaserver.sz100012"

	CalleeContainerKey = attribute.Key("callee_container")
	// 被调 ip
	// Stability: Development
	// Type: string
	//
	// Examples: "10.30.50.70"

	CalleeIPKey = attribute.Key("callee_ip")
	// 被调方法
	// Stability: Development
	// Type: string
	//
	// Examples: "GetPromQueryURLByTarget"

	CalleeMethodKey = attribute.Key("callee_method")
	// 被调服务，123 平台上的服务为 app.server，可能会和别的平台的服务重名
	// Stability: Development
	// Type: string
	//
	// Examples: "galileo.metaserver"

	CalleeServerKey = attribute.Key("callee_server")
	// 被调 service
	// Stability: Development
	// Type: string
	//
	// Examples: "trpc.galileo.metaserver.TargetDataService"

	CalleeServiceKey = attribute.Key("callee_service")
	// 主调 set
	// Stability: Development
	// Type: string
	//
	// Examples: "set.sz.1"

	CallerConSetidKey = attribute.Key("caller_con_setid")
	// 主调容器
	// Stability: Development
	// Type: string
	//
	// Examples: "test.galileo.apiserver.sz100012"

	CallerContainerKey = attribute.Key("caller_container")
	// 主调流量分组
	// Stability: Development
	// Type: string
	//
	// Examples: "qq"
	// Note: 常用于中台内部服务需要根据主调业务划分监控，通常需要用户主动填充，rpc 框架无法自动填充

	CallerGroupKey = attribute.Key("caller_group")
	// 主调 ip
	// Stability: Development
	// Type: string
	//
	// Examples: "10.20.30.40"

	CallerIPKey = attribute.Key("caller_ip")
	// 主调方法
	// Stability: Development
	// Type: string
	//
	// Examples: "getData"

	CallerMethodKey = attribute.Key("caller_method")
	// 主调服务，123 平台上的服务为 app.server，可能会和别的平台的服务重名
	// Stability: Development
	// Type: string
	//
	// Examples: "galileo.apiserver"

	CallerServerKey = attribute.Key("caller_server")
	// 主调 service
	// Stability: Development
	// Type: string
	//
	// Examples: "trpc.galileo.apiserver.DataService"

	CallerServiceKey = attribute.Key("caller_service")
	// 金丝雀流量
	// Stability: Development
	// Type: string
	//
	// Examples: "1"

	CanaryKey = attribute.Key("canary")
	// 部署城市，类似于 otel 的 cloud.region
	// Stability: Development
	// Type: string
	// Deprecated: use deployment.city instead
	//
	// Examples: "gz"

	CityKey = attribute.Key("city")
	// 返回码
	// Stability: Development
	// Type: string
	//
	// Examples:
	// "err_101",
	// "0",
	// "ret_100",
	//
	// Note: err_开头的默认表示异常，ret_开头的默认表示成功，特别的 0 表示成功

	CodeKey = attribute.Key("code")
	// 返回码类型，区分成功，异常，超时
	// Stability: Development
	// Type: Enum
	//
	// Examples: undefined

	CodeTypeKey = attribute.Key("code_type")
	// 本机所在的 set 名
	// Stability: Development
	// Type: string
	// Deprecated: use service.set.name instead
	//
	// Examples: "set.sz.1"

	ConSetidKey = attribute.Key("con_setid")
	// Stability: Development
	// Type: string
	// Deprecated: use container.name instead
	//
	// Examples: ""

	ContainerNameCamelKey = attribute.Key("containerName")
	// 本机容器名
	// Stability: Development
	// Type: string
	// Deprecated: use container.name instead
	//
	// Examples: "test.galileo.apiserver.sz100012"

	ContainerNameSnakeKey = attribute.Key("container_name")
	// Stability: Development
	// Type: string
	// Deprecated: use deployment.environment.name instead
	//
	// Examples: ""

	EnvKey = attribute.Key("env")
	// 用户环境
	// Stability: Development
	// Type: string
	// Deprecated: use deployment.environment.name instead
	//
	// Examples:
	// "formal",
	// "test",
	// "d22b30d0",
	//
	// Note: 细分的环境，例如每个人都可以在测试环境创建自己的特性环境，自行开发实现隔离。 env_name 除了固定的 formal、test 外，通常形如 d22b30d0 这样的格式

	EnvNameKey = attribute.Key("env_name")
	// 等同于 env_name
	// Stability: Development
	// Type: string
	// Deprecated: use env_name instead
	//
	// Examples:
	// "formal",
	// "test",
	// "d22b30d0",

	EnvnameKey = attribute.Key("envname")
	// 定义流量标签，如灰度、降级、重试
	// Stability: Development
	// Type: Enum
	//
	// Examples: "1"

	FlowTagKey = attribute.Key("flow_tag")
	// 本机 IP 地址
	// Stability: Development
	// Type: string
	// Deprecated: use host.ip instead
	//
	// Examples: "10.20.30.40"

	InstanceKey = attribute.Key("instance")
	// 区分正式环境和测试环境，因为和 k8s 的 namespace 有冲突，因此推荐换成 development.namespace
	// Stability: Stable
	// Type: Enum
	// Deprecated: use deployment.namespace instead
	//
	// Examples: undefined

	NamespaceKey = attribute.Key("namespace")
	// 组织名，用于商业版中代替 target 中的 platform
	// Stability: Development
	// Type: string
	// Deprecated: use telemetry.organization instead
	//
	// Examples: "513F30A49CF9"

	OrganizationIDKey = attribute.Key("organization_id")
	// 发布服务自身的版本号
	// Stability: Development
	// Type: string
	// Deprecated: use service.version instead
	//
	// Examples: "1.0"

	ReleaseVersionKey = attribute.Key("release_version")
	// SDK 名称
	// Stability: Development
	// Type: string
	// Deprecated: use telemetry.sdk.name instead
	//
	// Examples: "galileo"

	SDKNameKey = attribute.Key("sdk_name")
	// Stability: Development
	// Type: string
	// Deprecated: 废弃
	//
	// Examples: ""

	ServerKey = attribute.Key("server")
	// Stability: Development
	// Type: string
	// Deprecated: use service.set.name instead
	//
	// Examples: ""

	SetNameCamelKey = attribute.Key("setName")
	// 观测对象的唯一标识 ID，需要全局唯一
	// Stability: Stable
	// Type: string
	// Deprecated: use telemetry.target instead
	//
	// Examples: "PCG-123.galileo.metaserver"
	// Note: target 是 galileo 的核心概念之一，任何观测数据，都必须绑定到一个 target 上。 <p>target 是观测对象的唯一标识 ID，必须全局唯一，target 相同的，就认为是同一个对象。</p> <p>target 必须全局唯一，为了避免冲突，格式上分为两部分，用点分割。</p>
	//
	//   - 第一部分是平台 (platform)，第二部分是平台内的对象名称 (object_name)。
	//   - 第一个点之前的部分，是 platform，如 PCG-123，不同平台的 platform 是不同的。
	//   - 第一个点之后的部分，是 object_name，如 galileo.metaserver。 target = platform . object_name

	TargetKey = attribute.Key("target")
	// 用户扩展字段 1
	// Stability: Development
	// Type: string
	//
	// Examples: ""

	UserExt1Key = attribute.Key("user_ext1")
	// 用户扩展字段 2
	// Stability: Development
	// Type: string
	//
	// Examples: ""

	UserExt2Key = attribute.Key("user_ext2")
	// 用户扩展字段 3
	// Stability: Development
	// Type: string
	//
	// Examples: ""

	UserExt3Key = attribute.Key("user_ext3")
	// galileo SDK 版本号，构建 metrics resource 时使用
	// Stability: Development
	// Type: string
	// Deprecated: use telemetry.sdk.version instead
	//
	// Examples:
	// "v0.17.0",
	// "cpp-1.17.0",
	// "py-1.17.0",

	VersionKey = attribute.Key("version")
)

// Enum values for code_type
var (

	// 成功
	// Stability: development

	CodeTypeSuccess = CodeTypeKey.String("success")
	// 异常
	// Stability: development

	CodeTypeException = CodeTypeKey.String("exception")
	// 超时
	// Stability: development

	CodeTypeTimeout = CodeTypeKey.String("timeout")
)

// Enum values for flow_tag
var (

	// 灰度流量
	// Stability: none

	FlowTagGray = FlowTagKey.String("Gray")
	// 降级流量
	// Stability: none

	FlowTagDowngrade = FlowTagKey.String("Downgrade")
	// 重试流量
	// Stability: none

	FlowTagRetry = FlowTagKey.String("Retry")
)

// Enum values for namespace
var (

	// 测试环境
	// Stability: development

	NamespaceDevelopment = NamespaceKey.String("Development")
	// 正式环境
	// Stability: development

	NamespaceProduction = NamespaceKey.String("Production")
)

// Namespace: server
const (

	// 服务所属 owner
	// Stability: Development
	// Type: string
	// Deprecated: 废弃
	//
	// Examples: "toraxie, andyning"

	ServerOwnerKey = attribute.Key("server.owner")
)

// Namespace: set
const (

	// 本机所在的 set 名
	// Stability: Development
	// Type: string
	// Deprecated: use service.set.name instead
	//
	// Examples: "set.sz.1"

	SetNameKey = attribute.Key("set.name")
)

// Namespace: tps
const (

	// 租户
	// Stability: Development
	// Type: string
	// Deprecated: 废弃
	//
	// Examples: "default"

	TpsTenantIDKey = attribute.Key("tps.tenant.id")
)

// Namespace: trpc
const (

	// 等同于 callee_method
	// Stability: Development
	// Type: string
	//
	// Examples: "getData"

	TrpcCalleeMethodKey = attribute.Key("trpc.callee_method")
	// 等同于 env_name
	// Stability: Development
	// Type: string
	//
	// Examples: "galileo.apiserver"

	TrpcCalleeServerKey = attribute.Key("trpc.callee_server")
	// 等同于 callee_service
	// Stability: Development
	// Type: string
	//
	// Examples: "trpc.galileo.apiserver.DataService"

	TrpcCalleeServiceKey = attribute.Key("trpc.callee_service")
	// 主调接口 等同于 caller_method
	// Stability: Development
	// Type: string
	//
	// Examples: "getData"

	TrpcCallerMethodKey = attribute.Key("trpc.caller_method")
	// 等同于 caller_server
	// Stability: Development
	// Type: string
	//
	// Examples: "galileo.apiserver"

	TrpcCallerServerKey = attribute.Key("trpc.caller_server")
	// 主调 service, 等同于 caller_service
	// Stability: Development
	// Type: string
	//
	// Examples: "trpc.galileo.apiserver.DataService"

	TrpcCallerServiceKey = attribute.Key("trpc.caller_service")
	// 等同于 env_name
	// Stability: Development
	// Type: string
	// Deprecated: use deployment.environment.name instead
	//
	// Examples: "formal"

	TrpcEnvnameKey = attribute.Key("trpc.envname")
	// 等同于 namespace
	// Stability: Development
	// Type: string
	// Deprecated: use deployment.namespace instead
	//
	// Examples: "Production"

	TrpcNamespaceKey = attribute.Key("trpc.namespace")
	// 协议
	// Stability: Development
	// Type: string
	//
	// Examples: "trpc"

	TrpcProtocolKey = attribute.Key("trpc.protocol")
	// 等同于 code
	// Stability: Development
	// Type: int
	//
	// Examples: undefined

	TrpcStatusCodeKey = attribute.Key("trpc.status_code")
	// 返回码消息文本
	// Stability: Development
	// Type: string
	//
	// Examples: ""

	TrpcStatusMsgKey = attribute.Key("trpc.status_msg")
	// 等同于 code_type
	// Stability: Development
	// Type: string
	//
	// Examples: ""

	TrpcStatusTypeKey = attribute.Key("trpc.status_type")
)
