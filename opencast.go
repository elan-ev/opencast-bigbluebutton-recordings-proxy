package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
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

// opencast is the interface to the opencast service
type opencast interface {
	getOpencastResult(ctx context.Context, id string) (*opencastSearchResult, error)
}

// cacheEntry is a single entry in the cache
type cacheEntry struct {
	From   time.Time
	Result *opencastSearchResult
}

// opencastInMemory is an implementation of the opencast interface with in memory caching.
type opencastInMemory struct {
	client     *http.Client
	config     *config
	cache      map[string]*cacheEntry
	cacheMutex *sync.Mutex
}

// newOpencastInMemory is the constructor for a a new opencastInMemory struct.
func newOpencastInMemory(c *config) *opencastInMemory {
	return &opencastInMemory{
		config:     c,
		client:     &http.Client{Timeout: c.Opencast.RequestTimeout},
		cache:      make(map[string]*cacheEntry),
		cacheMutex: &sync.Mutex{},
	}
}

// getFromCache gets the result to the corresponding id from the cache, if present.
// Returns nil if nothing is in the cache.
func (o *opencastInMemory) getFromCache(id string) *opencastSearchResult {
	o.cacheMutex.Lock()
	defer o.cacheMutex.Unlock()
	ce, ok := o.cache[id]

	// Nothing found
	if !ok {
		return nil
	}

	// Check if expired
	if ce.From.Add(o.config.Opencast.CacheExpiration).Before(time.Now()) {
		return nil
	}

	return ce.Result
}

// saveToCache saves the result from Opencast in the cache.
func (o *opencastInMemory) saveToCache(id string, result *opencastSearchResult) {
	o.cacheMutex.Lock()
	defer o.cacheMutex.Unlock()
	o.cache[id] = &cacheEntry{
		From:   time.Now(),
		Result: result,
	}
}

// getOpencastResult gets the results for a specific id from the Opencast server.
// Uses the cache.
func (o *opencastInMemory) getOpencastResult(ctx context.Context, id string) (*opencastSearchResult, error) {
	// Try to get result from cache
	r := o.getFromCache(id)
	if r != nil {
		return r, nil
	}

	// Get a new result from the opencast server
	r, err := o.getNewOpencastResult(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to receive new search result for id %v from opencast server, %w", id, err)
	}

	// Save in cache
	o.saveToCache(id, r)
	return r, nil
}

// getNewOpencastResult gets the results for a specific id from the Opencast server.
// Does not use the cache.
func (o *opencastInMemory) getNewOpencastResult(ctx context.Context, id string) (*opencastSearchResult, error) {
	// Create request to server
	request, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%v/search/episode.json?id=%v", o.config.Opencast.URL, id), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create opencast request, %w", err)
	}
	if o.config.Opencast.Username != "" && o.config.Opencast.Password != "" {
		request.SetBasicAuth(o.config.Opencast.Username, o.config.Opencast.Password)
	}
	request.Header.Set("content-type", "application/json")

	// Do request
	response, err := o.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("unable to request opencast, %w", err)
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Println("unable to close request body after opencast call, %w", err)
		}
	}()

	// Marshal result
	result := &opencastSearchResult{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("unable to decode json response from opencast, %w", err)
	}

	return result, nil
}
