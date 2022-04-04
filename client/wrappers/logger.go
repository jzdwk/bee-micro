package wrappers

import (
	"context"
	"fmt"

	"github.com/asim/go-micro/v3/client"
)

// log wrapper logs every time a request is made
type logWrapper struct {
	client.Client
}

func (l *logWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	fmt.Printf("[wrapper] client request service: %s method: %s\n", req.Service(), req.Endpoint())
	return l.Client.Call(ctx, req, rsp)
}

// Implements client.Wrapper as logWrapper
func NewLogWrap(c client.Client) client.Client {
	return &logWrapper{c}
}
