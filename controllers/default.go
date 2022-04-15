package controllers

import (
	mybroker "bee-micro/broker"
	"bee-micro/config"
	"bee-micro/dao"
	"encoding/json"
	"fmt"
	"github.com/asim/go-micro/v3/broker"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/opentracing/opentracing-go"
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
	parentSpanCtx := c.Ctx.Request.Context().Value("parentSpanCtx")
	//操作db
	dao.WithTransaction("DeleteOneService", parentSpanCtx.(opentracing.SpanContext), func(o orm.Ormer) error {
		dao.DeleteService(o, "123")
		return nil
	})
	message := c.Ctx.Input.Param(":message")
	logs.Info("get param from uri, %s", message)
	//
	_, err := httplib.Get("http://myecs.jzd:65080/anything/get").DoRequest()
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		msg := Resp{Result: "fail", Message: fmt.Sprintf("do http bin request err, %s", err.Error())}
		c.Data["json"] = msg
		c.ServeJSON()
		return
	}
	c.Ctx.Output.SetStatus(http.StatusOK)
	conf, _ := config.GetKong()
	msg := Resp{Result: "success", Message: fmt.Sprintf("kong address from config center:[%s]", conf.Address)}
	c.Data["json"] = msg
	c.ServeJSON()
	//time.Sleep(20 * time.Second)
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
}
