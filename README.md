# beego&go-micro demo
使用go-micro v3版本结合beego的一个demo.目前整合了：

| 功能      | 服务端 |  客户端 |  三方件 |
| --------- | ------|-------|-------|
| 注册发现    |   ✓    | ✓    |etcd v3  |
| 负载均衡    |        | ✓      |     |
| 请求限流    |   ✓    |       |github.com/juju/ratelimit   |
| 熔断降级    |       | ✓      |github.com/afex/hystrix-go/hystrix   |
| 监控告警    |   ✓    |       |prometheus/grafana   |
| 配置中心    |   ✓    | ✓      |etcd v3 |
| 链路追踪    |   ✓    | ✓      |jaeger  |

主要修改了框架以下功能:

1. 修改http client实现，支持restful api
2. 修改基于jaeger的http client trace实现，支持将span context通过http header传递
3. 修改metrics/ratelimit/tracer位于http server端的wrapper实现，通过http handler重新封装
4. 提供基于beego filter实现的示例
