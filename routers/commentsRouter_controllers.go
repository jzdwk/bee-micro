package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

    beego.GlobalControllerRouter["go-micro-demo/controllers:MainController"] = append(beego.GlobalControllerRouter["go-micro-demo/controllers:MainController"],
        beego.ControllerComments{
            Method: "Get",
            Router: "/:message/get",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["go-micro-demo/controllers:MainController"] = append(beego.GlobalControllerRouter["go-micro-demo/controllers:MainController"],
        beego.ControllerComments{
            Method: "Post",
            Router: "/:message/post",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}
