module bee-micro

go 1.16

require (
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/asim/go-micro/plugins/broker/redis/v3 v3.7.0
	github.com/asim/go-micro/plugins/config/encoder/yaml/v3 v3.7.0
	github.com/asim/go-micro/plugins/config/source/consul/v3 v3.7.0
	github.com/asim/go-micro/plugins/registry/consul/v3 v3.7.0
	github.com/asim/go-micro/plugins/server/http/v3 v3.7.0
	github.com/asim/go-micro/plugins/transport/memory/v3 v3.0.0-20210630062103-c13bb07171bc
	github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3 v3.7.0
	github.com/asim/go-micro/plugins/wrapper/ratelimiter/ratelimit/v3 v3.7.0
	github.com/asim/go-micro/plugins/wrapper/ratelimiter/uber/v3 v3.7.0 // indirect
	github.com/asim/go-micro/v3 v3.7.1
	github.com/astaxie/beego v1.12.3
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.2.0
	github.com/juju/ratelimit v1.0.1
	github.com/prometheus/client_golang v1.11.0
)
