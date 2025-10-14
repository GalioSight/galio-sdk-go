# DEMO

1. simple:
   包含 galileo otp metric, galileo(opentelemetry) trace, galileo(opentelemetry) log OMP 3.0 上报

2. prometheus:
   包含 prometheus metric OMP 3.0 上报

3. dice:
   包含 OpenTelemetry metric/trace/log, OpenTelemetry 官方 demo, 没有 OMP 3.0 相关内容，主要演示自定义指标，trace 上报

4. opentelemetry:
   包含 OpenTelemetry metric OMP 3.0 上报，galileo 的 trace/log 和 OpenTelemetry 没有区别，因此 OpenTelemetry trace/log 请参考 simple
   适用于全新接入的服务

5. otelgrpc:
   包含 OpenTelemetry trace OTEL 1.0 schema URL 上报，即已经使用 otel 插桩的服务如何快速接入伽利略，
   并将 otel semconv 转换成 OMP 3.0 semconv, 从而支持伽利略平台 trace log 跳转等功能
