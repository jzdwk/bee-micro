/*
@Time : 21-2-4
@Author : jzd
@Project: apigw
*/
package dao

import (
	"bee-micro/initial"
	"bee-micro/models"
	"testing"
	"time"
)

func TestCreateService(t *testing.T) {
	initial.InitDb()
	apiService := &models.ApiService{
		Id:         "123",
		Name:       "name1",
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	err := CreateService(Ormer(), apiService)
	if err != nil {
		t.Fatalf("failed to insert service info : %v", err)
	}
	t.Log("create test success")
}
func TestDeleteService(t *testing.T) {
	initial.InitDb()

	err := DeleteService(Ormer(), "id1")
	if err != nil {
		t.Fatalf("failed to delete service info : %v", err)
	}
	t.Log("delete success")
}
