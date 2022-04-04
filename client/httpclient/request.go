package http

import (
	"github.com/asim/go-micro/v3/client"
	"github.com/asim/go-micro/v3/codec"
	"strings"
)

type httpRequest struct {
	service     string
	prefix      string
	method      string
	contentType string
	request     interface{}
	opts        client.RequestOptions
}

func newHTTPRequest(service, api string, request interface{}, contentType string, reqOpts ...client.RequestOption) client.Request {
	var opts client.RequestOptions
	for _, o := range reqOpts {
		o(&opts)
	}

	if len(opts.ContentType) > 0 {
		contentType = opts.ContentType
	}
	//GET:/demo
	apiInfo := strings.Split(api, ":")

	return &httpRequest{
		service:     service,
		prefix:      apiInfo[1],
		method:      apiInfo[0],
		request:     request,
		contentType: contentType,
		opts:        opts,
	}
}

func (h *httpRequest) ContentType() string {
	return h.contentType
}

func (h *httpRequest) Service() string {
	return h.service
}

func (h *httpRequest) Method() string {
	return h.method
}

func (h *httpRequest) Endpoint() string {
	return h.prefix
}

func (h *httpRequest) Codec() codec.Writer {
	return nil
}

func (h *httpRequest) Body() interface{} {
	return h.request
}

func (h *httpRequest) Stream() bool {
	return h.opts.Stream
}
