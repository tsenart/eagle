package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
)

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

func (r *registry) collect(results chan Result) {
	for {
		result := <-results
		labels := map[string]string{
			"target":   result.Target,
			"code":     result.Code,
			"endpoint": result.Endpoint,
		}

		r.codes.Increment(labels)
		r.latencies.Add(labels, result.Latency)
	}
}

func loadLoadTest(path string) (*LoadTest, error) {
	conf, err := NewConfig(path)
	if err != nil {
		return &LoadTest{}, err
	}

	test, err := NewLoadTest(conf.Name)
	if err != nil {
		return &LoadTest{}, err
	}

	if conf.Rate != 0 {
		test.Rate = uint64(conf.Rate)
	}

	for name, t := range conf.Tests {
		endpoints, err := t.Endpoints()
		if err != nil {
			return &LoadTest{}, err
		}

		test.Register(name, endpoints)
	}

	return test, nil
}

func main() {
	var (
		listen = flag.String("listen", ":7800", "Server listen address")
		config = flag.String("config", "./eagle.conf", "Path to config file")
	)
	flag.Parse()

	if *listen == "" || *config == "" {
		flag.Usage()
		os.Exit(1)
	}

	test, err := loadLoadTest(*config)
	if err != nil {
		fmt.Printf("could not load config from %s: %s\n", *config, err)
		os.Exit(1)
	}

	registry := newRegistry(map[string]string{"test": test.Name})

	results := make(chan Result)
	test.Run(results)
	go registry.collect(results)

	http.HandleFunc("/metrics", registry.Handler())
	log.Printf("Starting server on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, nil))
}
