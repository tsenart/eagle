package main

import (
	"errors"
	"fmt"
	"log"
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
}

// Result contains the result of a single HTTP request sent against
type Result struct {
	Test    string
	Code    string
	Latency float64
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

	t.layers = append(t.layers, &LoadTestLayer{name, endpoints})
	return nil
}

// Run tests all registered layers in parallel and passes the results to the
// given Result chan.
func (t *LoadTest) Run(c chan Result) {
	for {
		<-time.Tick(t.Duration)
		for _, layer := range t.layers {
			go func(l *LoadTestLayer) {
				l.test(t.Rate, t.Duration, c)
			}(layer)
		}
	}
}

func (l *LoadTestLayer) test(r uint64, d time.Duration, c chan Result) {
	// TODO(ts): Move to Register() to prevent unnecessary targets generaiton.
	var list []string
	for _, endpoint := range l.Endpoints {
		list = append(list, fmt.Sprintf("GET %s", endpoint))
	}

	// TODO(ts): Missing error handling.
	targets, _ := vegeta.NewTargets(list, nil, nil)
	results := vegeta.Attack(targets, r, d)
	for _, result := range results {
		c <- Result{
			Test:    l.Name,
			Code:    strconv.Itoa(int(result.Code)),
			Latency: float64(result.Latency.Nanoseconds()),
		}
	}

	// TODO(ts): move
	metrics := vegeta.NewMetrics(results)
	p := int(metrics.Success * float64(metrics.Requests))
	log.Printf("[%s] success: %d / %d", l.Name, p, metrics.Requests)
}
