/*
@Time : 2022/4/8
@Author : jzd
@Project: bee-micro
*/
package filter

import (
	"bee-micro/controllers"
	config2 "bee-micro/micro/config"
	"encoding/json"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
	"github.com/juju/ratelimit"
	"io"
	"time"
)

type rateLimitConfig struct {
	b *ratelimit.Bucket
	w bool
}

func NewRateLimit() (*rateLimitConfig, error) {
	conf, err := config2.GetRateLimit()
	if err != nil {
		return nil, err
	}
	logs.Info("get rate limit from config center, value %+v", conf)
	bucket := ratelimit.NewBucketWithRate(conf.Rate, conf.Capacity)
	return &rateLimitConfig{b: bucket, w: conf.Wait}, nil
}

func (rl *rateLimitConfig) Filter(ctx *context.Context) {
	if rl.w {
		time.Sleep(rl.b.Take(1))
	} else if rl.b.TakeAvailable(1) == 0 {
		//common err handler
		rsp := new(controllers.Resp)
		rsp.Result = "fail"
		rsp.Message = "too many requests"
		retJson, _ := json.Marshal(rsp)
		//controllers.CustomAbort(ctx,http.StatusForbidden,string(retJson))
		io.WriteString(ctx.ResponseWriter, string(retJson))
	}
	return
}
