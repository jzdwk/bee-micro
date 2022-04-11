/*
@Time : 2022/4/11
@Author : jzd
@Project: bee-micro
*/
package server

import "net/http"

type Wrapper interface {
	Wrapper(h http.Handler) http.Handler
}
