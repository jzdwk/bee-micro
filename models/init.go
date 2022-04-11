package models

import (
	"github.com/astaxie/beego/orm"
)

func init() {
	//print sql
	orm.Debug = true
	// init orm tables
	//apimd
	orm.RegisterModel(
		new(ApiService))

}
