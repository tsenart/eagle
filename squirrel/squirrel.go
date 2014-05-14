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

var eagleHeaders = map[string]string{
	"endpoint": eagle.HeaderEndpoint,
	"target":   eagle.HeaderTarget,
	"test":     eagle.HeaderTest,
}

func main() {
	var (
		listen      = flag.String("listen", ":7801", "Server listen address")
		delay       = flag.Duration("delay", 0, "Delay for responses")
		logRequests = flag.Bool("log.request", false, "logs http request info as JSON to stdout")

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

			requestTotal.Increment(labels)
			requestDuration.IncrementBy(labels, float64(d))
			requestDurations.Add(labels, float64(d))

			if *logRequests {
				logRequest(r, began)
			}
		}(time.Now(), r)

		time.Sleep(*delay)
		fmt.Fprint(w, "OK")
	})
	http.Handle(prometheus.ExpositionResource, prometheus.DefaultRegistry.Handler())

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
