package models

type InteractiveState int

const (
	ArchiveUploaded InteractiveState = iota
	ArchiveDispatchFailed
	ArchiveDispatchedToImporter
	ImportFailed
	ImportSuccess
	IsDeleted
)

type Interactive struct {
	ID           string `bson:"_id,omitempty"       json:"id,omitempty"`
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
	case ImportFailed:
		return "ImportFailed"
	case ImportSuccess:
		return "ImportSuccess"
	case IsDeleted:
		return "IsDeleted"
	}

	return "InteractiveStateUnknown"
}
