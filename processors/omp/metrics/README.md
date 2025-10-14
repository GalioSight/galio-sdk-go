# metrics

omp 协议的监控处理器实现。

omp 协议定义：<https://galiosight.ai/galio-sdk-go/proto>

omp 协议设计：

`processor.go`：处理器主文件，处理函数入口、定时导出管理。

`shard.go`：监控数据分片管理。

`multi.go`：多值点管理。

`point`：底层数据抽象，及各种类型实现。

`hash_bytes.go`：hash bytes 对象池、hash 计算辅助函数。

整体结构：processor → shard → multi → point。
