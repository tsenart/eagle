package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// defaultRate defines how many requests per second are sent to each
	// endpoint.
	defaultRate uint64 = 100

	// defaultInterval defines how long an ephemeral test should get executed.
	// Note that one-off load tests aren't supported at the moment.
	defaultInterval = 1 * time.Second

	// defaultTimeout defines how long we wait for a request to return.
	defaultTimeout = 1 * time.Second
)

var (
	HeaderEndpoint = "X-eagle-endpoint"
	HeaderTarget   = "X-eagle-target"
	HeaderTest     = "X-eagle-test"

	// For Prometheus:
	namespace  = "eagle"
	labelNames = []string{"target", "code", "endpoint"}
)

func main() {
	var (
		listen  = flag.String("listen", ":7800", "Server listen address.")
		name    = flag.String("test.name", "unknown", "Name of the test to run.")
		path    = flag.String("test.path", "/", "Path to hit on the targets")
		rate    = flag.Uint64("test.rate", defaultRate, "Number of requests to send during test duration.")
		timeout = flag.Duration("test.timeout", defaultTimeout, "Time until a request is discarded")

		ts = targets{}
	)
	flag.Var(&ts, "test.target", `Target to hit by the test with the following format: -test.target="NAME:address/url"`)
	flag.Parse()

	if *listen == "" || len(ts) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	var (
		test     = newTest(*name, *path, *rate, defaultInterval, *timeout, ts)
		registry = newRegistry(prometheus.Labels{"test": test.name})
		resultc  = make(chan result)
	)

	test.run(resultc)
	go registry.collect(resultc)

	http.Handle("/metrics", prometheus.Handler())

	log.Printf("Starting server on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, nil))
}

type registry struct {
	latencies *prometheus.SummaryVec
	codes     *prometheus.CounterVec
}

func newRegistry(constLabels prometheus.Labels) *registry {
	var (
		latencies = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace:   namespace,
				Name:        "request_durations_nanoseconds",
				Help:        "The total duration of HTTP requests (nanoseconds).",
				ConstLabels: constLabels,
			},
			labelNames,
		)
		// TODO: Remove 'codes'. 'latencies' above already provides
		// 'eagle_request_durations_nanoseconds_count', which contains
		// the same value. However, rules have to be adjusted before
		// 'codes' can be removed.
		codes = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Name:        "response_codes_total",
				Help:        "The total number of requests per HTTP Status Code.",
				ConstLabels: constLabels,
			},
			labelNames,
		)
	)

	prometheus.MustRegister(latencies)
	prometheus.MustRegister(codes)

	return &registry{latencies, codes}
}

func (r *registry) collect(resultc chan result) {
	for {
		result := <-resultc
		labels := prometheus.Labels{
			"target":   result.target,
			"code":     fmt.Sprint(result.Code),
			"endpoint": result.endpoint,
		}

		r.codes.With(labels).Inc()
		r.latencies.With(labels).Observe(float64(result.Latency.Nanoseconds()))
	}
}
