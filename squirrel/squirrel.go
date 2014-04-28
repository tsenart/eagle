package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})

	log.Printf("Starting server on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, nil))
}
