package client

import (
	httpClient "bee-micro/client/http"
	"bee-micro/controllers"
	client2 "bee-micro/wrappers/client"
	"context"
	"fmt"
	"github.com/asim/go-micro/plugins/registry/consul/v3"
	"github.com/asim/go-micro/v3/client"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/selector"
	"testing"
	"time"
)

func TestHttpCli(t *testing.T) {
	//get service reg
	reg := consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{"myecs.jzd:65085"}
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
		//3. hystrix breaker
		client.Wrap(client2.NewHystrixWrapper()),
		//4. client rate-limit
		//client.Wrap(ratelimiter.NewClientWrapper(ratelimit.NewBucket(time.Second,int64(1)),false)),
	)
	for i := 0; i < 10; i++ {
		doGetRequest(t, c)
		//doPostRequest(t, c)
	}
	//doPostRequest(t, c)
}

func doGetRequest(t *testing.T, c client.Client) {
	request := c.NewRequest("http-demo", "GET:/demo/hello/for-test/get", nil, client.WithContentType("application/json"))
	var response controllers.Resp
	if err := c.Call(context.Background(), request, &response); err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(fmt.Printf("do get request success, %s, %s", response.Result, response.Message))
}

func doPostRequest(t *testing.T, c client.Client) {
	req := struct {
		Name string
		Age  int
	}{"jzd", 123}
	request := c.NewRequest("http-demo", "POST:/demo/hello/for-test/post", req, client.WithContentType("application/json"))
	var response controllers.Resp
	if err := c.Call(context.Background(), request, &response); err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(fmt.Printf("do get request success, %s, %s", response.Result, response.Message))
}
