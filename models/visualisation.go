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
	ID       string              `bson:"_id,omitempty"       json:"id,omitempty"`
	SHA      string              `bson:"sha,omitempty"       json:"sha,omitempty"`
	FileName string              `bson:"file_name,omitempty" json:"file_name,omitempty"`
	State    *InteractiveState `bson:"state,omitempty"     json:"state,omitempty"`
}
