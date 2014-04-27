package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func ok(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK")
}

func main() {
	var (
		listen = flag.String("listen", ":7800", "Server listen address")
	)
	flag.Parse()

	if *listen == "" {
		flag.Usage()
		os.Exit(1)
	}

	http.HandleFunc("/", ok)

	log.Printf("Starting server on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, nil))
}
