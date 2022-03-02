package models

import "time"

type InteractiveState int

const (
	ArchiveUploaded InteractiveState = iota
	ArchiveDispatchFailed
	ArchiveDispatchedToImporter
	ImportFailure
	ImportSuccess
)

type InteractiveMetadata struct { // TODO : Geography
	Title           string    `bson:"title"         	 		 json:"title"`
	PrimaryTopic    string    `bson:"primary_topic" 	 		 json:"primary_topic"`
	Topics          []string  `bson:"topics"        			 json:"topics"`
	Surveys         []string  `bson:"surveys"          			 json:"surveys"`
	ReleaseDate     time.Time `bson:"release_date"     			 json:"release_date"`
	Uri             string    `bson:"uri"              			 json:"uri"`
	Edition         string    `bson:"edition,omitempty"          json:"edition,omitempty"`
	Keywords        []string  `bson:"keywords,omitempty"         json:"keywords,omitempty"`
	MetaDescription string    `bson:"meta_description,omitempty" json:"meta_description,omitempty"`
	Source          string    `bson:"source,omitempty"           json:"source,omitempty"`
	Summary         string    `bson:"summary,omitempty"          json:"summary,omitempty"`
}

type InteractiveInfo struct {
	ID       string              `json:"id,omitempty"`
	Metadata InteractiveMetadata `json:"metadata,omitempty"`
}

type InteractiveUpdated struct {
	ImportStatus bool                `json:"importstatus,omitempty"`
	Metadata     InteractiveMetadata `json:"metadata,omitempty"`
}

type Interactive struct {
	ID           string               `bson:"_id,omitempty"       json:"_id,omitempty"`
	SHA          string               `bson:"sha,omitempty"       json:"sha,omitempty"`
	FileName     string               `bson:"file_name,omitempty" json:"file_name,omitempty"`
	State        string               `bson:"state,omitempty"     json:"state,omitempty"`
	Active       *bool                `bson:"active,omitempty"    json:"active,omitempty"`
	Metadata     *InteractiveMetadata `bson:"metadata,omitempty"  json:"metadata,omitempty"`
}

func (s InteractiveState) String() string {
	switch s {
	case ArchiveUploaded:
		return "ArchiveUploaded"
	case ArchiveDispatchFailed:
		return "ArchiveDispatchFailed"
	case ArchiveDispatchedToImporter:
		return "ArchiveDispatchedToImporter"
	case ImportFailure:
		return "ImportFailure"
	case ImportSuccess:
		return "ImportSuccess"
	}

	return "InteractiveStateUnknown"
}
