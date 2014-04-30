package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"
)

// result contains the result of a single HTTP request sent against an
// endpoint.
type result struct {
	target   string
	endpoint string

	vegeta.Result
}

// test represents a load test.
type test struct {
	name     string
	path     string
	rate     uint64
	interval time.Duration
	targets  targets
}

// run tests all registered targets in parallel and passes the results to the
// given Result chan.
func (t *test) run(c chan result) {
	for _, target := range t.targets {
		target.test(t, c)
	}
}

// newTest returns a Test object.
func newTest(name, path string, rate uint64, i time.Duration, ts targets) *test {
	return &test{
		name:     name,
		path:     path,
		rate:     rate,
		interval: i,
		targets:  ts,
	}
}

// target represents a set of endpoints to test.
type target struct {
	name      string
	endpoints []string
}

// attack uses vegeta to do a set of http requests against the given endpoint.
func (t *target) attack(test *test, ep string, resultc chan result) {
	hdr := http.Header{}
	hdr.Add(HeaderEndpoint, ep)
	hdr.Add(HeaderTarget, t.name)
	hdr.Add(HeaderTest, test.name)

	// TODO(xla): Avoid panic.
	targets, err := vegeta.NewTargets([]string{fmt.Sprintf("GET %s", ep)}, nil, hdr)
	if err != nil {
		panic(err)
	}

	results := vegeta.Attack(targets, test.rate, test.interval)
	for _, r := range results {
		resultc <- result{
			target:   t.name,
			endpoint: ep,
			Result:   r,
		}
	}

	// TODO(ts): move
	metrics := vegeta.NewMetrics(results)
	log.Printf(
		"[%s/%s] success: %d / %d (50th: %d 95th: %d 99th: %d)",
		t.name,
		ep,
		int(metrics.Success*float64(metrics.Requests)),
		metrics.Requests,
		metrics.Latencies.P50.Nanoseconds()/1000,
		metrics.Latencies.P95.Nanoseconds()/1000,
		metrics.Latencies.P99.Nanoseconds()/1000,
	)
}

// test iterates over all endpoints and spawns an attack routine for each.
func (t *target) test(test *test, c chan result) {
	for _, ep := range t.endpoints {
		go func(ep string, c chan result) {
			for {
				t.attack(test, ep, c)
			}
		}(ep, c)
	}
}

// newTarget parses the given address and tries to construct endpoints from it.
func newTarget(name, addr string) (target, error) {
	t := target{
		name:      name,
		endpoints: []string{},
	}

	u, err := url.Parse(addr)
	if err == nil && u.Scheme != "" {
		t.endpoints = append(t.endpoints, u.String())
	} else {
		_, addrs, err := net.LookupSRV("", "", addr)
		if err != nil {
			return t, err
		}

		for _, a := range addrs {
			host := strings.Trim(a.Target, ".")
			e := fmt.Sprintf("http://%s:%d/", host, a.Port)
			t.endpoints = append(t.endpoints, e)
		}
	}

	if len(t.endpoints) == 0 {
		return t, fmt.Errorf("no endpoints for target '%s'", name)
	}

	return t, nil
}

type targets map[string]target

func (t targets) Set(s string) error {
	sp := strings.SplitN(s, ":", 2)

	if len(sp) < 2 {
		return fmt.Errorf("invalid target format")
	}

	target, err := newTarget(sp[0], sp[1])
	if err != nil {
		return err
	}

	t[target.name] = target

	return nil
}

func (t targets) String() string {
	s := []string{}
	for _, target := range t {
		ts := fmt.Sprintf("%s: %s", target.name, strings.Join(target.endpoints, ", "))
		s = append(s, ts)
	}
	return strings.Join(s, "\n")
}
