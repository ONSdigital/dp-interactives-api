package models

import (
	"fmt"
	"github.com/ONSdigital/dp-interactives-api/internal/data"
	"strings"
	"time"
)

type State int

const (
	ArchiveUploaded State = iota
	ArchiveDispatchFailed
	ArchiveDispatchedToImporter
	ImportFailure
	ImportSuccess
)

var (
	states = map[string]State{
		"archiveuploaded":             ArchiveUploaded,
		"archivedispatchfailed":       ArchiveDispatchFailed,
		"archivedispatchedtoimporter": ArchiveDispatchedToImporter,
		"importfailure":               ImportFailure,
		"importsuccess":               ImportSuccess,
	}
)

func (s State) String() string {
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
	default:
		return "Unknown"
	}
}

func ParseState(s string) (State, bool) {
	enum, ok := states[strings.ToLower(s)]
	return enum, ok
}

// If getlinkedtocollection is true, i.e visit index page through a collection
//    disregard other metadata fields and pick only collectionid
// else
//    usual filter behaviour
type Filter struct {
	AssociateCollection bool      `json:"associate_collection,omitempty"`
	Metadata            *Metadata `json:"metadata,omitempty"`
}

type Metadata struct {
	Title             string `bson:"title"                    json:"title"                      mod:"trim" validate:"required"`
	Label             string `bson:"label"                    json:"label"                      mod:"trim" validate:"required,alphanum"`
	InternalID        string `bson:"internal_id"              json:"internal_id"                mod:"trim" validate:"required,alphanum"`
	CollectionID      string `bson:"collection_id,omitempty"  json:"collection_id,omitempty"`
	HumanReadableSlug string `bson:"slug,omitempty"           json:"slug,omitempty"`
	ResourceID        string `bson:"resource_id,omitempty"    json:"resource_id,omitempty"`
}

func (i *Metadata) Update(update *Metadata, slugGen data.Generator) *Metadata {
	if update == nil {
		return i
	}
	if i == nil {
		i = &Metadata{ResourceID: update.ResourceID, HumanReadableSlug: update.HumanReadableSlug}
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

//(i think) omitempty reuqired for all fields for update to work correctly - otherwise we overwrite incorrectly
type Interactive struct {
	ID          string      `bson:"_id,omitempty"               json:"id,omitempty"`
	Archive     *Archive    `bson:"archive,omitempty"           json:"archive,omitempty"`
	Metadata    *Metadata   `bson:"metadata,omitempty"          json:"metadata,omitempty"`
	Published   *bool       `bson:"published,omitempty"         json:"published,omitempty"`
	State       string      `bson:"state,omitempty"             json:"state,omitempty"`
	LastUpdated *time.Time  `bson:"last_updated,omitempty"      json:"last_updated,omitempty"`
	HTMLFiles   []*HTMLFile `bson:"html_files,omitempty"        json:"html_files,omitempty"`
	//Mongo only
	Active *bool  `bson:"active,omitempty"            json:"-"`
	SHA    string `bson:"sha,omitempty"               json:"-"`
	//JSON only
	URL string `bson:"-" json:"url,omitempty"`
	URI string `bson:"-" json:"uri,omitempty"`
}

func (i *Interactive) SetJSONAttribs(domain string) {
	if i != nil && i.Metadata != nil {
		i.URI = fmt.Sprintf("/%s/%s-%s", "interactives", i.Metadata.HumanReadableSlug, i.Metadata.ResourceID)
		i.URL = fmt.Sprintf("%s%s/%s", domain, i.URI, "embed")

		for _, f := range i.HTMLFiles {
			f.URI = fmt.Sprintf("%s/%s", i.URI, f.URI)
		}
	}
}

func (i *Interactive) CanPublish() (ok bool) {
	var state State
	if i != nil {
		if state, ok = ParseState(i.State); !ok {
			return
		}
	}
	return state == ImportSuccess
}

type Archive struct {
	Name                string `bson:"name,omitempty"                   json:"name,omitempty"`
	Size                int64  `bson:"size_in_bytes,omitempty"          json:"size_in_bytes,omitempty"`
	UploadRootDirectory string `bson:"upload_root_directory,omitempty"  json:"upload_root_directory,omitempty"`
	ImportMessage       string `bson:"import_message,omitempty"         json:"import_message,omitempty"`
	//flag from importer - api uses this to determine state
	ImportSuccessful bool `bson:"-"                        json:"import_successful,omitempty"`
}

type HTMLFile struct {
	Name string `bson:"name,omitempty" json:"name,omitempty"`
	URI  string `bson:"uri,omitempty" json:"uri,omitempty"`
}
