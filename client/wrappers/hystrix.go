package wrappers

import (
	"context"
	"github.com/asim/go-micro/v3/server"
	"github.com/astaxie/beego/logs"
	"net/http"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/asim/go-micro/v3/client"
)

type clientWrapper struct {
	client.Client
}

func (c *clientWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	// 命令名的写法参考官方插件，服务名和方法名拼接
	name := req.Service() + "." + req.Endpoint()
	// 自定义当前命令的熔断配置，除了超时时间还有很多其他配置请自行研究
	config := hystrix.CommandConfig{
		Timeout: 2000,
	}
	hystrix.ConfigureCommand(name, config)
	return hystrix.Do(name,
		func() error {
			// 这里调用了真正的服务
			return c.Client.Call(ctx, req, rsp, opts...)
		},
		// 降级函数，定义调用失败的处理
		func(err error) error {
			// 因为是示例程序，只处理请求超时这一种错误的降级，其他错误仍抛给上级调用函数
			if err != hystrix.ErrTimeout {
				return err
			}
			logs.Error("client calls failed, err %v", err.Error())
			return nil
		})
}

// NewHystrixWrapper returns a hystrix client Wrapper.
func NewHystrixWrapper() client.Wrapper {
	return func(c client.Client) client.Client {
		return &clientWrapper{c}
	}
}

//NewHystrixServerWrapper
func NewHystrixServerWrapper(opt interface{}) server.HandlerWrapper {
	return nil
}

type serverWrapper struct {
	http.Server
}
