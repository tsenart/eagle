package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"
)

const (
	// DefaultRate defines how many requests per second are sent to each
	// endpoint.
	DefaultRate = uint64(100)

	// DefaultDuration defines how long an ephemeral test should get executed.
	// Note that one-off load tests aren't supported at the moment.
	DefaultDuration = 1 * time.Second
)

// LoadTest holds the configuration of loadtest.
type LoadTest struct {
	Name     string
	Rate     uint64
	Duration time.Duration

	layers []*LoadTestLayer
}

// LoadTestLayer configures a single layer of a LoadTest.
type LoadTestLayer struct {
	Name      string
	Endpoints []string

	rate         uint64
	duration     time.Duration
	loadTestName string
}

// Result contains the result of a single HTTP request sent against
type Result struct {
	Target   string
	Code     string
	Latency  float64
	Endpoint string
}

// NewLoadTest builds a new LoadTest with the given name.
func NewLoadTest(name string) (*LoadTest, error) {
	if name == "" {
		return &LoadTest{}, errors.New("empty loadtest name")
	}

	return &LoadTest{
		Name:     name,
		Rate:     DefaultRate,
		Duration: DefaultDuration,
		layers:   []*LoadTestLayer{},
	}, nil
}

// Register adds a new layer test to the LoadTest.
func (t *LoadTest) Register(name string, endpoints []string) error {
	if name == "" {
		return errors.New("missing layer name")
	}

	if len(endpoints) == 0 {
		return errors.New("missing layer endpoints")
	}

	t.layers = append(t.layers, &LoadTestLayer{
		Name:         name,
		Endpoints:    endpoints,
		rate:         t.Rate,
		duration:     t.Duration,
		loadTestName: t.Name,
	})

	return nil
}

// Run tests all registered layers in parallel and passes the results to the
// given Result chan.
func (t *LoadTest) Run(c chan Result) {
	for _, layer := range t.layers {
		layer.test(c)
	}
}

func (l *LoadTestLayer) test(c chan Result) {
	for _, ep := range l.Endpoints {
		go func(ep string, c chan Result) {
			for {
				<-time.Tick(l.duration)
				l.attack(ep, c)
			}
		}(ep, c)
	}
}

func (l *LoadTestLayer) attack(ep string, resultc chan Result) {
	hdr := http.Header{}
	hdr.Add("X-eagle-endpoint", ep)
	hdr.Add("X-eagle-target", l.Name)
	hdr.Add("X-eagle-test", l.loadTestName)

	// TODO(ts): Missing error handling.
	targets, _ := vegeta.NewTargets([]string{fmt.Sprintf("GET %s", ep)}, nil, hdr)
	results := vegeta.Attack(targets, l.rate, l.duration)
	for _, result := range results {
		resultc <- Result{
			Target:   l.Name,
			Code:     strconv.Itoa(int(result.Code)),
			Latency:  float64(result.Latency.Nanoseconds()),
			Endpoint: ep,
		}
	}

	// TODO(ts): move
	metrics := vegeta.NewMetrics(results)
	log.Printf(
		"[%s] success: %d / %d (50th: %d 95th: %d 99th: %d)",
		l.Name,
		int(metrics.Success*float64(metrics.Requests)),
		metrics.Requests,
		metrics.Latencies.P50.Nanoseconds()/1000,
		metrics.Latencies.P95.Nanoseconds()/1000,
		metrics.Latencies.P99.Nanoseconds()/1000,
	)
}
