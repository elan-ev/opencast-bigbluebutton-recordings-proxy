package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type opencastSearchResult struct {
	SearchResults struct {
		Total  int `json:"total"`
		Result struct {
			Mediapackage struct {
				Duration    int       `json:"duration"`
				ID          string    `json:"id"`
				Start       time.Time `json:"start"`
				Title       string    `json:"title"`
				Attachments struct {
					Attachment []struct {
						Type string `json:"type"`
						URL  string `json:"url"`
					} `json:"attachment"`
				} `json:"attachments"`
			} `json:"mediapackage"`
		} `json:"result"`
	} `json:"search-results"`
}

func (s *server) getOpencastResult(ctx context.Context, id string) (*opencastSearchResult, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%v/search/episode.json?id=%v", s.config.OpencastURL, id), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create opencast request, %w", err)
	}

	if s.config.Username != "" && s.config.Password != "" {
		request.SetBasicAuth(s.config.Username, s.config.Password)
	}

	request.Header.Set("content-type", "application/json")

	response, err := s.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("unable to request opencast, %w", err)
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Println("unable to close request body after opencast call, %w", err)
		}
	}()
	result := &opencastSearchResult{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("unable to decode json response from opencast, %w", err)
	}
	return result, nil
}
