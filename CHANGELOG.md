# Change Log

## v0.19.2 (未发布)

- traces:添加 SpanIDInjector，支持指定 SpanID (!781)

## v0.19.1 (2025-04-22)

- {metrics}: 修改 OpenTelemetry metrics 上报协议，从 grpc 修改为 http (!774)
- semconv: OMP v1.0.0 SchemaURL 合并为 omp/v1.0.0 (!755)
- {metrics,traces,logs,profiles,ocp，自监控}: 在 http header 中增加 "X-Galileo-API-Key"，用于接口鉴权 (!770)

## v0.19.0 (2025-03-25)

- semconv: 增加 sse 被调监控上报字段 (!744)
- semconv: 增加 llm 主调监控上报字段 (!726)

## v0.18.1 (2025-02-10)

- go 自监控：修复自监控初始化使用默认地址导致海外上报会报错的问题 (!669)
- traces: workflow path 采样更改为后置采样，解决误将 sampled 状态传给下游，造成下游非预期的命中采样 (!671)
- selflog：自监控日志支持修改日志路径 (!673)

## v0.18.0 (2024-12-26)

- metrics: 增加 OpenTelemetry metrics 协议和 API 支持 (!634)
- {metrics,traces,logs}: OpenTelemetry SDK 从 v1.14.0 升级到 v1.28.0，golang 从 1.18 升级到 1.21，避免业务使用高版本 otel 时，或与天机阁一起编译时的编译错误 (!637)
- traces: 捕获 sonic 的 panic 并输出日志，用于分析 sonic panic 的问题 (!646)
- {metrics,traces,logs,profiles}: 修复 NewConfig 本地配置不生效的问题 (!648)
- go.mod: 直接引用 opentelemetry-go-ecosystem 以保持和云观 Oteam 的编译兼容性 (!649)
- metrics: 主被调指标增加 flow_tag 字段 (!650,!653)
- events: 自定义事件上报支持增加自定义 tag (!655)
- metrics: 修复在有代理的情况下网络判定不正确的问题，避免在 DevCloud 及办公网上报指标数据会报错误 (!657)
- metrics: 启动时设置失败重试次数，避免没有配置热更新时失败重试次数一直是 0 的问题 (!657)
- semconv: 使用规范的 OMP 2.0 SchemaURL 上报 (!659)
- go.mod: 解决 prometheus/common v0.48.0 及以上版本编译错误的问题 (!665)

## v0.17.1 (2024-11-21)

- metrics: 上报时增加 release_version 字段，方便进行 CD 可观测 (!625)
- ocp: 修复机器负载特别高协程被卡死时，retry 函数未执行会导致进程 panic 的问题 (!627)

## v0.17.0 (2024-10-30)

- metrics: 增加 prometheus push 上报能力，支持直接使用 prometheus API 进行数据上报 (!587)
- ocp: 增加 UnregisterResource 方法，方便将已经过期的 target 注销掉 (!603)
- ocp: 移除掉 ocp 协议字段中的 `json:omitempty` 属性，否则会导致无法将 false, 0, "" 等零值下发到 SDK (!603)

## v0.16.5 (2024-09-29)

- logs: 增加 log_traced_type 配置，支持直接通过配置开启命中染色突破日志级别，和 cpp 的配置保持一致 (!582)
- metrics: 修复维度屏蔽配置无法热更新 (!584)

## v0.16.4 (2024-09-03)

- {metrics,traces,logs}: 发送给后台的数据的请求头里面都带上 TenantHeaderKey 和 TargetHeaderKey，方便后台进行数据路由和故障隔离 (!572)
- traces: 增加高版本 otel 编译的 case (!574)

## v0.16.3 (2024-08-21)

- version: 升级版本号 (!571)

## v0.16.2 (2024-08-14)

- go.mod: github.com/golang 改成 google.golang.org/protobuf，避免直接引用过时的库 (!569)

## v0.16.1 (2024-08-13)

- {traces,logs,metrics}: 不再修复 namespace, 统一由后台修复 (!566)

## v0.16.0 (2024-08-13)

## v0.15.5 (2024-08-13)

- {traces,logs,metrics}: 不再修复 namespace, 统一由后台修复 (27224972)

## v0.15.4 (2024-08-02)

- metrics: Namespace 不正确时，统一修正为正式环境，与后台默认行为保持一致 (!561)
- selflog：修复错误日志无法关闭的问题，移除全局变量 log.SelfLogger，优化日志格式，支持完整行号 (!562)

## v0.15.3 (2024-07-31)

- {traces,logs}: ServiceName 从四段式改成两段式，与 ObjectName 保持一致，对应 OpenTelemetry 的 service.name (!554)
- metrics: Namespace 不正确时，统一修正为测试环境 (!554)
- traces: 修复容器名字段不正确的 bug，将容器名字段恢复成 container.name, 方便 WEB UI 页面通过容器名过滤数据 (!555)

## v0.15.2 (2024-07-09)

- traces: 支持针对主调接口、被调接口分别配置采样率，在接口流量不平衡时，以更小的成本来采集感兴趣的 trace (!537)
- traces: 当 body 中有非法字符时，自动切换到标准模式进行序列化，以避免数据持续上报失败 (!545)

## v0.15.1 (2024-07-08)

- {traces,logs}: 增加 panic 捕获，避免 protobuf encode 导致的 panic (!542)

## v0.15.0 (2024-06-13)

- {metrics,profiles,traces,logs}: 支持外网、内网上报。内网使用 http，外网使用 https。支持通过配置指定接入点 (!531)
- {traces,logs}: 默认上报方式从 grpc 改成 http (!531)
- {metrics,profiles}: 对直连 ip 进行网络测试，排除连接不上的 ip, 以避免网络无法直连时的上报错误 (!531)

## v0.14.2 (2024-06-06)

- go.mod: 不引入 trpc-go，以避免 trpc-go 的 v0 和 v2 版本冲突导致的问题 (!525)

## v0.14.1 (2024-05-17)

- ocp：修复监控项分桶配置不能热更新的问题。 (!514)

## v0.14.0 (2024-05-09)

- logs: 修复通过 WithContextDyeingLevel 设置的只上报染色日志的功能非预期的问题 (!497)
- traces: 修复 base.WithSpan 调用后置采样时机不对的问题 (!498)
- ocp: 允许 trpc 服务上报 admin 端口号 (!505)

## v0.13.12 (2024-04-10)

- metrics: 修复数据点更新偶发的 panic (!491)

## v0.13.11 (2024-03-22)

- configs: ocp 热更新支持本地配置优先 (!472)
- traces: 支持用户自定义采样器 (!480)
  :warning: 不兼容的变更：废弃 CleanRPCMethod 方法 :warning:
  废弃原因见[tapd]()

## v0.13.10 (2024-03-04)

- traces: affinity 字段增加 kind，方便确认哪个服务发起的 (!459)
- configs: 支持 ocp 配置热更新，base SDK 也可以定时拉取染色名单、耗时分桶等信息进行热更新了 (!465)

## v0.13.9 (2024-02-22)

- traces: 重构后置采样逻辑，支持记录后置采样策略结果 (!457)

## v0.13.8 (2024-02-19)

- logs: 修复异步日志中 Write 时没有深拷贝导致部分日志不能正常上报的 bug (!453)
- ocp: 当 ocp 地址为空时，不再发送请求 (!453)

## v0.13.7 (2024-01-29)

- metrics: 数据导出失败重试时，使用域名上报，避免 otp 服务发布时直连导致的连接问题 (!445)
- logs: 支持 Sync 方法，业务可以调用 Sync 方法，避免在进程退出时丢失日志的问题 (!448)

## v0.13.6 (2024-01-24)

- {traces,logs}: add 调试开关可以打印出 invalid UTF-8 错误时具体的字段 (!433)

## v0.13.5 (2024-01-12)

- metrics: histogram 指标上报时，不再上报 count 为 0 的桶，以减少上报量 (!435)
- metrics: 调整 histogram 默认分桶配置，以提升 5ms 以下或 1s 以上耗时指标的数据精度 (!435)

## v0.13.4 (2024-01-03)

- traces: 修正当 workflow path 采样没有命中时，没有正确的传递 parent path 值 (!427)

## v0.13.3 (2023-12-28)

- logs: only_trace_log 下普通日志也不打印 (!272)
- ocp: ocp 协议增加 server、client 采样率配置 (!421)
- traces: 将 trpc Namespace 和 env 在 processor 统一处理，解决 inter span 无环境量 (!404)

## v0.13.2 (2023-12-11)

- traces: min count 采样只计算被调，减少默认的 min_count 采样，减少用户的 trace 成本。 (!405)

## v0.13.1 (2023-11-28)

- metrics: 移除元数据上报，减少一个协程，略微减少 CPU 和内存 (!394)
- example: 修复示例文档中 DirectIpPort 问题，导致后端迁移时数据上报有问题的 bug (!394)

## v0.13.0 (2023-11-27)

- SDK: 支持 golang 1.18，方便部分老业务继续使用 (!392)
- ci: 增加流水线：
  - go1.18+Oteam 0.4.1 +galileo latest
  - go1.20+Oteam 0.5.2 + galileo latest
  - go1.21+Oteam latest + galileo latest

## v0.12.1 (2023-11-20)

- traces: 升级 otel 版本，避免业务使用高版本 otel 时，或与天机阁一起编译时的编译错误 (!386)

## v0.12.0 (2023-11-14)

## v0.11.3 (2023-11-14)

- logs: 修正日志等级解析错误导致未上报日志 (!379)
- logs: 提高 ctx_core 兼容性，避免 trpc-go-cls 等插件报错 (!383)

## v0.11.2 (2023-11-06)

## v0.11.1 (2023-11-02)

- config: 修正因 autocorrect 更改的 default config 值 (!375)

## v0.11.0 (2023-11-02)

- ocp: 支持 inplace 更新本地配置 (!366)
- config: 默认禁用随机采样 (!368, !370)
- config: 支持关闭染色采样 (!368)
- traces: TraceState, TraceContext 性能优化，所有场景都有 2-5 倍性能提。(!371)
- traces: MinCount 采样使用 haxmap 无锁 map 替换 sync.Map 实现，性能略有提升。 (!371)
- logs: 修复日志等级初始化错误 (!374)

## v0.11.0-rc.0 (2023-10-27)

- ocp: 支持上报本地配置
- omp: 增加 canary 字段表示金丝雀流量
- semconv: 增加 rpc.canary 字段表示金丝雀流量
- semconv: 整理包结构，清理冗余包
- semconv: 更新文档
- traces: 增加过载保护丢弃 span 自监控上报
- traces: 后置采样器配置支持热更新
- profiles: 修复增量计算 mutex 和 block 部分数据负值的问题
- logs: 重构日志结构，废弃 CoreOptions 等方法 (!364)
- logs: 增加 ContextWith 方法，方便扩展功能，提供默认 WithContextDyeingLevel, WithContextSampleLevel 内置实现 (!364)

## v0.10.5 (2023-09-26)

- profiles: 优化 profile 增量采集逻辑，大幅优化增量采集的内存消耗和内存分配。

## v0.10.4 (2023-09-26)

## v0.10.3 (2023-09-25)

## v0.10.2 (2023-09-13)

- traces: 更新 log mode 枚举值，默认禁用 flow log

## v0.10.1 (2023-08-25)

- traces: 支持过载保护，根据上游采样类型设置阈值，超过阈值时不继承上游采样

## v0.10.0 (2023-08-17)

- {go.mod,traces,logs}: 升级 otel SDK 版本到 v1.16.0, proto/otlp 到 v1.0.0.
- go 升级到 1.19.

## v0.9.7 (2023-08-16)

- profiles: 优化 CPU profile warn，仅在 cpu_profile_rate 设置为非默认值（100 Hz）时打印 warn，详情[参考]()。

## v0.9.6 (2023-08-15)

- metrics: 秒级聚合，增加 1s 5s 10s 处理器。
- metrics: 减少聚合器初始化内存，优化低基数场景的内存占用。

## v0.9.5 (2023-07-24)

- logs: 新增 eventLogger，事件上报不依赖于远程日志开启

## v0.9.3 (2023-07-24)

- profiles: 增加 profile 数据关联 trace 染色

## v0.9.2

- model: 增加 NewResource 方法
- example: 完善示例代码可读性

## v0.9.1 (2023-06-29)

- profiles: 修复 trace exporter Start 循环调用导致 stack overflow 问题

## v0.9.0 (2023-06-26)

- profiles: 增加 profiles 采集逻辑，支持 CPU、heap、mutex、block、goroutine 的 profiles 数据采集
- profiles: 支持 heap、mutex、block 数据的增量采集
- profiles: 支持 profile 数据关联 trace 的 span id

## v0.8.6 (2023-06-09)

traces: 修复 kafka 0.11 以下版本不支持 header，导致写数据失败的问题

## v0.8.5 (2023-06-02)

- {metrics,traces}: 修复被调监控缺少主调 method 问题。

## v0.8.3 (2023-05-24)

- traces: 修复 workflow path，上游只有 workflow 采样，下游命中 follow parent 用户采样的问题

## v0.5.0 (2023-04-13)

- version: 更新版本号到 v0.5.0，以便符合业界规范。每次新增特性，增加一个次版本号，修订号只用于修复 bug。

## v0.3.22 (2023-04-11)

- logs: 增加日志 API, 增加事件 API
- traces: 增加采样标记 galileo.state 字段
- traces: 增加 workflow path 默认采样策略
- traces: 支持布隆过滤器染色。
- traces: 修复调用 kafka 时，追踪信息未透传下去的 bug。
- metrics: 增加秒级监控。
- metrics: 修改指标处理逻辑，采用聚合数据直接发送，不再保留在内存中，减少内存开销。
