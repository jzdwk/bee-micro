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

func GetConfig() (config.Config, error) {
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

type Config struct {
	Server Server `json:"server"`
	Consul Consul `json:"consul"`
}

type Server struct {
	Name string `json:"name"`
	Port string `json:"port"`
}

type Consul struct {
	Address string `json:"address"`
}

type MysqlConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	UserName string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// GetMysqlFromConsul 获取mysql的配置
func GetMysqlFromConsul(config config.Config, path ...string) (*MysqlConfig, error) {
	mysqlConfig := &MysqlConfig{}
	//获取配置
	err := config.Get(path...).Scan(mysqlConfig)
	if err != nil {
		return nil, err
	}
	return mysqlConfig, nil
}

// 获取ServerInfo
func GetConsul(config config.Config, path ...string) (*Consul, error) {
	conf := &Consul{}
	//获取配置
	err := config.Get(path...).Scan(conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
