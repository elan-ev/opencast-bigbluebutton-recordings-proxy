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
				Duration    int                             `json:"duration"`
				ID          string                          `json:"id"`
				Start       time.Time                       `json:"start"`
				Title       string                          `json:"title"`
				Attachments opencastMediaPackageAttachments `json:"attachments"`
			} `json:"mediapackage"`
		} `json:"result"`
	} `json:"search-results"`
}

type opencastMediaPackageAttachments struct {
	Attachment []opencastMediaPackageAttachment `json:"attachment"`
}

type opencastMediaPackageAttachment struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// UnmarshalJSON customize the unmarshaling of opencastMediaPackageAttachments
// which can be either // "", {"attachement": {...}}
// or {"attachement": [{...},{...}]}.
func (omas *opencastMediaPackageAttachments) UnmarshalJSON(b []byte) error {
	var err error

	// Try to parse as string
	var s string
	if err = json.Unmarshal(b, &s); err == nil {
		omas.Attachment = []opencastMediaPackageAttachment{}
		return nil
	}

	// Try to parse as single attachement
	v := struct {
		Attachment opencastMediaPackageAttachment `json:"attachment"`
	}{}
	if err = json.Unmarshal(b, &v); err == nil {
		omas.Attachment = []opencastMediaPackageAttachment{v.Attachment}
		return nil
	}

	// Try to parse as a list of attachments
	vs := struct {
		Attachment []opencastMediaPackageAttachment `json:"attachment"`
	}{}
	if err = json.Unmarshal(b, &vs); err == nil {
		omas.Attachment = vs.Attachment
		return nil
	}

	return err
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
