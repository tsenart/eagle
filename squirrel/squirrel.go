package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	var (
		listen = flag.String("listen", ":7801", "Server listen address")

		requestDuration  = prometheus.NewCounter()
		requestDurations = prometheus.NewDefaultHistogram()
		requestTotal     = prometheus.NewCounter()
	)
	flag.Parse()

	if *listen == "" {
		flag.Usage()
		os.Exit(1)
	}

	prometheus.Register("squirrel_requests_total", "Total number of requests made", prometheus.NilLabels, requestTotal)
	prometheus.Register("squirrel_requests_duration_nanoseconds_total", "Total amount of time squirrel has spent to answer requests in nanoseconds", prometheus.NilLabels, requestDuration)
	prometheus.Register("squirrel_requests_duration_nanoseconds", "Amounts of time squirrel has spent answering requests in nanoseconds", prometheus.NilLabels, requestDurations)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer func(began time.Time, r *http.Request) {
			d := time.Since(began)
			labels := map[string]string{
				"method": strings.ToLower(r.Method),
				"path":   "/",
				"code":   strconv.Itoa(http.StatusOK),
				"target": r.Header.Get("X-eagle-target"),
			}

			requestTotal.Increment(labels)
			requestDuration.IncrementBy(labels, float64(d))
			requestDurations.Add(labels, float64(d))
		}(time.Now(), r)

		fmt.Fprint(w, "OK")
	})
	http.Handle(prometheus.ExpositionResource, prometheus.DefaultRegistry.Handler())

	log.Printf("Starting server on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, nil))
}
