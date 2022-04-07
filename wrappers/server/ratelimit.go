/*
@Time : 2022/4/7
@Author : jzd
@Project: bee-micro
*/
package server

import (
	"github.com/asim/go-micro/v3/errors"
	"github.com/astaxie/beego/logs"
	"github.com/juju/ratelimit"
	"net/http"
	"time"
)

// NewRateLimitHandlerWrapper takes a rate limiter and wait flag and returns a api  Wrapper.
func NewRateLimitHandlerWrapper(h http.Handler, b *ratelimit.Bucket, wait bool) http.Handler {
	fn := limit(b, wait, "go.micro.server")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := fn(); err != nil {
			logs.Error("rate-limit err, %v", err.Error())
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
		}
		h.ServeHTTP(w, r)
	})
}

func limit(b *ratelimit.Bucket, wait bool, errId string) func() error {
	return func() error {
		if wait {
			time.Sleep(b.Take(1))
		} else if b.TakeAvailable(1) == 0 {
			return errors.New(errId, "too many request", 429)
		}
		return nil
	}
}
