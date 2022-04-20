/*
@Time : 2022/4/1
@Author : jzd
@Project: bee-micro
*/
package broker

import (
	redisBroker "github.com/asim/go-micro/plugins/broker/redis/v3"
	"github.com/asim/go-micro/v3/broker"
	"github.com/astaxie/beego/logs"
)

var RedisBk broker.Broker

const BrokerTopic = "bee-micro.get.message"

func Init() error {
	//broker
	RedisBk = redisBroker.NewBroker(broker.Addrs("myecs.jzd:65079"))
	if err := RedisBk.Init(); err != nil {
		logs.Error("redis broker init error: %v", err)
		return err
	}
	if err := RedisBk.Connect(); err != nil {
		logs.Error("redis broker connect error: %v", err)
		return err
	}
	//subscribe topic
	if _, err := RedisBk.Subscribe(BrokerTopic, msgHandler); err != nil {
		logs.Error("redis broker subscribe error: %v", err)
		return err
	}
	return nil
}

func msgHandler(e broker.Event) error {
	msg := e.Message().Body
	logs.Info("get message from broker, msg %v", string(msg))
	return nil
}
