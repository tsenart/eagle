package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"
)

const (
	DefaultRatePerSecond = uint64(100)
	DefaultDuration      = 5 * time.Second
)

func main() {
	var (
		url = flag.String("url", "", "URL to run load tests against.")
	)
	flag.Parse()

	if *url == "" {
		flag.Usage()
		os.Exit(1)
	}

	list := []string{fmt.Sprintf("GET %s", *url)}
	targets, err := vegeta.NewTargets(list, nil, nil)
	if err != nil {
		log.Fatalf("invalid URL format %s: %s", *url, err)
	}

	results := vegeta.Attack(targets, DefaultRatePerSecond, DefaultDuration)
	metrics := vegeta.NewMetrics(results)

	fmt.Printf("Mean latency: %s", metrics.Latencies.Mean)
}
