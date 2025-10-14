# Debug 

debug 模块主要方便用户进行异常调试，目前支持打印 invalid utf-8 错误时具体的字段。
使用时通过设置环境变量 GALILEO_SDK_DEBUG，支持组合的方式打开多个 debugger，支持的 debugger 如下：

## utf8 debugger

通过设置环境变量 GALILEO_SDK_DEBUG 为 utf8，框架在 log/trace 导出时遇到 invalid utf-8 错误时会自动在 trpc 的日志中打印相关信息。
