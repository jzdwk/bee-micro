/*
@Time : 2022/4/11
@Author : jzd
@Project: bee-micro
*/
package server

import (
	"bee-micro/util"
	"context"
	"fmt"
	opentracingMicro "github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	"github.com/astaxie/beego/logs"
	httpSnoop "github.com/felixge/httpsnoop"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"math/rand"
	"net/http"
)

var sf = 100

type tracerWrapper struct {
	spanCtx opentracing.SpanContext
	ctx     context.Context
}

func NewTracerWrapper() *tracerWrapper {
	return &tracerWrapper{}
}

// TracerWrapper tracer wrapper
func (tr *tracerWrapper) Wrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var sp opentracing.Span
		md := make(map[string]string)
		spanName := fmt.Sprintf("Http Server Span: %s %s%s", r.Method, r.Host, r.URL.Path)
		spanCtx, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		if err != nil {
			sp = opentracing.GlobalTracer().StartSpan(spanName)
		} else {
			sp = opentracing.GlobalTracer().StartSpan(spanName, opentracing.ChildOf(spanCtx))
		}
		defer sp.Finish()
		if err := opentracing.GlobalTracer().Inject(sp.Context(),
			opentracing.TextMap,
			opentracing.TextMapCarrier(md)); err != nil {
			logs.Error("inject span err, %s", err.Error())
		}
		tr.spanCtx = sp.Context()
		m := httpSnoop.CaptureMetrics(h, w, r)
		ext.HTTPMethod.Set(sp, r.Method)
		ext.HTTPUrl.Set(sp, r.URL.EscapedPath())
		ext.HTTPStatusCode.Set(sp, uint16(m.Code))
		if m.Code >= http.StatusBadRequest {
			ext.Error.Set(sp, true)
			//log to span
			util.TracerLogError(sp, "trace finish, server response error")
			return
		} else if rand.Intn(100) > sf {
			ext.SamplingPriority.Set(sp, 0)
		}
	})
}

// TracerWrapper tracer wrapper
func (tr *tracerWrapper) Wrapper2(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span, err := opentracingMicro.StartSpanFromContext(context.TODO(), opentracing.GlobalTracer(), r.URL.Path)
		if err != nil {
			logs.Error("start span from context err, %s", err.Error())
			return
		}
		tr.ctx = ctx
		defer span.Finish()
		/*		wp := &httpWp.StatusCodeTracker{ResponseWriter: w, Status: http.StatusOK}
				h.ServeHTTP(wp.WrappedResponseWriter(), r)*/
		//h.ServeHTTP(w,r)
		m := httpSnoop.CaptureMetrics(h, w, r)
		if m.Code >= http.StatusBadRequest {
			ext.Error.Set(span, true)
			//todo add logs to span
		} else if rand.Intn(100) > sf {
			ext.SamplingPriority.Set(span, 0)
		}
	})
}