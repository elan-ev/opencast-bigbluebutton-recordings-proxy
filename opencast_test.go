package main

import (
	"reflect"
	"testing"
)

func Test_opencastMediaPackageAttachments_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		b       []byte
		want    *opencastMediaPackageAttachments
		wantErr bool
	}{
		{
			name: "Empty string",
			b:    []byte(`""`),
			want: &opencastMediaPackageAttachments{
				[]opencastMediaPackageAttachment{},
			},
		},
		{
			name: "Single attachment",
			b: []byte(`{
				"attachment": {
					"type": "type",
					"url": "url"
				}}`),
			want: &opencastMediaPackageAttachments{
				Attachment: []opencastMediaPackageAttachment{
					{Type: "type", URL: "url"},
				},
			},
		},
		{
			name: "Multiple attachments",
			b: []byte(`{
				"attachment":[
					{"type": "type", "url": "url"},
					{"type": "type", "url": "url"}
				]}`),
			want: &opencastMediaPackageAttachments{
				Attachment: []opencastMediaPackageAttachment{
					{Type: "type", URL: "url"},
					{Type: "type", URL: "url"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			omas := &opencastMediaPackageAttachments{}
			if err := omas.UnmarshalJSON(tt.b); (err != nil) != tt.wantErr {
				t.Errorf("opencastMediaPackageAttachments.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(omas, tt.want) {
				t.Errorf("opencastMediaPackageAttachments.UnmarshalJSON() = %v, want %v", omas, tt.want)
			}
		})
	}
}
