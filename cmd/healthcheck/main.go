package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

func main() {
	healthcheck := flag.Bool("healthcheck", false, "perform healthcheck and exit")
	flag.Parse()

	if *healthcheck {
		port := os.Getenv("HTTP_PORT")
		if port == "" {
			port = "4000"
		}
		resp, err := http.Get(fmt.Sprintf("http://localhost:%s/health", port))
		if err != nil {
			os.Exit(1)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		return
	}
}
