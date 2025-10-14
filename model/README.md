# model

本目录为 ocp omp otp 协议的 go 对象及 抽象定义、辅助函数、对象池等。

ocp omp otp 协议描述：<https://galiosight.ai/galio-sdk-go/proto>

协议设计：

`ocp.pb.go omp.pb.go otp.pb.go` 文件均是自动生成，在 proto 目录执行 `make all`，就行。

`omp_metrics.go` 抽象 omp 监控项为 OMPMetric，减少实现方冗余代码。

`otp_metrics.go` 抽象 otp 监控项为 OTPMetric，减少实现方冗余代码。

`labels.go` 抽象标签为 Labels，减少实现方冗余代码。

`client_metrics.go` 主调监控实现 OMPMetric + OTPMetric 的辅助函数。

`server_metrics.go` 被调监控实现 OMPMetric + OTPMetric 的辅助函数。

`normal_metrics.go` 属性监控实现 OMPMetric + OTPMetric 的辅助函数。

`custom_metrics.go` 自定义监控实现 OMPMetric + OTPMetric 的辅助函数。

`*_pool.go` 定义对象池，方便对象复用，减少 gc。
