package client

import (
	httpClient "bee-micro/client/httpclient"
	"bee-micro/config"
	"context"
	"github.com/asim/go-micro/plugins/registry/consul/v3"
	micro_opentracing "github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	"github.com/asim/go-micro/v3/client"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/selector"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaeger_config "github.com/uber/jaeger-client-go/config"
	"io"
	"log"

	//opentracing "github.com/opentracing/opentracing-go"
	"testing"
	"time"
)

func TestHttpCli(t *testing.T) {
	//tracing
	tracer, io, err := NewTracer("http-demo-tracing-cli", "192.168.143.146:6831")
	if err != nil {
		log.Fatal(err)
	}
	defer io.Close()
	opentracing.SetGlobalTracer(tracer)

	//读取配置中心
	cfg, _ := config.GetConfig()

	info, _ := config.GetConfigInfo(cfg, "config")

	//get service reg
	reg := consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{info.Consul.Address}
	})

	//get service selector
	s := selector.NewSelector(selector.Registry(reg), selector.SetStrategy(selector.RoundRobin))

	//new http client
	c := httpClient.NewClient(
		//1. lb selector
		client.Selector(s),
		//2. timeout setting
		client.DialTimeout(time.Second*10),
		client.RequestTimeout(time.Second*10),
		//3. hystrix
		//client.Wrap(wrappers.NewHystrixWrapper()),
		//tracing
		client.Wrap(micro_opentracing.NewClientWrapper(opentracing.GlobalTracer())),
	)

	doGetRequest(t, c)
	//doPostRequest(t, c)
}

func doGetRequest(t *testing.T, c client.Client) {
	request := c.NewRequest("http-demo", "GET:/demo/hello/for-test/get", nil, client.WithContentType("application/json"))
	var response Resp
	if err := c.Call(context.Background(), request, &response); err != nil {
		t.Error(err.Error())
		return
	}
	t.Log("do get request success")
}

func doPostRequest(t *testing.T, c client.Client) {

	req := struct {
		Name string
		Age  int
	}{"jzd", 123}
	request := c.NewRequest("http-demo", "POST:/demo/hello/for-test/post", req, client.WithContentType("application/json"))
	var response Resp
	if err := c.Call(context.Background(), request, &response); err != nil {
		t.Error(err.Error())
		return
	}
	t.Log("do post request success")
}

//msg
type Resp struct {
	Method  string
	Message string
}

//创建链路追踪实例
func NewTracer(serviceName string, addr string) (opentracing.Tracer, io.Closer, error) {
	cfg := &jaeger_config.Configuration{
		ServiceName: serviceName,
		Sampler: &jaeger_config.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaeger_config.ReporterConfig{
			BufferFlushInterval: 1 * time.Second,
			LogSpans:            true,
			LocalAgentHostPort:  addr,
		},
	}
	return cfg.NewTracer()
}
