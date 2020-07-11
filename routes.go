package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *server) routes() {
	router := mux.NewRouter()

	router.Path("/api/getRecordings").
		Methods(http.MethodGet).
		HandlerFunc(
			s.logRequest(s.proxyBBBRecordings()),
		)

	// Set router to srv handler
	s.srv.Handler = router
}
