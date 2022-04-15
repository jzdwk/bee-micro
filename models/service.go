/*
@Time : 21-2-1
@Author : jzd
@Project: apigw
*/
package models

import "time"

type ApiService struct {
	Id         string    `orm:"column(id);pk;type(char);size(36)" json:"id,omitempty"`
	Name       string    `orm:"column(name);type(text);unique" json:"name"`
	CreateTime time.Time `orm:"column(create_time);type(datetime);auto_now_add" json:"createTime"`
	UpdateTime time.Time `orm:"column(update_time);type(datetime);auto_now" json:"updateTime"`
}
