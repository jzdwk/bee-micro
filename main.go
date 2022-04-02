package main

import (
	mybroker "bee-micro/broker"
	_ "bee-micro/routers"
	"flag"
	"fmt"
	"github.com/asim/go-micro/plugins/registry/consul/v3"
	httpServer "github.com/asim/go-micro/plugins/server/http/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/server"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"time"
)

var port = flag.String("port", "8010", "port")

func main() {
	//conf beego
	beego.BConfig.CopyRequestBody = true
	//consul
	reg := consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{"myecs.jzd:65085"}
	})
	//http server
	srv := httpServer.NewServer(
		server.Name(fmt.Sprintf("http-demo-:%v", port)),
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
	)
	service.Init()
	if err := service.Run(); err != nil {
		logs.Error("init service err")
		return
	}
}
