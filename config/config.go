package config

import (
	"github.com/asim/go-micro/plugins/config/source/consul/v3"
	"github.com/asim/go-micro/v3/config"
	"strconv"
)

const (
	Host   = "myecs.jzd"
	Port   = 65085
	Prefix = "/micro/config"
)

var AppConfig config.Config

func Init() error {
	if conf, err := getConfig(); err != nil {
		return err
	} else {
		AppConfig = conf
	}
	return nil
}
func getConfig() (config.Config, error) {
	//添加配置中心
	// etcd key/value 模式
	// etcdsource := etcd.NewSource(
	// 	//设置配置中心地址
	// 	etcd.WithAddress(Host+":"+strconv.FormatInt(Port, 10)),
	// 	//设置前缀，不设置默认为 /micro/config
	// 	etcd.WithPrefix(Prefix),
	// 	//是否移除前缀，这里设置为true 表示可以不带前缀直接获取对应配置
	// 	etcd.StripPrefix(true),
	// )
	//consule
	consulsource := consul.NewSource(
		//设置配置中心地址
		consul.WithAddress(Host+":"+strconv.FormatInt(Port, 10)),
		//设置前缀，不设置默认为 /micro/config
		consul.WithPrefix(Prefix),
		//是否移除前缀，这里设置为true 表示可以不带前缀直接获取对应配置
		consul.StripPrefix(true),
	)
	//配置初始化
	conf, err := config.NewConfig()
	if err != nil {
		return conf, err
	}
	//加载配置
	err = conf.Load(consulsource)
	return conf, err
}

type metric struct {
	Address string `json:"address"`
}

type Consul struct {
	Address string `json:"address"`
}

type rateLimit struct {
	Rate     float64 `json:"rate" default:"100"`
	Capacity int64   `json:"capacity" default:"100"`
	Wait     bool    `json:"wait" default:"false"`
}

// 获取Consul
func GetConsul() (*Consul, error) {
	conf := &Consul{}
	//获取配置
	err := AppConfig.Get("consul").Scan(conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

// 获取RateLimit
func GetRateLimit() (*rateLimit, error) {
	conf := &rateLimit{}
	//获取配置
	err := AppConfig.Get("ratelimit").Scan(conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

// 获取RateLimit
func GetMetric() (*metric, error) {
	conf := &metric{}
	//获取配置
	err := AppConfig.Get("metric").Scan(conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
