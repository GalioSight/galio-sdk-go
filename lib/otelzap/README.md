# otelzap

OpenTelemetry 协议 zap 实现。

实现参考自：

<https://github.com/open-telemetry/opentelemetry-go/blob/main/sdk/trace/batch_span_processor.go>

总体流程：用户打日志 -> core -> write syncer -> exporter -> OpenTelemetry collector。

write syncer 主流程：Write -> 判定写条件 -> 转协议 -> Enqueue -> 写 queue

write syncer 异步流程：

异步 1：processQueue -> 读 queue -> 凑齐 batch -> 判定凑足 -> 调用 exporter ExportLogs

异步 2：processQueue -> 定时任务 -> 调用 exporter ExportLogs
