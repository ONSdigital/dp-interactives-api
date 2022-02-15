package models

type InteractiveState int

const (
	ArchiveUploaded InteractiveState = iota
	ArchiveDispatchFailed
	ArchiveDispatchedToImporter
	ImportFailure
	ImportSuccess
	IsDeleted
)

type InteractiveUpdated struct {
	ImportStatus bool              `json:"importstatus,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type Interactive struct {
	ID           string `bson:"_id,omitempty"       json:"_id,omitempty"`
	SHA          string `bson:"sha,omitempty"       json:"sha,omitempty"`
	FileName     string `bson:"file_name,omitempty" json:"file_name,omitempty"`
	State        string `bson:"state,omitempty"     json:"state,omitempty"`
	MetadataJson string `bson:"metadata,omitempty"  json:"metadata,omitempty"`
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
	case IsDeleted:
		return "IsDeleted"
	}

	return "InteractiveStateUnknown"
}
