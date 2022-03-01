package models

type InteractiveState int

const (
	ArchiveUploaded InteractiveState = iota
	ArchiveDispatchFailed
	ArchiveDispatchedToImporter
	ImportFailure
	ImportSuccess
)

type Interactive struct {
	ID           string  `bson:"_id,omitempty"       json:"_id,omitempty"`
	SHA          string  `bson:"sha,omitempty"       json:"sha,omitempty"`
	State        string  `bson:"state,omitempty"     json:"state,omitempty"`
	Active       *bool   `bson:"active,omitempty"    json:"active,omitempty"`
	MetadataJson string  `bson:"metadata,omitempty"  json:"metadata,omitempty"`
	Archive      Archive `bson:"archive,omitempty"   json:"archive,omitempty"`
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
