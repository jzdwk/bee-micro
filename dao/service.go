/*
@Time : 21-2-1
@Author : jzd
@Project: apigw
*/
package dao

import (
	"bee-micro/models"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

func CreateService(o orm.Ormer, m *models.ApiService) error {
	logs.Info("service dao: add service model %+v ", *m)
	sql := `INSERT INTO api_service 
			( id, name, create_time, update_time)
			VALUES
			( ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )`
	_, err := o.Raw(sql, m.Id, m.Name, m.CreateTime, m.UpdateTime).Exec()
	if err != nil {
		logs.Error("service dao: add service model info err %v.", err.Error())
		return err
	}
	return nil
}

func DeleteService(o orm.Ormer, id string) error {
	if _, err := o.Delete(&models.ApiService{Id: id}); err != nil {
		logs.Error("service dao: delete api service model err.%v, id: %v", err.Error(), id)
		return err
	}
	return nil
}
