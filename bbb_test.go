package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_server_verifyBBBChecksum(t *testing.T) {
	tests := []struct {
		name      string
		bbbSecret string
		request   string
		verified  bool
	}{
		{
			name:      "No security",
			bbbSecret: "",
			request:   "https://example.com/bigbluebutton/api/getRecordings?meetingID=af760103-13bb-452a-b9a2-01ea38e4edc1",
			verified:  true,
		},
		{
			name:      "Wrong checksum",
			bbbSecret: "secret",
			request:   "https://example.com/bigbluebutton/api/getRecordings?meetingID=af760103-13bb-452a-b9a2-01ea38e4edc1&checksum=this-is-definitely-wrong",
			verified:  false,
		},
		{
			name:      "Correct checksum",
			bbbSecret: "secret",
			request:   "https://example.com/bigbluebutton/api/getRecordings?meetingID=af760103-13bb-452a-b9a2-01ea38e4edc1&checksum=a26132a3d454f44628c9b171368cf11a44229af0",
			verified:  true,
		},
		{
			name:      "Correct checksum, another order",
			bbbSecret: "secret",
			request:   "https://example.com/bigbluebutton/api/getRecordings?checksum=a26132a3d454f44628c9b171368cf11a44229af0&meetingID=af760103-13bb-452a-b9a2-01ea38e4edc1",
			verified:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Server and set BBB secret
			s := &server{
				config: &config{BigBlueButton: struct {
					Secret string "yaml:\"secret\""
				}{tt.bbbSecret}},
			}

			// Create handler to call
			var handler http.HandlerFunc
			if tt.verified {
				handler = func(w http.ResponseWriter, r *http.Request) {
					t.Log("Called handler properly")
				}
			} else {
				handler = func(w http.ResponseWriter, r *http.Request) {
					t.Error("Uncalled Handler called. This should not happen.")
				}
			}
			handler = s.verifyBBBChecksum(handler)

			// Call handler with a http test server
			handler.ServeHTTP(
				httptest.NewRecorder(),
				httptest.NewRequest("GET", tt.request, nil),
			)
		})
	}
}
