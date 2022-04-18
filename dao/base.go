package dao

import (
	"bee-micro/wrappers/server"
	"context"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/opentracing/opentracing-go"
	uuid "github.com/satori/go.uuid"
	"sync"
)

var (
	globalOrm orm.Ormer
	once      sync.Once
)

var UUID = func() string {
	return uuid.NewV4().String()
}

type sql string

// singleton init ormer ,only use for normal db operation
// if you begin transaction，please use WithTransaction
func Ormer() orm.Ormer {
	once.Do(func() {
		globalOrm = orm.NewOrm()
	})
	return globalOrm
}

// WithTransaction helper for transaction
func WithTransaction(ctx context.Context, method string, handler func(o orm.Ormer) error) (context.Context, error) {
	sp, tranCtx := opentracing.StartSpanFromContext(ctx, method)
	requestId := tranCtx.Value(server.HttpXRequestID)
	defer sp.Finish()

	o := orm.NewOrm()
	if err := o.Begin(); err != nil {
		sp.SetTag("method err", fmt.Sprintf("begin transaction failed: %v, rqeuset id %s", err, requestId))
		logs.Error("begin transaction failed: %v", err)
		return nil, err
	}
	if err := handler(o); err != nil {
		if e := o.Rollback(); e != nil {
			sp.SetTag("method err", fmt.Sprintf("rollback transaction failed: %v, request id %s", e, requestId))
			logs.Error("rollback transaction failed: %v", e)
			return nil, err
		}

		return nil, err
	}
	if err := o.Commit(); err != nil {
		sp.SetTag("method err", fmt.Sprintf("commit transaction failed: %v, request id %s", err, requestId))
		logs.Error("commit transaction failed: %v", err)
		return nil, err
	}
	sp.SetTag(method, fmt.Sprintf("success, request id %s", requestId))

	//set child span to ctx
	ctxWithSpan := opentracing.ContextWithSpan(ctx, sp)
	md := make(map[string]string)
	if err := opentracing.GlobalTracer().Inject(sp.Context(),
		opentracing.TextMap,
		opentracing.TextMapCarrier(md)); err != nil {
		logs.Error("inject span err, %s", err.Error())
	}
	return ctxWithSpan, nil
}
