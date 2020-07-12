package main

import (
	"encoding/xml"
	"fmt"
	"regexp"
)

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

// makeBBBResponse creates a response like the BigBluebutton API endpoint /api/getRecordings,
// see https://docs.bigbluebutton.org/dev/api.html#getrecordings.
func (s *server) makeBBBResponse(r *opencastSearchResult) *bbbRecordingsResponse {

	// Gather all preview images with postfix "/player+preview"
	images := []bbbImage{}
	regex := regexp.MustCompile(`.+/player\+preview`)
	for _, attachement := range r.SearchResults.Result.Mediapackage.Attachments.Attachment {
		if regex.Match([]byte(attachement.Type)) {
			images = append(images, bbbImage{Text: attachement.URL})
		}
	}

	// Create Response
	return &bbbRecordingsResponse{
		Returncode: "SUCCESS",
		Recordings: bbbRecordings{
			Recording: []bbbRecording{
				{
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
				},
			},
		},
	}

}
