# runtimes

运行时监控

磁盘监控参考：

fd 监控参考：<https://github.com/VictoriaMetrics/metrics>

go 监控参考：<https://github.com/VictoriaMetrics/metrics>

最近 256 次 gc 的 99 分位（分位值算好了）：`go_gc_duration_seconds{quantile="0.99"}`

gc 的耗时直方图（分位值要自己算）：`go_gc_pause_seconds_bucket{}`

pid 监控参考：

其他进程监控参考：<https://github.com/VictoriaMetrics/metrics>
