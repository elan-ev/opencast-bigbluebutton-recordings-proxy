package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// server is the struct representing the http server.
type server struct {
	opencast opencast
	config   *config
	srv      *http.Server
}

func (s *server) logRequest(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf(
			"%v: Request: Remote Address=%v, Host=%v, User Agent=%v, Method=%v, URI=%v, Proto=%v.\n",
			time.Now(), r.RemoteAddr, r.Host, r.UserAgent(), r.Method, r.URL.RequestURI(), r.Proto)
		h(w, r)
	}
}

func (s *server) responseError(w http.ResponseWriter, internalErr error,
	externalErr string, code int) {
	log.Printf("%v: internal_error=%v, external_error=%v, code=%v",
		time.Now(), internalErr, externalErr, code)
	http.Error(w, externalErr, code)
}

func (s *server) proxyBBBRecordings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		recordID := r.FormValue("recordID")
		meetingID := r.FormValue("meetingID")
		var ids []string
		if recordID != "" {
			ids = strings.Split(recordID, ",")
		} else if meetingID != "" {
			ids = strings.Split(meetingID, ",")
		} else {
			s.responseError(w,
				errors.New("unable to get all recordings, this is not implemented"),
				"", http.StatusNotImplemented)
			return
		}

		opencastResults := []*opencastSearchResult{}
		for _, id := range ids {
			// Get result from opencast interface
			opencastResult, err := s.opencast.getOpencastResult(r.Context(), id)
			if err != nil {
				s.responseError(w,
					fmt.Errorf("unable to get opencast result, %w", err),
					"", http.StatusBadRequest)
				return
			}
			if opencastResult.SearchResults.Total == 0 {
				s.responseError(w, nil, "not found", http.StatusNotFound)
				return
			}

			opencastResults = append(opencastResults, opencastResult)

		}

		result, err := xml.Marshal(s.makeBBBResponse(opencastResults))
		if err != nil {
			s.responseError(w,
				fmt.Errorf("unable to marshal bbb response as xml, %w", err),
				"", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		_, err = w.Write(result)
		if err != nil {
			s.responseError(w,
				fmt.Errorf("unable to write body, %w", err),
				"", http.StatusInternalServerError)
			return
		}
	}
}
