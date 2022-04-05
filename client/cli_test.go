package client

import (
	httpClient "bee-micro/client/httpclient"
	"bee-micro/client/wrappers"
	"bee-micro/config"
	"context"
	"github.com/asim/go-micro/plugins/registry/consul/v3"
	"github.com/asim/go-micro/v3/client"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/selector"
	"testing"
	"time"
)

func TestHttpCli(t *testing.T) {
	//读取配置中心
	cfg, _ := config.GetConfig()

	info, _ := config.GetConsul(cfg, "consul")

	//get service reg
	reg := consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{info.Address}
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
		client.Wrap(wrappers.NewHystrixWrapper()),
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
