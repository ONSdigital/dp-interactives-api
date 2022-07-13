package models

import "time"

type ArchiveFile struct {
	ID            string     `bson:"_id,omitempty"            json:"id,omitempty"`
	InteractiveID string     `bson:"interactive_id,omitempty" json:"interactive_id,omitempty"`
	LastUpdated   *time.Time `bson:"last_updated,omitempty"   json:"last_updated,omitempty"`
	Name          string     `bson:"name,omitempty"           json:"name,omitempty"`
	Mimetype      string     `bson:"mimetype,omitempty"       json:"mimetype,omitempty"`
	Size          int64      `bson:"size_in_bytes,omitempty"  json:"size_in_bytes,omitempty"`
	URI           string     `bson:"uri,omitempty"            json:"uri,omitempty"`
}
