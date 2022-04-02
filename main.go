package main

import (
	mybroker "bee-micro/broker"
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
	//conf beego
	beego.BConfig.CopyRequestBody = true
	//beego.InsertFilter("/*",beego.AfterExec, mymetrics.Filter)
	//consul
	reg := consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{"myecs.jzd:65085"}
	})
	//http server
	serverName := fmt.Sprintf("http-demo-:%v", port)
	serverID := uuid.Must(uuid.NewUUID()).String()
	serverVersion := "v1.0"
	srv := httpServer.NewServer(
		server.Name(serverName),
		server.Address(fmt.Sprintf("localhost:%v", port)),
	)
	//http controller
	if err := srv.Handle(srv.NewHandler(beego.BeeApp.Handlers)); err != nil {
		logs.Error(err.Error())
		return
	}
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
		//msg broker
		micro.Broker(mybroker.RedisBk),
		//metrics
		micro.WrapHandler(promwrapper.NewHandlerWrapper(
			promwrapper.ServiceName(serverName),
			promwrapper.ServiceVersion(serverVersion),
			promwrapper.ServiceID(serverID),
		)),
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
