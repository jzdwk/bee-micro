package main

import (
	mybroker "bee-micro/broker"
	"bee-micro/config"
	_ "bee-micro/routers"
	srvWrapper "bee-micro/wrappers/server"
	"flag"
	"fmt"
	"github.com/asim/go-micro/plugins/registry/consul/v3"
	httpServer "github.com/asim/go-micro/plugins/server/http/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/server"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/google/uuid"
	"github.com/juju/ratelimit"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"time"
)

var port = flag.String("port", "8010", "port")
var register = "myecs.jzd:65085"

func main() {
	//load config from consul
	cfg, _ := config.GetConfig()
	conf, err := config.GetConsul(cfg, "consul")
	if err != nil {
		return
	}
	logs.Info("read from config center, value %+v", conf)
	//conf beego
	beego.BConfig.CopyRequestBody = true
	//consul
	reg := consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{register}
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
		/*		server.WrapHandler(promwrapper.NewHandlerWrapper(
				promwrapper.ServiceName(serverName),
				promwrapper.ServiceVersion(serverVersion),
				promwrapper.ServiceID(serverID)))*/)

	//rate limit
	apiWithRateLimit := srvWrapper.NewRateLimitHandlerWrapper(beego.BeeApp.Handlers, ratelimit.NewBucketWithRate(float64(1), int64(1)), false)
	opt := srvWrapper.Options{Name: serverName, ID: serverID, Version: serverVersion}
	apiWithMetric := srvWrapper.NewPrometheusHandlerWrapper(apiWithRateLimit, opt)
	//metric
	if err := srv.Handle(srv.NewHandler(apiWithMetric)); err != nil {
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
		//tracing
		//logging
	)
	go PrometheusBoot()
	service.Init()
	if err := service.Run(); err != nil {
		logs.Error("init service err. err %v", err.Error())
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
