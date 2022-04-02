package routers

import (
	"bee-micro/controllers"
	"github.com/astaxie/beego"
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
