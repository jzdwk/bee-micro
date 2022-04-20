package metrics

import (
	"fmt"
	"github.com/asim/go-micro/v3/logger"
	httpSnoop "github.com/felixge/httpsnoop"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
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

type metricWrapper struct {
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

func NewMetricWrapper(Name, Version, ID string) *metricWrapper {
	return &metricWrapper{Name: Name, Version: Version, ID: ID}
}

func (o *metricWrapper) Wrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		endpoint := r.RemoteAddr
		timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
			us := v * 1000000 // make microseconds
			timeCounterSummary.WithLabelValues(o.Name, o.Version, o.ID, endpoint).Observe(us)
			timeCounterHistogram.WithLabelValues(o.Name, o.Version, o.ID, endpoint).Observe(v)
		}))
		defer timer.ObserveDuration()
		m := httpSnoop.CaptureMetrics(h, w, r)
		if m.Code > http.StatusBadRequest {
			opsCounter.WithLabelValues(o.Name, o.Version, o.ID, endpoint, "failure").Inc()
		}
		opsCounter.WithLabelValues(o.Name, o.Version, o.ID, endpoint, "success").Inc()
	})
}
