package main

import (
	"github.com/asim/go-micro/plugins/registry/consul/v3"
	httpServer "github.com/asim/go-micro/plugins/server/http/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/server"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	_ "go-micro-demo/routers"
)

func main() {
	//conf beego
	beego.BConfig.CopyRequestBody = true
	//consul
	reg := consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{"myecs.jzd:65085"}
	})
	srv := httpServer.NewServer(
		server.Name("http-demo"),
		server.Address(":8010"),
	)

	if err := srv.Handle(srv.NewHandler(beego.BeeApp.Handlers)); err != nil {
		logs.Error(err.Error())
		return
	}

	service := micro.NewService(
		micro.Server(srv),
		micro.Address(":8100"),
		micro.Registry(reg),
	)
	service.Init()
	if err := service.Run(); err != nil {
		logs.Error("init service err")
		return
	}
}
