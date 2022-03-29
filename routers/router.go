package routers

import (
	"github.com/astaxie/beego"
	"go-micro-demo/controllers"
)

func init() {
	//route config
	ns := beego.NewNamespace("/demo",
		beego.NSNamespace("/hello",
			beego.NSInclude(
				&controllers.MainController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
