package controllers

import (
	"encoding/json"
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

// @Title create api service
// @Description create api backend service
// @Success 201 success message
// @Failure 400 param parse error
// @Failure 500 server internal error
// @router /:message/get [get]
func (c *MainController) Get() {
	message := c.Ctx.Input.Param(":message")
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = Message{Method: c.Ctx.Request.Method, Message: message}
	c.ServeJSON()
}

// @Title create api service
// @Description create api backend service
// @Success 201 success message
// @Failure 400 param parse error
// @Failure 500 server internal error
// @router /:message/post [post]
func (c *MainController) Post() {
	var v PostInfo
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &v); err != nil {
		logs.Error("get body error. %v", err)
		return
	}
	message := c.Ctx.Input.Param(":message")
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = Message{Method: c.Ctx.Request.Method, Message: message}
	c.ServeJSON()
}
