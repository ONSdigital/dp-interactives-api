package models

import (
	"github.com/ONSdigital/dp-api-clients-go/v2/interactives"
)

func ToRest(i *Interactive) (*interactives.Interactive, error) {

	response := &interactives.Interactive{ID: i.ID, Metadata: i.Metadata}
	if i.Archive.Name != "" {
		response.Archive = &interactives.InteractiveArchive{Name: i.Archive.Name}
		if len(i.Archive.Files) > 0 {
			response.Archive.Size = i.Archive.Size
			for _, f := range i.Archive.Files {
				response.Archive.Files = append(response.Archive.Files, &interactives.InteractiveFile{
					Name:     f.Name,
					Mimetype: f.Mimetype,
					Size:     f.Size,
				})
			}
		}
	}

	return response, nil
}
