package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func mainWithError() error {
	config := &config{
		OpencastURL: "https://develop.opencast.org",
		Username:    "admin",
		Password:    "opencast",
		Address:     "127.0.0.1:8000",
	}

	s := server{
		client: &http.Client{Timeout: 10 * time.Second},
		config: config,
		srv: &http.Server{
			Addr: config.Address,
		},
	}
	s.routes()
	return s.srv.ListenAndServe()
}

func main() {
	if err := mainWithError(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
