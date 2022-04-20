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

## 目录结构
- /conf beego配置目录，无意义
- /controllers 路由控制
- /dao db操作
- /initial db初始化
- /models db表/模型定义
- /routers 路由定义
- /util 公共库
- /micro go-micro改造/封装库
  - /breaker 熔断封装
  - /broker 异步消息
  - /client http client重写
  - /config 配置中心封装
  - /filter beego过滤器示例
  - /logger 日志封装
  - /metrics 监控重写
  - /ratelimit 限流重写
  - /tracer 链路跟踪重写
