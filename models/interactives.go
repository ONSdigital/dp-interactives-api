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
	ImportMessage    string      `json:"import_message,omitempty"`
	Interactive      Interactive `json:"interactive,omitempty"`
}

// Mongo/HTTP models

type InteractiveMetadata struct {
	Title             string `bson:"title"                    json:"title"`
	Label             string `bson:"label"                    json:"label"`
	InternalID        string `bson:"internal_id"              json:"internal_id"`
	CollectionID      string `bson:"collection_id,omitempty"  json:"collection_id,omitempty"`
	HumanReadableSlug string `bson:"slug,omitempty"           json:"slug,omitempty"`
	ResourceID        string `bson:"resource_id,omitempty"    json:"resource_id,omitempty"`
}

func (i *InteractiveMetadata) Update(update *InteractiveMetadata, slugGen Generator) *InteractiveMetadata {
	if update == nil {
		return i
	}
	if update.Label != "" {
		i.Label = update.Label
		i.HumanReadableSlug = slugGen(update.Label)
	}
	if update.Title != "" {
		i.Title = update.Title
	}
	if update.InternalID != "" {
		i.InternalID = update.InternalID
	}
	if update.CollectionID != "" {
		i.CollectionID = update.CollectionID
	}
	return i
}

type Interactive struct {
	ID        string               `bson:"_id,omitempty"               json:"id,omitempty"`
	Archive   *Archive             `bson:"archive,omitempty"           json:"archive,omitempty"`
	Metadata  *InteractiveMetadata `bson:"metadata,omitempty"          json:"metadata,omitempty"`
	Published *bool                `bson:"published,omitempty"         json:"published,omitempty"`
	//Mongo only
	SHA           string  `bson:"sha,omitempty"               json:"-"`
	State         string  `bson:"state,omitempty"             json:"-"`
	Active        *bool   `bson:"active,omitempty"            json:"-"`
	ImportMessage *string `bson:"import_message,omitempty"    json:"-"`
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
