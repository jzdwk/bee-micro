/*
@Time : 2022/4/18
@Author : jzd
@Project: bee-micro
*/
package util

import uuid "github.com/satori/go.uuid"

var UUID = func() string {
	return uuid.NewV4().String()
}
