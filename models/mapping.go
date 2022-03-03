package models

import (
	"encoding/json"
)

func Map(in *Interactive) (*Interactive, error) {
	metadata := map[string]string{}
	err := json.Unmarshal([]byte(in.MetadataJson), &metadata)
	if err != nil {
		return nil, err
	}

	response := &Interactive{ID: in.ID, Metadata: metadata}
	if in.Archive.Name != "" {
		response.Archive = &Archive{Name: in.Archive.Name}
		if len(in.Archive.Files) > 0 {
			response.Archive.Size = in.Archive.Size
			for _, f := range in.Archive.Files {
				response.Archive.Files = append(response.Archive.Files, &File{
					Name:     f.Name,
					Mimetype: f.Mimetype,
					Size:     f.Size,
				})
			}
		}
	}

	return response, nil
}
