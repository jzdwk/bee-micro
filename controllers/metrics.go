/*
@Time : 2022/4/2
@Author : jzd
@Project: bee-micro
*/
package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

type MetricsController struct {
	beego.Controller
}

// @Title post metrics data
// @Description post metrics data
// @Success 200 success message
// @router / [post]
func (c *MetricsController) Post() {
	var v interface{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &v); err != nil {
		logs.Error("get body error. %v", err)
		return
	}
	logs.Info(v)
}
