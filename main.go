package main

import (
	mybroker "bee-micro/broker"
	"bee-micro/config"
	_ "bee-micro/routers"
	"flag"
	"fmt"
	"github.com/asim/go-micro/plugins/registry/consul/v3"
	httpServer "github.com/asim/go-micro/plugins/server/http/v3"
	promwrapper "github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/server"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"time"
)

var port = flag.String("port", "8010", "port")

func main() {
	//load config from consul
	cfg, _ := config.GetConfig()
	conf, err := config.GetConsul(cfg, "consul")
	if err != nil {
		return
	}
	//conf beego
	beego.BConfig.CopyRequestBody = true
	//consul
	reg := consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{conf.Address}
	})
	//http server
	serverName := fmt.Sprintf("http-demo")
	serverID := uuid.Must(uuid.NewUUID()).String()
	serverVersion := "v1.0"
	srv := httpServer.NewServer(
		server.Name(serverName),
		server.Address(fmt.Sprintf("localhost:%v", port)),
		server.Broker(mybroker.RedisBk),
		//wrap in server
		/*server.WrapHandler(limiter.NewHandlerWrapper(ratelimit.NewBucket(time.Second,int64(1)),false)),

		server.WrapHandler(promwrapper.NewHandlerWrapper(
			promwrapper.ServiceName(serverName),
			promwrapper.ServiceVersion(serverVersion),
			promwrapper.ServiceID(serverID)),*/
	)
	//http controller
	if err := srv.Handle(srv.NewHandler(beego.BeeApp.Handlers)); err != nil {
		logs.Error(err.Error())
		return
	}
	//filter doesn't work?
	//beego.InsertFilter("/demo/*",beego.AfterExec, mymetrics.Filter)
	//redis broker
	if err := mybroker.Init(); err != nil {
		return
	}
	//init micro service
	service := micro.NewService(
		//health check
		micro.RegisterTTL(time.Second*10),
		micro.RegisterInterval(time.Second*10),
		//backend server
		micro.Server(srv),
		micro.Address(":8100"),
		//service registry
		micro.Registry(reg),
		//msg broker, default http broker
		micro.Broker(mybroker.RedisBk),
		//metrics
		micro.WrapHandler(promwrapper.NewHandlerWrapper(
			promwrapper.ServiceName(serverName),
			promwrapper.ServiceVersion(serverVersion),
			promwrapper.ServiceID(serverID))),

		//circuit breaker&limit, todo server wrapper
		//micro.WrapHandler(wrappers.NewHystrixServerWrapper()),*/
		//rate limit, 10 request per second, 50 alive requests in total
		//tracing
		//logging
	)
	go PrometheusBoot()
	service.Init()
	if err := service.Run(); err != nil {
		logs.Error("init service err")
		return
	}
}

func PrometheusBoot() {
	http.Handle("/metrics", promhttp.Handler())
	// 启动web服务，监听8085端口
	go func() {
		err := http.ListenAndServe("localhost:8085", nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()
}
