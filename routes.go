package main

import (
	"net/http"
)

func (s *server) routes() {
	router := http.NewServeMux()

	router.HandleFunc(
		"/bigbluebutton/api/getRecordings",
		s.logRequest(s.proxyBBBRecordings()),
	)

	// Set router to srv handler
	s.srv.Handler = router
}
