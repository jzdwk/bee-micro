package main

import (
	"bee-micro/initial"
	micro2 "bee-micro/micro"
	mybroker "bee-micro/micro/broker"
	config2 "bee-micro/micro/config"
	"bee-micro/micro/metrics"
	ratelimit2 "bee-micro/micro/ratelimit"
	tracer2 "bee-micro/micro/tracer"
	_ "bee-micro/routers"
	"flag"
	"fmt"
	etcdv3 "github.com/asim/go-micro/plugins/registry/etcd/v3"
	httpServer "github.com/asim/go-micro/plugins/server/http/v3"
	"github.com/asim/go-micro/v3"
	_ "github.com/asim/go-micro/v3/plugins"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/server"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/google/uuid"
	"github.com/juju/ratelimit"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"time"
)

var (
	port     = flag.String("port", "8010", "port")
	register = "myecs.jzd:65379"
	jaeger   = "myecs.jzd:65031"
)

func main() {
	//load config from consul
	if err := config2.Init(); err != nil {
		logs.Error("init consul config center err, %v", err.Error())
		return
	}
	conf, err := config2.GetService()
	if err != nil {
		logs.Error("get consul from config center err, %v", err.Error())
		return
	}
	logs.Info("read from config center, config center address %v", conf.Address)
	//conf beego
	beego.BConfig.CopyRequestBody = true

	//consul
	//"github.com/asim/go-micro/plugins/registry/consul/v3"
	/*reg := consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{register}
	})*/
	//etcd
	reg := etcdv3.NewRegistry(func(op *registry.Options) {
		op.Addrs = []string{
			register,
		}
	})
	//http server
	serverName := fmt.Sprintf("http-demo")
	serverID := uuid.Must(uuid.NewUUID()).String()
	serverVersion := "v1.0"
	srv := httpServer.NewServer(
		server.Name(serverName),
		server.Address(fmt.Sprintf("localhost:%v", port)),
		server.Broker(mybroker.RedisBk))

	//rate limit
	/*rl, err := filter.NewRateLimit()
	if err != nil {
		logs.Error("new rate limit filter err, %v", err.Error())
		return
	}
	beego.InsertFilter("/demo/*", beego.BeforeRouter, rl.Filter, false)*/

	//wrapper init
	wrappers := make([]micro2.Wrapper, 0, 20)
	var apiHandler http.Handler
	apiHandler = beego.BeeApp.Handlers
	//1. rate limit
	rl, err := config2.GetRateLimit()
	if err != nil {
		logs.Error("get rate limit config from config center err, %s", err.Error())
		return
	}
	logs.Info("get rate limit from config center, value %+v", conf)
	bucket := ratelimit.NewBucketWithRate(rl.Rate, rl.Capacity)
	wrappers = append(wrappers, ratelimit2.NewRateLimitWrapper(bucket, rl.Wait))

	//2. tracer
	tr, io, err := tracer2.NewTracer("http-demo-tracing", jaeger)
	if err != nil {
		log.Fatal(err)
	}
	defer io.Close()
	opentracing.SetGlobalTracer(tr)
	wrappers = append(wrappers, tracer2.NewTracerWrapper())
	//3. metric
	wrappers = append(wrappers, metrics.NewMetricWrapper(serverName, serverVersion, serverID))
	for i := len(wrappers); i > 0; i-- {
		apiHandler = (wrappers[i-1]).Wrapper(apiHandler)
	}
	if err := srv.Handle(srv.NewHandler(apiHandler)); err != nil {
		logs.Error("new http server handler err, %v", err.Error())
		return
	}
	//redis broker
	/*if err := mybroker.Init(); err != nil {
		logs.Error("init broker err, %v", err.Error())
		return
	}*/
	//init micro service
	service := micro.NewService(
		//health check
		micro.RegisterTTL(time.Second*10),
		micro.RegisterInterval(time.Second*1000),
		//backend server
		micro.Server(srv),
		micro.Address(":8100"),
		//service registry
		micro.Registry(reg),
		//msg broker, default http broker
		//micro.Broker(mybroker.RedisBk),
		//logging
	)
	go PrometheusBoot()
	//db init
	initial.InitDb()
	//run micro
	service.Init()
	if err := service.Run(); err != nil {
		logs.Error("init service err, %v", err.Error())
		return
	}
}

func PrometheusBoot() {
	http.Handle("/metrics", promhttp.Handler())
	conf, err := config2.GetMetric()
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
