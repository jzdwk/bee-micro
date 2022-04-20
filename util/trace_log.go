/*
@Time : 2022/4/11
@Author : jzd
@Project: bee-micro
*/
package util

import (
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

//max memory 512MB
const maxMemory = 1024 * 1024 * 512

type fieldAdapter []log.Field

func logToTracer(span opentracing.Span, level string, msg string) {
	fa := fieldAdapter(make([]log.Field, 2))
	fa = append(fa, log.String("event", msg))
	fa = append(fa, log.String("level", level))
	span.LogFields(fa...)
}

func TracerLogError(span opentracing.Span, msg string) {
	logToTracer(span, "error", msg)
}
