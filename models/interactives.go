package models

import (
	"fmt"
	"strings"
	"time"
)

type InteractiveState int

const (
	ArchiveUploaded InteractiveState = iota
	ArchiveDispatchFailed
	ArchiveDispatchedToImporter
	ImportFailure
	ImportSuccess
)

var (
	states = map[string]InteractiveState{
		"ArchiveUploaded":             ArchiveUploaded,
		"ArchiveDispatchFailed":       ArchiveDispatchFailed,
		"ArchiveDispatchedToImporter": ArchiveDispatchedToImporter,
		"ImportFailure":               ImportFailure,
		"ImportSuccess":               ImportSuccess,
	}
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

func ParseState(s string) (InteractiveState, bool) {
	enum, ok := states[strings.ToLower(s)]
	return enum, ok
}

// HTTP request

type PatchRequest struct {
	Attribute   string      `json:"attribute,omitempty"`
	Interactive Interactive `json:"interactive,omitempty"`
}

// If getlinkedtocollection is true, i.e visit index page through a collection
//    disregard other metadata fields and pick only collectionid
// else
//    usual filter behaviour
type InteractiveFilter struct {
	AssociateCollection bool                 `json:"associate_collection,omitempty"`
	Metadata            *InteractiveMetadata `json:"metadata,omitempty"`
}

// Mongo/HTTP models

type InteractiveMetadata struct {
	Title             string `bson:"title"                    json:"title"                      mod:"trim" validate:"required"`
	Label             string `bson:"label"                    json:"label"                      mod:"trim" validate:"required,alphanum"`
	InternalID        string `bson:"internal_id"              json:"internal_id"                mod:"trim" validate:"required,alphanum"`
	CollectionID      string `bson:"collection_id,omitempty"  json:"collection_id,omitempty"`
	HumanReadableSlug string `bson:"slug,omitempty"           json:"slug,omitempty"`
	ResourceID        string `bson:"resource_id,omitempty"    json:"resource_id,omitempty"`
}

func (i *InteractiveMetadata) Update(update *InteractiveMetadata, slugGen Generator) *InteractiveMetadata {
	if update == nil {
		return i
	}
	if i == nil {
		i = &InteractiveMetadata{ResourceID: update.ResourceID, HumanReadableSlug: update.HumanReadableSlug}
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

func (i *InteractiveMetadata) HasData() bool {
	if i.Title == "" && i.Label == "" && i.CollectionID == "" && i.InternalID == "" {
		return false
	}
	return true
}

//(i think) omitempty reuqired for all fields for update to work correctly - otherwise we overwrite incorrectly
type Interactive struct {
	ID          string               `bson:"_id,omitempty"               json:"id,omitempty"`
	Archive     *Archive             `bson:"archive,omitempty"           json:"archive,omitempty"`
	Metadata    *InteractiveMetadata `bson:"metadata,omitempty"          json:"metadata,omitempty"`
	Published   *bool                `bson:"published,omitempty"         json:"published,omitempty"`
	State       string               `bson:"state,omitempty"             json:"state,omitempty"`
	LastUpdated *time.Time           `bson:"last_updated,omitempty"      json:"last_updated,omitempty"`
	//Mongo only
	Active *bool  `bson:"active,omitempty"            json:"-"`
	SHA    string `bson:"sha,omitempty"               json:"-"`
	//JSON only
	URL string `bson:"-" json:"url,omitempty"`
}

func (i *Interactive) SetURL(domain string) {
	if i != nil && i.Metadata != nil {
		i.URL = fmt.Sprintf("%s/%s/%s-%s/embed", domain, "interactives", i.Metadata.HumanReadableSlug, i.Metadata.ResourceID)
	}
}

func (i *Interactive) CanPublish() (ok bool) {
	var state InteractiveState
	if i != nil {
		if state, ok = ParseState(i.State); !ok {
			return
		}
	}
	return state == ImportSuccess
}

type Archive struct {
	Name             string  `bson:"name,omitempty"           json:"name,omitempty"`
	Size             int64   `bson:"size_in_bytes,omitempty"  json:"size_in_bytes,omitempty"`
	Files            []*File `bson:"files,omitempty"          json:"files,omitempty"`
	ImportMessage    string  `bson:"import_message,omitempty" json:"import_message,omitempty"`
	ImportSuccessful bool    `bson:"-"                        json:"import_successful,omitempty"`
}

type File struct {
	Name     string `bson:"name,omitempty"          json:"name,omitempty"`
	Mimetype string `bson:"mimetype,omitempty"      json:"mimetype,omitempty"`
	Size     int64  `bson:"size_in_bytes,omitempty" json:"size_in_bytes,omitempty"`
}
