package client

import (
	httpClient "bee-micro/micro/client/http"
	tracer2 "bee-micro/micro/tracer"
	"context"
	"github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/v3/client"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/selector"
	"github.com/opentracing/opentracing-go"
	"log"
	"testing"
	"time"
)

var (
	register = "myecs.jzd:65379"
	jaeger   = "myecs.jzd:65031"
)

func TestHttpCli(t *testing.T) {
	//tracing
	tr, io, err := tracer2.NewTracer("http-demo-tracing", jaeger)
	if err != nil {
		log.Fatal(err)
	}
	defer io.Close()
	opentracing.SetGlobalTracer(tr)
	//get service reg
	reg := etcd.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{register}
	})
	//get service selector
	s := selector.NewSelector(selector.Registry(reg), selector.SetStrategy(selector.RoundRobin))

	//new http client
	c := httpClient.NewClient(
		//0.set http client tracer
		true,
		//1. lb selector
		client.Selector(s),
		//2. timeout setting
		client.DialTimeout(time.Second*100),
		client.RequestTimeout(time.Second*100),
		//3. hystrix
		//client.Wrap(wrappers.NewHystrixWrapper()),
		//4. rate limit
		//client.Wrap(ratelimiter.NewClientWrapper(ratelimit.NewBucket(time.Second,int64(1)),false)),
		//5.tracing
		//client.Wrap(micro_opentracing.NewClientWrapper(opentracing.GlobalTracer())),
	)
	//client wrap example
	//c = clientWrapper.NewLogWrap(c)
	for i := 0; i < 10; i++ {
		doGetRequest(t, c)
	}
	//doPostRequest(t, c)
}

func doGetRequest(t *testing.T, c client.Client) {
	request := c.NewRequest("http-demo", "GET:/demo/hello/for-test/get", nil, client.WithContentType("application/json"))
	var response interface{}
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
