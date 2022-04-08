/*
@Time : 2022/4/7
@Author : jzd
@Project: bee-micro
*/
package server

import (
	"bee-micro/controllers"
	"encoding/json"
	"github.com/asim/go-micro/v3/errors"
	"github.com/astaxie/beego/logs"
	"github.com/juju/ratelimit"
	"io"
	"net/http"
	"time"
)

// NewRateLimitHandlerWrapper takes a rate limiter and wait flag and returns a api  Wrapper.
func NewRateLimitHandlerWrapper(h http.Handler, b *ratelimit.Bucket, wait bool) http.Handler {
	fn := RateLimit(b, wait)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := fn(); err != nil {
			logs.Error("rate-limit err, %v", err.Error())
			rsp := new(controllers.Resp)
			rsp.Result = "fail"
			rsp.Message = "too many requests"
			retJson, _ := json.Marshal(rsp)
			io.WriteString(w, string(retJson))
			return
		}
		h.ServeHTTP(w, r)
	})
}

func RateLimit(b *ratelimit.Bucket, wait bool) func() error {
	errId := "go.micro.server"
	return func() error {
		if wait {
			time.Sleep(b.Take(1))
		} else if b.TakeAvailable(1) == 0 {
			return errors.New(errId, "too many request", 429)
		}
		return nil
	}
}
