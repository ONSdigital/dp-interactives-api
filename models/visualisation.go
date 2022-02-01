package models

type VisualisationState int

const (
	ArchiveUploaded VisualisationState = iota
	ArchiveDispatchFailed
	ArchiveDispatchedToImporter
	ImportFailed
	ImportSuccess
)

type Visualisation struct {
	ID       string             `bson:"_id,omitempty"       json:"id,omitempty"`
	SHA      string             `bson:"sha,omitempty"       json:"sha,omitempty"`
	FileName string             `bson:"file_name,omitempty" json:"file_name,omitempty"`
	State    VisualisationState `bson:"state,omitempty"     json:"state,omitempty"`
}
