package main

import (
	mybroker "bee-micro/broker"
	"bee-micro/config"
	"bee-micro/filter"
	_ "bee-micro/routers"
	"bee-micro/tracer"
	"flag"
	"fmt"
	"github.com/asim/go-micro/plugins/registry/consul/v3"
	httpServer "github.com/asim/go-micro/plugins/server/http/v3"
	tracePlugin "github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/server"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

var (
	port     = flag.String("port", "8010", "port")
	register = "myecs.jzd:65085"
	jaeger   = "myecs.jzd:65031"
)

func main() {
	//load config from consul
	if err := config.Init(); err != nil {
		logs.Error("init consul config center err, %v", err.Error())
		return
	}
	conf, err := config.GetConsul()
	if err != nil {
		logs.Error("get consul from config center err, %v", err.Error())
		return
	}
	logs.Info("read from config center, config center address %v", conf.Address)
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
		server.Broker(mybroker.RedisBk))

	//jaeger
	jaegerTracer, closer, err := tracer.NewTracer(serverName, jaeger)
	if err != nil {
		logs.Error("new jaeger tracer err, %v", err.Error())
		return
	}
	defer closer.Close()
	opentracing.SetGlobalTracer(jaegerTracer)

	rl, err := filter.NewRateLimit()
	if err != nil {
		logs.Error("new rate limit filter err, %v", err.Error())
		return
	}
	//pr := filter.NewPrometheusMonitor("prometheus", serverName)
	beego.InsertFilter("/demo/*", beego.BeforeRouter, rl.Filter, false)
	//beego.InsertFilter("/demo/*", beego.FinishRouter, pr.Filter, false)
	op := filter.Options{Name: serverName, ID: serverID, Version: serverVersion}
	beego.InsertFilter("/demo/*", beego.FinishRouter, op.Filter, false)

	//wrapper impl
	/*	apiWithRateLimit := srvWrapper.NewRateLimitHandlerWrapper(beego.BeeApp.Handlers, ratelimit.NewBucketWithRate(float64(1), int64(1)), false)
		opt := srvWrapper.Options{Name: serverName, ID: serverID, Version: serverVersion}
		apiWithMetric := srvWrapper.NewPrometheusHandlerWrapper(apiWithRateLimit, opt)*/
	if err := srv.Handle(srv.NewHandler(beego.BeeApp.Handlers)); err != nil {
		logs.Error("new http server handler err, %v", err.Error())
		return
	}
	//redis broker
	if err := mybroker.Init(); err != nil {
		logs.Error("init broker err, %v", err.Error())
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
		micro.WrapHandler(tracePlugin.NewHandlerWrapper(opentracing.GlobalTracer())),
		//logging
	)
	go PrometheusBoot()
	service.Init()
	if err := service.Run(); err != nil {
		logs.Error("init service err, %v", err.Error())
		return
	}
}

func PrometheusBoot() {
	http.Handle("/metrics", promhttp.Handler())
	conf, err := config.GetMetric()
	if err != nil {
		logs.Error("get metric config from config center err, %v", err.Error())
		return
	}
	//启动web服务，监听8085端口
	go func() {
		err := http.ListenAndServe(conf.Address, nil)
		if err != nil {
			logs.Error("listen and server on %s err, %v", conf.Address, err.Error())
		}
	}()
}
