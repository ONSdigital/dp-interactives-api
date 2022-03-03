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

// HTTP request

type InteractiveUpdate struct {
	ImportSuccessful *bool       `json:"import_successful,omitempty"`
	Interactive      Interactive `json:"interactive,omitempty"`
}

// Mongo/HTTP models

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
	ID      string   `bson:"_id,omitempty"           json:"id,omitempty"`
	Archive *Archive `bson:"archive,omitempty"       json:"archive,omitempty"`
	//Mongo only
	SHA    string `bson:"sha,omitempty"       json:"-"`
	State  string `bson:"state,omitempty"     json:"-"`
	Active *bool  `bson:"active,omitempty"    json:"-"`
	// HTTP only
	Metadata *InteractiveMetadata `bson:"-"            json:"metadata,omitempty"`
}

type Archive struct {
	Name  string  `bson:"name,omitempty"          json:"name,omitempty"`
	Size  int64   `bson:"size_in_bytes,omitempty" json:"size_in_bytes,omitempty"`
	Files []*File `bson:"files,omitempty"         json:"files,omitempty"`
}

type File struct {
	Name     string `bson:"name,omitempty"          json:"name,omitempty"`
	Mimetype string `bson:"mimetype,omitempty"      json:"mimetype,omitempty"`
	Size     int64  `bson:"size_in_bytes,omitempty" json:"size_in_bytes,omitempty"`
}
