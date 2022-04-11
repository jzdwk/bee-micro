package filter

import (
	"context"
	"fmt"
	tracePlugin "github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	beeContext "github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
	"github.com/opentracing/opentracing-go"
	opentracinglog "github.com/opentracing/opentracing-go/log"
)

type tracerConfig struct {
	ot  opentracing.Tracer
	ctx context.Context
}

func NewTracerConfig(ctx context.Context, ot *opentracing.Tracer) *tracerConfig {
	return &tracerConfig{ot: *ot, ctx: ctx}
}

func (tr *tracerConfig) Filter(beeCtx *beeContext.Context) {

	//opentracing.Extract(open.HTTPHeaders, open.HTTPHeadersCarrier(r.Header))

	if tr.ot == nil {
		tr.ot = opentracing.GlobalTracer()
	}
	name := fmt.Sprintf("%s/%s", beeCtx.Request.Host, beeCtx.Request.RequestURI)
	_, span, err := tracePlugin.StartSpanFromContext(tr.ctx, tr.ot, name)
	if err != nil {
		logs.Error("tracer err, %s", err.Error())
	}

	defer span.Finish()
	if beeCtx.ResponseWriter.Status > 400 {
		span.LogFields(opentracinglog.String("error", err.Error()))
		span.SetTag("error", true)
	}
}
