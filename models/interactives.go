package models

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

type Interactive struct {
	ID      string   `bson:"_id,omitempty"           json:"id,omitempty"`
	Archive *Archive `bson:"archive,omitempty"       json:"archive,omitempty"`
	//Mongo only
	SHA          string `bson:"sha,omitempty"       json:"-"`
	State        string `bson:"state,omitempty"     json:"-"`
	Active       *bool  `bson:"active,omitempty"    json:"-"`
	MetadataJson string `bson:"metadata,omitempty"  json:"-"`
	// HTTP only
	Metadata map[string]string `bson:"-"            json:"metadata,omitempty"`
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
