package main

import (
	"crypto/sha1"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

const checksumError = `<response>
<returncode>FAILED</returncode>
<messageKey>checksumError</messageKey>
<message>You did not pass the checksum security check</message>
</response>`

type bbbRecordingsResponse struct {
	XMLName    xml.Name      `xml:"response"`
	Text       string        `xml:",chardata"`
	Returncode string        `xml:"returncode"`
	Recordings bbbRecordings `xml:"recordings"`
}

type bbbRecordings struct {
	Text      string         `xml:",chardata"`
	Recording []bbbRecording `xml:"recording"`
}

type bbbRecording struct {
	Text              string      `xml:",chardata"`
	RecordID          string      `xml:"recordID"`
	MeetingID         string      `xml:"meetingID"`
	InternalMeetingID string      `xml:"internalMeetingID"`
	Name              string      `xml:"name"`
	IsBreakout        bool        `xml:"isBreakout"`
	Published         bool        `xml:"published"`
	State             string      `xml:"state"`
	StartTime         int64       `xml:"startTime"`
	EndTime           int64       `xml:"endTime"`
	Participants      int         `xml:"participants"`
	Metadata          bbbMetadata `xml:"metadata"`
	Playback          bbbPlayback `xml:"playback"`
}

type bbbMetadata struct {
	Text        string `xml:",chardata"`
	IsBreakout  bool   `xml:"isBreakout"`
	MeetingName string `xml:"meetingName"`
	GlListed    bool   `xml:"gl-listed"`
	MeetingID   string `xml:"meetingId"`
}

type bbbPlayback struct {
	Text   string      `xml:",chardata"`
	Format []bbbFormat `xml:"format"`
}

type bbbFormat struct {
	Text           string     `xml:",chardata"`
	Type           string     `xml:"type"`
	URL            string     `xml:"url"`
	ProcessingTime int        `xml:"processingTime"`
	Length         int        `xml:"length"`
	Preview        bbbPreview `xml:"preview"`
}

type bbbPreview struct {
	Text   string    `xml:",chardata"`
	Images bbbImages `xml:"images"`
}

type bbbImages struct {
	Text  string     `xml:",chardata"`
	Image []bbbImage `xml:"image"`
}

type bbbImage struct {
	Text   string `xml:",chardata"`
	Alt    string `xml:"alt,attr"`
	Height string `xml:"height,attr"`
	Width  string `xml:"width,attr"`
}

func (s *server) makeBBBRecording(r *opencastSearchResult) bbbRecording {

	// Gather all preview images with postfix "/player+preview"
	images := []bbbImage{}
	regex := regexp.MustCompile(`.+/player\+preview`)
	for _, attachement := range r.SearchResults.Result.Mediapackage.Attachments.Attachment {
		if regex.Match([]byte(attachement.Type)) {
			images = append(images, bbbImage{Text: attachement.URL})
		}
	}

	return bbbRecording{
		RecordID:          r.SearchResults.Result.Mediapackage.ID,
		MeetingID:         r.SearchResults.Result.Mediapackage.ID,
		InternalMeetingID: r.SearchResults.Result.Mediapackage.ID,
		Name:              r.SearchResults.Result.Mediapackage.Title,
		IsBreakout:        false,
		Published:         false,
		State:             "published",
		StartTime:         r.SearchResults.Result.Mediapackage.Start.Unix(),
		EndTime:           r.SearchResults.Result.Mediapackage.Start.Unix() + int64(r.SearchResults.Result.Mediapackage.Duration),
		Participants:      3,
		Metadata: bbbMetadata{
			IsBreakout:  false,
			GlListed:    false,
			MeetingName: r.SearchResults.Result.Mediapackage.Title,
			MeetingID:   r.SearchResults.Result.Mediapackage.ID,
		},
		Playback: bbbPlayback{
			Format: []bbbFormat{
				{
					Type:    "opencast",
					URL:     fmt.Sprintf("%v/play/%v", s.config.Opencast.URL, r.SearchResults.Result.Mediapackage.ID),
					Preview: bbbPreview{Images: bbbImages{Image: images}},
				},
			},
		},
	}
}

// makeBBBResponse creates a response like the BigBluebutton API endpoint /api/getRecordings,
// see https://docs.bigbluebutton.org/dev/api.html#getrecordings.
func (s *server) makeBBBResponse(rs []*opencastSearchResult) *bbbRecordingsResponse {

	// Get recordings
	recordings := []bbbRecording{}
	for _, r := range rs {
		recordings = append(recordings, s.makeBBBRecording(r))

	}

	// Create Response
	return &bbbRecordingsResponse{
		Returncode: "SUCCESS",
		Recordings: bbbRecordings{Recording: recordings},
	}
}

// verifyBBBChecksum is a middleware to verify the checksum.
// For more details about what has to be done here, take a look at
// https://docs.bigbluebutton.org/dev/api.html#api-security-model.
func (s *server) verifyBBBChecksum(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// If the secret is empty, disable checksum verification
		if s.config.BigBlueButton.Secret == "" {
			h(w, r)
			return
		}

		// Create sha1
		hash := sha1.New()

		// Add api part to hash
		splittedURL := strings.Split(r.URL.Path, "/")
		if len(splittedURL) < 1 {
			s.responseError(w, fmt.Errorf("unable to split url path, %v", r.URL.Path),
				"", http.StatusInternalServerError)
			return
		}
		_, err := io.WriteString(hash, splittedURL[len(splittedURL)-1])
		if err != nil {
			s.responseError(w, fmt.Errorf("unable to add API part to sha1 hash, %v", err),
				"", http.StatusInternalServerError)
			return
		}

		// Add query without checksum to hash
		reg := regexp.MustCompile("(&checksum=[a-z0-9]+$|checksum=[a-z0-9]+&)")
		_, err = io.WriteString(hash, reg.ReplaceAllString(r.URL.RawQuery, ""))
		if err != nil {
			s.responseError(w, fmt.Errorf("unable to add query part to sha1 hash, %v", err),
				"", http.StatusInternalServerError)
			return
		}

		// Get BBB secret
		_, err = io.WriteString(hash, s.config.BigBlueButton.Secret)
		if err != nil {
			s.responseError(w, fmt.Errorf("unable to add secret part to sha1 hash, %v", err),
				"", http.StatusInternalServerError)
			return
		}

		// Compare hashes
		wanted := fmt.Sprintf("%x", string(hash.Sum(nil)))
		got := r.URL.Query().Get("checksum")
		if wanted != got {
			w.Header().Set("Content-Type", "application/xml")
			s.responseError(w, fmt.Errorf("wrong hash, wanted='%v', got=%v", wanted, got),
				checksumError,
				http.StatusBadRequest)
			return
		}

		// Checksum is verified properly, call the next handler.
		h(w, r)
	}

}
