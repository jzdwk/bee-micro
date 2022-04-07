package controllers

import (
	mybroker "bee-micro/broker"
	"bee-micro/metrics"
	"encoding/json"
	"github.com/asim/go-micro/v3/broker"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"net/http"
)

type MainController struct {
	beego.Controller
}

type Message struct {
	Method  string
	Message string
}

type PostInfo struct {
	Name string
	Age  int
}

// @Title get test
// @Description get test
// @Success 200 success message
// @router /:message/get [get]
func (c *MainController) Get() {
	message := c.Ctx.Input.Param(":message")
	c.Ctx.Output.SetStatus(http.StatusOK)
	msg := Message{Method: c.Ctx.Request.Method, Message: message}
	c.Data["json"] = msg
	c.ServeJSON()
	//time.Sleep(20 * time.Second)
	//do it in beego filter
	metrics.Filter(c.Ctx)
}

// @Title post test
// @Description post test
// @Success 200 success message
// @router /:message/post [post]
func (c *MainController) Post() {
	var v PostInfo
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &v); err != nil {
		logs.Error("get body error. %v", err)
		return
	}
	message := c.Ctx.Input.Param(":message")
	//put msg to broker
	brokerMsg := &broker.Message{
		Header: nil,
		Body:   c.Ctx.Input.RequestBody,
	}
	if err := mybroker.RedisBk.Publish(mybroker.BrokerTopic, brokerMsg); err != nil {
		c.Data["json"] = "ERROR"
		c.ServeJSON()
	}
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = Message{Method: c.Ctx.Request.Method, Message: message}
	c.ServeJSON()
	//do it in beego filter
	metrics.Filter(c.Ctx)
}
