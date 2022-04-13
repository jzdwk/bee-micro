/*
@Time : 2022/4/13
@Author : jzd
@Project: bee-micro
*/
package client

import (
	"context"
	"fmt"
	"github.com/asim/go-micro/v3/client"
	"github.com/asim/go-micro/v3/metadata"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

type traceWrapper struct {
	client.Client
}

func (tr *traceWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	var clientSpan opentracing.Span
	tracer := opentracing.GlobalTracer()
	name := fmt.Sprintf("Http Client Span: %s %s%s", req.Method(), req.Service(), req.Endpoint())
	// Find parent span.
	// First try to get span within current service boundary.
	// If there doesn't exist, try to get it from go-micro metadata(which is cross boundary)
	md, ok := metadata.FromContext(ctx)
	if !ok {
		md = make(metadata.Metadata)
	}
	var spanOpts []opentracing.StartSpanOption
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		spanOpts = append(spanOpts, opentracing.ChildOf(parentSpan.Context()))
	} else if spanCtx, err := tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(md)); err == nil {
		spanOpts = append(spanOpts, opentracing.ChildOf(spanCtx))
	}
	clientSpan = tracer.StartSpan(name, spanOpts...)
	defer clientSpan.Finish()
	// Set some tags on the clientSpan to annotate that it's the client span. The additional HTTP tags are useful for debugging purposes.
	ext.SpanKindRPCClient.Set(clientSpan)
	ext.HTTPUrl.Set(clientSpan, req.Endpoint())
	ext.HTTPMethod.Set(clientSpan, req.Method())
	// cannot Inject the client span context into the headers
	/*if err := tracer.Inject(clientSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header));err != nil{
		return err
	}*/
	return tr.Client.Call(ctx, req, rsp, opts...)
}
