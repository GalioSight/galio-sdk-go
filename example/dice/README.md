# 演示 官方 OpenTelemetry SDK 如何结合 galileo SDK 实现

dice 程序是 OpenTelemetry 官方的程序 [说明](https://opentelemetry.io/docs/languages/go/getting-started/), [代码](https://github.com/open-telemetry/opentelemetry-go/tree/v1.25.0/example/dice)

这个 case 演示高版本 otel 如何结合 galileo SDK 编译。

如果你无法编译，请确认你的 go.mod 这两个包和你的 otel 版本一致

```
go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.25.0
go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.25.0
```
