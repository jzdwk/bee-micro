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
	mt := beego.NewNamespace("/metics", beego.NSInclude())
	beego.AddNamespace(ns, mt)
}
