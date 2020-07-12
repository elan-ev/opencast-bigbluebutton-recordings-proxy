package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func mainWithError() error {

	// Read config
	c, err := newConfig("config.yml")
	if err != nil {
		return fmt.Errorf("unable to get configuration, %w", err)
	}
	log.Println("Configuration file read")

	s := server{
		config: c,
		srv: &http.Server{
			Addr: c.Server.Address,
		},
		opencast: newOpencastInMemory(c),
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
