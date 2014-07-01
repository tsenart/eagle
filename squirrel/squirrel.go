package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	eagle "github.com/soundcloud/eagle"
)

var (
	namespace    = "squirrel"
	labelNames   = []string{"method", "path", "code", "endpoint", "target", "test"}
	eagleHeaders = map[string]string{
		"endpoint": eagle.HeaderEndpoint,
		"target":   eagle.HeaderTarget,
		"test":     eagle.HeaderTest,
	}
)

func main() {
	var (
		listen      = flag.String("listen", ":7801", "Server listen address")
		delay       = flag.Duration("delay", 0, "Delay for responses")
		logRequests = flag.Bool("log.request", false, "logs http request info as JSON to stdout")

		requestDuration = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "requests_duration_nanoseconds_total",
				Help:      "Total amount of time squirrel has spent to answer requests in nanoseconds",
			},
			labelNames,
		)
		requestDurations = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace: namespace,
				Name:      "requests_duration_nanoseconds",
				Help:      "Amounts of time squirrel has spent answering requests in nanoseconds",
			},
			labelNames,
		)
		requestTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "requests_total",
				Help:      "Total number of requests made",
			},
			labelNames,
		)
	)
	flag.Parse()

	if *listen == "" {
		flag.Usage()
		os.Exit(1)
	}

	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(requestDurations)
	prometheus.MustRegister(requestTotal)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer func(began time.Time, r *http.Request) {
			duration := float64(time.Since(began))
			labels := prometheus.Labels{
				"method": strings.ToLower(r.Method),
				"path":   r.URL.Path,
				"code":   strconv.Itoa(http.StatusOK),
			}

			for name, hdr := range eagleHeaders {
				v := r.Header.Get(hdr)
				if len(v) == 0 {
					v = "unknown"
				}

				labels[name] = v
			}

			requestTotal.With(labels).Inc()
			requestDuration.With(labels).Add(duration)
			requestDurations.With(labels).Observe(duration)

			if *logRequests {
				logRequest(r, began)
			}
		}(time.Now(), r)

		time.Sleep(*delay)
		fmt.Fprint(w, "OK")
	})
	http.Handle("/metrics", prometheus.Handler())

	log.Printf("Starting server on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, nil))
}

type logReq struct {
	Header map[string][]string `json:"header,omitempty"`
	Method string              `json:"method"`
	Path   string              `json:"path"`
	Time   int64               `json:"time"`
}

func logRequest(r *http.Request, t time.Time) {
	l := logReq{
		Header: r.Header,
		Method: r.Method,
		Path:   r.URL.Path,
		Time:   t.UnixNano(),
	}
	b, err := json.Marshal(l)
	if err != nil {
		panic(err)
	}
	log.Println(string(b))
}
