package api

import (
	"archive/zip"
	"bytes"
	"path/filepath"
	"strings"

	"github.com/ONSdigital/dp-interactives-api/models"

	"github.com/pkg/errors"
)

var (
	ErrNoIndexHtml = errors.New("interactive must contain 1 htm(l) file")
)

//this is a tactical solution - we need ot know about html files on new upload
//so zebedee collection json populated - this is for preview - as we process the
//zip async - zebedee doesnt get this info quick enough

func Open(name string, zipFile []byte) (*models.Archive, error) {
	size := int64(len(zipFile))
	zipReader, err := zip.NewReader(bytes.NewReader(zipFile), size)
	if err != nil {
		return nil, err
	}

	var hasHtmFile bool
	var htmlFiles []*models.File
	for _, f := range zipReader.File {
		filename := filepath.Base(f.Name)
		fileExt := filepath.Ext(f.Name)
		if strings.EqualFold(fileExt, ".html") || strings.EqualFold(fileExt, ".htm") {
			hasHtmFile = true
			//we only care about html files right now for preview
			//the patch from importer will overwrite with full details
			htmlFiles = append(htmlFiles, &models.File{
				Name:     filename,
				Mimetype: "tbc",
				Size:     int64(f.UncompressedSize64),
				URI:      f.Name,
			})
		}
	}

	if !hasHtmFile {
		return nil, ErrNoIndexHtml
	}

	return &models.Archive{
		Name:  name,
		Size:  size,
		Files: htmlFiles,
	}, nil
}
