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

type Interactive struct {
	ID       string               `bson:"_id,omitempty"       json:"_id,omitempty"`
	SHA      string               `bson:"sha,omitempty"       json:"sha,omitempty"`
	State    string               `bson:"state,omitempty"     json:"state,omitempty"`
	Active   *bool                `bson:"active,omitempty"    json:"active,omitempty"`
	Metadata *InteractiveMetadata `bson:"metadata,omitempty"  json:"metadata,omitempty"`
	Archive  Archive              `bson:"archive,omitempty"   json:"archive,omitempty"`
}

type Archive struct {
	Name  string  `bson:"name,omitempty"`
	Size  int64   `bson:"size_in_bytes,omitempty"`
	Files []*File `bson:"files,omitempty"`
}

type File struct {
	Name     string `bson:"name,omitempty"`
	Mimetype string `bson:"mimetype,omitempty"`
	Size     int64  `bson:"size_in_bytes,omitempty"`
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
