package controllers

import (
	mybroker "bee-micro/broker"
	"encoding/json"
	"github.com/asim/go-micro/v3/broker"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"net/http"
)

type MainController struct {
	beego.Controller
}

//test post body
type PostInfo struct {
	Name string
	Age  int
}

//common resp
type Resp struct {
	Result  string
	Message string
}

// @Title get test
// @Description get test
// @Success 200 success message
// @router /:message/get [get]
func (c *MainController) Get() {
	message := c.Ctx.Input.Param(":message")
	c.Ctx.Output.SetStatus(http.StatusOK)
	msg := Resp{Result: "success", Message: message}
	c.Data["json"] = msg
	c.ServeJSON()
	//time.Sleep(20 * time.Second)
	//do it in beego filter
	//prometheus.Filter(c.Ctx)
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
	c.Data["json"] = Resp{Result: "success", Message: message}
	c.ServeJSON()
	//do it in beego filter
	//prometheus.Filter(c.Ctx)
}
