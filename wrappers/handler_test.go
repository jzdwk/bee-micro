/*
@Time : 2022/4/8
@Author : jzd
@Project: bee-micro
*/
package wrappers

import (
	"context"
	"github.com/asim/go-micro/v3/server"
)

// HandlerFunc represents a single method of a handler. It's used primarily
// for the wrappers. What's handed to the actual method is the concrete
// request and response types.
type HandlerFunc func(ctx context.Context, req server.Request, rsp interface{}) error

// HandlerWrapper wraps the HandlerFunc and returns the equivalent
type HandlerWrapper func(HandlerFunc) HandlerFunc

// NewHandlerWrapper accepts an opentracing Tracer and returns a Handler Wrapper
func NewHandlerWrapper(sth interface{}) server.HandlerWrapper {
	return func(h server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			//do sth
			if err := h(ctx, req, rsp); err != nil {
				//do another
				return err
			}
			return nil
		}
	}
}
