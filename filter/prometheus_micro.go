/*
@Time : 2022/4/8
@Author : jzd
@Project: bee-micro
*/
package filter

import (
	"fmt"
	"github.com/asim/go-micro/v3/logger"
	"github.com/astaxie/beego/context"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// default metric prefix
	DefaultMetricPrefix = "micro_"
	// default label prefix
	DefaultLabelPrefix   = "micro_"
	opsCounter           *prometheus.CounterVec
	timeCounterSummary   *prometheus.SummaryVec
	timeCounterHistogram *prometheus.HistogramVec
)

type Options struct {
	Name    string
	Version string
	ID      string
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

func (op *Options) Filter(ctx *context.Context) {
	endpoint := ctx.Request.RemoteAddr
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		us := v * 1000000 // make microseconds
		timeCounterSummary.WithLabelValues(op.Name, op.Version, op.ID, endpoint).Observe(us)
		timeCounterHistogram.WithLabelValues(op.Name, op.Version, op.ID, endpoint).Observe(v)
	}))
	defer timer.ObserveDuration()
	if ctx.ResponseWriter.Status > 400 {
		opsCounter.WithLabelValues(op.Name, op.Version, op.ID, endpoint, "failure").Inc()
	}
	opsCounter.WithLabelValues(op.Name, op.Version, op.ID, endpoint, "success").Inc()
}
