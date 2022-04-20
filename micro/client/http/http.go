// Package http provides a http client
package http

import (
	"bee-micro/util"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/asim/go-micro/v3/broker"
	"github.com/asim/go-micro/v3/client"
	"github.com/asim/go-micro/v3/codec"
	raw "github.com/asim/go-micro/v3/codec/bytes"
	errors "github.com/asim/go-micro/v3/errors"
	"github.com/asim/go-micro/v3/metadata"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/selector"
	"github.com/asim/go-micro/v3/transport"
)

type httpClient struct {
	once  sync.Once
	opts  client.Options
	trace bool
}

/*func init() {
	cmd.DefaultClients["http"] = NewClient
}
*/
func (h *httpClient) next(request client.Request, opts client.CallOptions) (selector.Next, error) {
	service := request.Service()
	// get proxy
	if prx := os.Getenv("MICRO_PROXY"); len(prx) > 0 {
		service = prx
	}

	// get proxy address
	if prx := os.Getenv("MICRO_PROXY_ADDRESS"); len(prx) > 0 {
		opts.Address = []string{prx}
	}

	// return remote address
	if len(opts.Address) > 0 {
		return func() (*registry.Node, error) {
			return &registry.Node{
				Address: opts.Address[0],
				Metadata: map[string]string{
					"protocol": "http",
				},
			}, nil
		}, nil
	}

	// only get the things that are of mucp protocol
	selectOptions := append(opts.SelectOptions, selector.WithFilter(
		selector.FilterLabel("protocol", "http"),
	))

	// get next nodes from the selector
	next, err := h.opts.Selector.Select(service, selectOptions...)
	if err != nil && err == selector.ErrNotFound {
		return nil, errors.NotFound("go.micro.client", err.Error())
	} else if err != nil {
		return nil, errors.InternalServerError("go.micro.client", err.Error())
	}

	return next, nil
}

func (h *httpClient) call(ctx context.Context, node *registry.Node, req client.Request, rsp interface{}, opts client.CallOptions) error {
	// set the address
	address := node.Address
	header := make(http.Header)
	var ctxMd metadata.Metadata
	if md, ok := metadata.FromContext(ctx); ok {
		for k, v := range md {
			header.Set(k, v)
		}
		ctxMd = md
	} else {
		ctxMd = make(metadata.Metadata)
	}

	// set timeout in nanoseconds
	header.Set("Timeout", fmt.Sprintf("%d", opts.RequestTimeout))
	// set the content type for the request
	header.Set("Content-Type", req.ContentType())

	// get codec
	cf, err := h.newHTTPCodec(req.ContentType())
	if err != nil {
		return errors.InternalServerError("go.micro.client", err.Error())
	}

	// marshal request
	b, err := cf.Marshal(req.Body())
	if err != nil {
		return errors.InternalServerError("go.micro.client", err.Error())
	}

	buf := &buffer{bytes.NewBuffer(b)}
	defer buf.Close()

	// start with / or not
	prefix := req.Endpoint()
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	rawurl := "http://" + address + prefix

	// parse rawurl
	URL, err := url.Parse(rawurl)
	if err != nil {
		return errors.InternalServerError("go.micro.client", err.Error())
	}

	//open http tracer
	var clientSpan opentracing.Span
	if h.trace {
		//
		var requestId string
		if parentId := ctx.Value(util.HttpXRequestID); parentId != nil && parentId.(string) != "" {
			requestId = parentId.(string)
			header.Set(util.HttpXRequestID, requestId)
		} else {
			requestId = util.UUID()
			header.Set(util.HttpXRequestID, requestId)
		}

		tracer := opentracing.GlobalTracer()
		name := fmt.Sprintf("Http Client Span: %s %s%s", req.Method(), req.Service(), req.Endpoint())
		var opts []opentracing.StartSpanOption
		if spanFromCtx := opentracing.SpanFromContext(ctx); spanFromCtx != nil {
			opts = append(opts, opentracing.ChildOf(spanFromCtx.Context()))
		} else if spanFromTextMap, err := tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(ctxMd)); spanFromTextMap != nil && err == nil {
			opts = append(opts, opentracing.ChildOf(spanFromTextMap))
		}
		clientSpan = tracer.StartSpan(name, opts...)
		defer clientSpan.Finish()
		// Set some tags on the clientSpan to annotate that it's the client span. The additional HTTP tags are useful for debugging purposes.
		ext.SpanKindRPCClient.Set(clientSpan)
		ext.HTTPUrl.Set(clientSpan, rawurl)
		ext.HTTPMethod.Set(clientSpan, req.Method())
		clientSpan.SetTag(util.HttpXRequestID, requestId)
		// Inject the client span context into the headers
		tracer.Inject(clientSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(header))
	}
	//new request
	hreq := &http.Request{
		Method:        req.Method(),
		URL:           URL,
		Header:        header,
		Body:          buf,
		ContentLength: int64(len(b)),
		Host:          address,
	}

	// make the request
	hrsp, err := http.DefaultClient.Do(hreq.WithContext(ctx))
	if err != nil {
		//add tracer log if necessary
		if h.trace {
			ext.Error.Set(clientSpan, true)
			//log to span
			util.TracerLogError(clientSpan, "trace finish, client trace error")
		}
		return errors.InternalServerError("go.micro.client", err.Error())
	}
	defer hrsp.Body.Close()

	// parse response
	b, err = ioutil.ReadAll(hrsp.Body)
	if err != nil {
		return errors.InternalServerError("go.micro.client", err.Error())
	}

	// unmarshal
	if err := cf.Unmarshal(b, rsp); err != nil {
		return errors.InternalServerError("go.micro.client", err.Error())
	}

	return nil
}

func (h *httpClient) stream(ctx context.Context, node *registry.Node, req client.Request, opts client.CallOptions) (client.Stream, error) {
	// set the address
	address := node.Address
	header := make(http.Header)
	if md, ok := metadata.FromContext(ctx); ok {
		for k, v := range md {
			header.Set(k, v)
		}
	}

	// set timeout in nanoseconds
	header.Set("Timeout", fmt.Sprintf("%d", opts.RequestTimeout))
	// set the content type for the request
	header.Set("Content-Type", req.ContentType())

	// get codec
	cf, err := h.newHTTPCodec(req.ContentType())
	if err != nil {
		return nil, errors.InternalServerError("go.micro.client", err.Error())
	}

	cc, err := net.Dial("tcp", address)
	if err != nil {
		return nil, errors.InternalServerError("go.micro.client", fmt.Sprintf("Error dialing: %v", err))
	}

	return &httpStream{
		address: address,
		context: ctx,
		closed:  make(chan bool),
		conn:    cc,
		codec:   cf,
		header:  header,
		reader:  bufio.NewReader(cc),
		request: req,
	}, nil
}

func (h *httpClient) newHTTPCodec(contentType string) (Codec, error) {
	if c, ok := defaultHTTPCodecs[contentType]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("Unsupported Content-Type: %s", contentType)
}

func (h *httpClient) newCodec(contentType string) (codec.NewCodec, error) {
	if c, ok := h.opts.Codecs[contentType]; ok {
		return c, nil
	}
	if cf, ok := defaultRPCCodecs[contentType]; ok {
		return cf, nil
	}
	return nil, fmt.Errorf("Unsupported Content-Type: %s", contentType)
}

func (h *httpClient) Init(opts ...client.Option) error {
	for _, o := range opts {
		o(&h.opts)
	}
	return nil
}

func (h *httpClient) Options() client.Options {
	return h.opts
}

func (h *httpClient) NewMessage(topic string, msg interface{}, opts ...client.MessageOption) client.Message {
	return newHTTPMessage(topic, msg, "application/proto", opts...)
}

func (h *httpClient) NewRequest(serviceName, api string, req interface{}, reqOpts ...client.RequestOption) client.Request {
	return newHTTPRequest(serviceName, api, req, h.opts.ContentType, reqOpts...)
}

func (h *httpClient) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	// make a copy of call opts
	callOpts := h.opts.CallOptions
	for _, opt := range opts {
		opt(&callOpts)
	}

	// get next nodes from the selector
	next, err := h.next(req, callOpts)
	if err != nil {
		return err
	}

	// check if we already have a deadline
	d, ok := ctx.Deadline()
	if !ok {
		// no deadline so we create a new one
		ctx, _ = context.WithTimeout(ctx, callOpts.RequestTimeout)
	} else {
		// got a deadline so no need to setup context
		// but we need to set the timeout we pass along
		opt := client.WithRequestTimeout(d.Sub(time.Now()))
		opt(&callOpts)
	}

	// should we noop right here?
	select {
	case <-ctx.Done():
		return errors.New("go.micro.client", fmt.Sprintf("%v", ctx.Err()), 408)
	default:
	}

	// make copy of call method
	hcall := h.call

	// wrap the call in reverse
	for i := len(callOpts.CallWrappers); i > 0; i-- {
		hcall = callOpts.CallWrappers[i-1](hcall)
	}

	// return errors.New("go.micro.client", "request timeout", 408)
	call := func(i int) error {
		// call backoff first. Someone may want an initial start delay
		t, err := callOpts.Backoff(ctx, req, i)
		if err != nil {
			return errors.InternalServerError("go.micro.client", err.Error())
		}

		// only sleep if greater than 0
		if t.Seconds() > 0 {
			time.Sleep(t)
		}

		// select next node
		node, err := next()
		if err != nil && err == selector.ErrNotFound {
			return errors.NotFound("go.micro.client", err.Error())
		} else if err != nil {
			return errors.InternalServerError("go.micro.client", err.Error())
		}

		// make the call
		err = hcall(ctx, node, req, rsp, callOpts)
		h.opts.Selector.Mark(req.Service(), node, err)
		return err
	}

	ch := make(chan error, callOpts.Retries)
	var gerr error

	for i := 0; i < callOpts.Retries; i++ {
		go func() {
			ch <- call(i)
		}()

		select {
		case <-ctx.Done():
			return errors.New("go.micro.client", fmt.Sprintf("%v", ctx.Err()), 408)
		case err := <-ch:
			// if the call succeeded lets bail early
			if err == nil {
				return nil
			}

			retry, rerr := callOpts.Retry(ctx, req, i, err)
			if rerr != nil {
				return rerr
			}

			if !retry {
				return err
			}

			gerr = err
		}
	}

	return gerr
}

func (h *httpClient) Stream(ctx context.Context, req client.Request, opts ...client.CallOption) (client.Stream, error) {
	// make a copy of call opts
	callOpts := h.opts.CallOptions
	for _, opt := range opts {
		opt(&callOpts)
	}

	// get next nodes from the selector
	next, err := h.next(req, callOpts)
	if err != nil {
		return nil, err
	}

	// check if we already have a deadline
	d, ok := ctx.Deadline()
	if !ok {
		// no deadline so we create a new one
		ctx, _ = context.WithTimeout(ctx, callOpts.RequestTimeout)
	} else {
		// got a deadline so no need to setup context
		// but we need to set the timeout we pass along
		opt := client.WithRequestTimeout(d.Sub(time.Now()))
		opt(&callOpts)
	}

	// should we noop right here?
	select {
	case <-ctx.Done():
		return nil, errors.New("go.micro.client", fmt.Sprintf("%v", ctx.Err()), 408)
	default:
	}

	call := func(i int) (client.Stream, error) {
		// call backoff first. Someone may want an initial start delay
		t, err := callOpts.Backoff(ctx, req, i)
		if err != nil {
			return nil, errors.InternalServerError("go.micro.client", err.Error())
		}

		// only sleep if greater than 0
		if t.Seconds() > 0 {
			time.Sleep(t)
		}

		node, err := next()
		if err != nil && err == selector.ErrNotFound {
			return nil, errors.NotFound("go.micro.client", err.Error())
		} else if err != nil {
			return nil, errors.InternalServerError("go.micro.client", err.Error())
		}

		stream, err := h.stream(ctx, node, req, callOpts)
		h.opts.Selector.Mark(req.Service(), node, err)
		return stream, err
	}

	type response struct {
		stream client.Stream
		err    error
	}

	ch := make(chan response, callOpts.Retries)
	var grr error

	for i := 0; i < callOpts.Retries; i++ {
		go func() {
			s, err := call(i)
			ch <- response{s, err}
		}()

		select {
		case <-ctx.Done():
			return nil, errors.New("go.micro.client", fmt.Sprintf("%v", ctx.Err()), 408)
		case rsp := <-ch:
			// if the call succeeded lets bail early
			if rsp.err == nil {
				return rsp.stream, nil
			}

			retry, rerr := callOpts.Retry(ctx, req, i, err)
			if rerr != nil {
				return nil, rerr
			}

			if !retry {
				return nil, rsp.err
			}

			grr = rsp.err
		}
	}

	return nil, grr
}

func (h *httpClient) Publish(ctx context.Context, p client.Message, opts ...client.PublishOption) error {
	options := client.PublishOptions{
		Context: context.Background(),
	}
	for _, o := range opts {
		o(&options)
	}

	md, ok := metadata.FromContext(ctx)
	if !ok {
		md = make(map[string]string)
	}
	md["Content-Type"] = p.ContentType()
	md["Micro-Topic"] = p.Topic()

	cf, err := h.newCodec(p.ContentType())
	if err != nil {
		return errors.InternalServerError("go.micro.client", err.Error())
	}

	var body []byte

	// passed in raw data
	if d, ok := p.Payload().(*raw.Frame); ok {
		body = d.Data
	} else {
		b := &buffer{bytes.NewBuffer(nil)}
		if err := cf(b).Write(&codec.Message{Type: codec.Event}, p.Payload()); err != nil {
			return errors.InternalServerError("go.micro.client", err.Error())
		}
		body = b.Bytes()
	}

	h.once.Do(func() {
		h.opts.Broker.Connect()
	})

	topic := p.Topic()

	// get proxy
	if prx := os.Getenv("MICRO_PROXY"); len(prx) > 0 {
		options.Exchange = prx
	}

	// get the exchange
	if len(options.Exchange) > 0 {
		topic = options.Exchange
	}

	return h.opts.Broker.Publish(topic, &broker.Message{
		Header: md,
		Body:   body,
	})
}

func (h *httpClient) String() string {
	return "http"
}

func newClient(trace bool, opts ...client.Option) client.Client {
	options := client.Options{
		CallOptions: client.CallOptions{
			Backoff:        client.DefaultBackoff,
			Retry:          client.DefaultRetry,
			Retries:        client.DefaultRetries,
			RequestTimeout: client.DefaultRequestTimeout,
			DialTimeout:    transport.DefaultDialTimeout,
		},
	}

	for _, o := range opts {
		o(&options)
	}

	if len(options.ContentType) == 0 {
		options.ContentType = "application/proto"
	}

	if options.Broker == nil {
		options.Broker = broker.DefaultBroker
	}

	if options.Registry == nil {
		options.Registry = registry.DefaultRegistry
	}

	if options.Selector == nil {
		options.Selector = selector.NewSelector(
			selector.Registry(options.Registry),
		)
	}

	rc := &httpClient{
		once:  sync.Once{},
		opts:  options,
		trace: trace,
	}

	c := client.Client(rc)

	// wrap in reverse
	for i := len(options.Wrappers); i > 0; i-- {
		c = options.Wrappers[i-1](c)
	}

	return c
}

func NewClient(trace bool, opts ...client.Option) client.Client {
	return newClient(trace, opts...)
}
