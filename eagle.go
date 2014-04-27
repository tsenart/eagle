package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	vegeta "github.com/tsenart/vegeta/lib"
)

const (
	DefaultRate     = uint64(100)
	DefaultDuration = 1 * time.Second
)

type test struct {
	targets  vegeta.Targets
	rate     uint64
	duration time.Duration
}

func newTest(endpoint string) (test, error) {
	if endpoint == "" {
		return test{}, fmt.Errorf("missing url")
	}

	list := []string{fmt.Sprintf("GET %s", endpoint)}
	targets, err := vegeta.NewTargets(list, nil, nil)
	if err != nil {
		return test{}, fmt.Errorf("invalid URL format %s: %s", endpoint, err)
	}

	return test{targets, DefaultRate, DefaultDuration}, nil
}

type registry struct {
	prometheus.Registry
	latencies prometheus.Histogram
	codes     prometheus.Counter
}

func newRegistry() *registry {
	var (
		r         = prometheus.NewRegistry()
		latencies = prometheus.NewDefaultHistogram()
		codes     = prometheus.NewCounter()
	)

	r.Register(
		"eagle_requst_durations_mircoseconds",
		"The total duration of HTTP requests (microseconds).",
		prometheus.NilLabels,
		latencies,
	)
	r.Register(
		"eagle_response_codes_total",
		"The total number of requests per HTTP Status Code",
		prometheus.NilLabels,
		codes,
	)

	return &registry{r, latencies, codes}
}

func (r *registry) collect(results vegeta.Results) {
	for _, result := range results {
		code := strconv.Itoa(int(result.Code))
		latency := float64(result.Latency.Nanoseconds()) / 1000

		r.codes.Increment(map[string]string{"code": code})
		r.latencies.Add(prometheus.NilLabels, latency)
	}
}

func main() {
	var (
		listen = flag.String("listen", ":7800", "Server listen address")
		url    = flag.String("url", "", "URL to run load tests against.")
	)
	flag.Parse()

	if *listen == "" {
		flag.Usage()
		os.Exit(1)
	}

	test, err := newTest(*url)
	if err != nil {
		fmt.Printf("%s\n\n", err.Error())
		flag.Usage()
		os.Exit(1)
	}

	registry := newRegistry()

	results := make(chan vegeta.Results)
	go func() {
		for {
			results <- vegeta.Attack(test.targets, test.rate, test.duration)
		}
	}()
	go func() {
		for {
			select {
			case r := <-results:
				registry.collect(r)
			}
		}
	}()

	http.HandleFunc("/metrics", registry.Handler())
	log.Printf("Starting server on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, nil))
}
