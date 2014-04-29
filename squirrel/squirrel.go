package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/exp"
)

func main() {
	var (
		listen = flag.String("listen", ":7801", "Server listen address")
	)
	flag.Parse()

	if *listen == "" {
		flag.Usage()
		os.Exit(1)
	}

	exp.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})
	exp.Handle(prometheus.ExpositionResource, prometheus.DefaultHandler)

	log.Printf("Starting server on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, exp.DefaultCoarseMux))
}
