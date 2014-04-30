package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
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
)

var (
	HeaderEndpoint = "X-eagle-endpoint"
	HeaderTarget   = "X-eagle-target"
	HeaderTest     = "X-eagle-test"
)

func main() {
	var (
		listen = flag.String("listen", ":7800", "Server listen address.")
		name   = flag.String("test.name", "unknown", "Name of the test to run.")
		path   = flag.String("test.path", "/", "Path to hit on the targets")
		rate   = flag.Uint64("test.rate", defaultRate, "Number of requests to send during test duration.")

		ts = targets{}
	)
	flag.Var(&ts, "test.target", `Target to hit by the test with the following format: -test.target="NAME:address/url"`)
	flag.Parse()

	if *listen == "" {
		flag.Usage()
		os.Exit(1)
	}

	var (
		test     = newTest(*name, *path, *rate, defaultInterval, ts)
		registry = newRegistry(map[string]string{"test": test.name})
		resultc  = make(chan result)
	)

	test.run(resultc)
	go registry.collect(resultc)

	http.HandleFunc("/metrics", registry.Handler())

	log.Printf("Starting server on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, nil))
}

type registry struct {
	prometheus.Registry
	latencies prometheus.Histogram
	codes     prometheus.Counter
}

func newRegistry(baseLabels map[string]string) *registry {
	var (
		r         = prometheus.NewRegistry()
		latencies = prometheus.NewDefaultHistogram()
		codes     = prometheus.NewCounter()
	)

	r.Register(
		"eagle_request_durations_nanoseconds",
		"The total duration of HTTP requests (nanoseconds).",
		baseLabels,
		latencies,
	)
	r.Register(
		"eagle_response_codes_total",
		"The total number of requests per HTTP Status Code",
		baseLabels,
		codes,
	)

	return &registry{r, latencies, codes}
}

func (r *registry) collect(resultc chan result) {
	for {
		result := <-resultc
		labels := map[string]string{
			"target":   result.target,
			"code":     strconv.Itoa(int(result.Code)),
			"endpoint": result.endpoint,
		}

		r.codes.Increment(labels)
		r.latencies.Add(labels, float64(result.Latency.Nanoseconds()))
	}
}
