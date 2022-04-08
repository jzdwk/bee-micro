package server

import (
	"fmt"
	"github.com/asim/go-micro/v3/client"
	"github.com/asim/go-micro/v3/logger"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
)

var (
	// default metric prefix
	DefaultMetricPrefix = "micro_"
	// default label prefix
	DefaultLabelPrefix = "micro_"

	opsCounter           *prometheus.CounterVec
	timeCounterSummary   *prometheus.SummaryVec
	timeCounterHistogram *prometheus.HistogramVec
)

type Options struct {
	Name    string
	Version string
	ID      string
}

type Option func(*Options)

func ServiceName(name string) Option {
	return func(opts *Options) {
		opts.Name = name
	}
}

func ServiceVersion(version string) Option {
	return func(opts *Options) {
		opts.Version = version
	}
}

func ServiceID(id string) Option {
	return func(opts *Options) {
		opts.ID = id
	}
}

func init() {

	if opsCounter == nil {
		opsCounter = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%srequest_total", DefaultMetricPrefix),
				Help: "Requests processed, partitioned by endpoint and status",
			},
			[]string{
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "name"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "version"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "id"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "endpoint"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "status"),
			},
		)
	}

	if timeCounterSummary == nil {
		timeCounterSummary = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: fmt.Sprintf("%slatency_microseconds", DefaultMetricPrefix),
				Help: "Request latencies in microseconds, partitioned by endpoint",
			},
			[]string{
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "name"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "version"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "id"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "endpoint"),
			},
		)
	}

	if timeCounterHistogram == nil {
		timeCounterHistogram = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: fmt.Sprintf("%srequest_duration_seconds", DefaultMetricPrefix),
				Help: "Request time in seconds, partitioned by endpoint",
			},
			[]string{
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "name"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "version"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "id"),
				fmt.Sprintf("%s%s", DefaultLabelPrefix, "endpoint"),
			},
		)
	}

	for _, collector := range []prometheus.Collector{opsCounter, timeCounterSummary, timeCounterHistogram} {
		if err := prometheus.DefaultRegisterer.Register(collector); err != nil {
			// if already registered, skip fatal
			if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
				logger.Fatal(err)
			}
		}
	}

}

type wrapper struct {
	options  Options
	callFunc client.CallFunc
	client.Client
}

func NewPrometheusHandlerWrapper(h http.Handler, options Options) http.Handler {
	w := &wrapper{
		options: options,
	}
	return http.HandlerFunc(func(rsp http.ResponseWriter, req *http.Request) {
		endpoint := req.RemoteAddr
		timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
			us := v * 1000000 // make microseconds
			timeCounterSummary.WithLabelValues(w.options.Name, w.options.Version, w.options.ID, endpoint).Observe(us)
			timeCounterHistogram.WithLabelValues(w.options.Name, w.options.Version, w.options.ID, endpoint).Observe(v)
		}))
		defer timer.ObserveDuration()
		h.ServeHTTP(rsp, req)
		//can not get rsp message
		opsCounter.WithLabelValues(w.options.Name, w.options.Version, w.options.ID, endpoint, "success").Inc()
		//opsCounter.WithLabelValues(w.options.Name, w.options.Version, w.options.ID, endpoint, "failure").Inc()
	})
}
