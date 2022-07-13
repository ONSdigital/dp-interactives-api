package zip

import (
	"archive/zip"
	"errors"
	"github.com/ONSdigital/dp-interactives-api/models"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrNoIndexHtml = errors.New("interactive must contain 1 htm(l) file")
)

//this is a tactical solution - we need to know about html files on new upload
//so zebedee collection json populated as expected for preview
//because we process the zip async - zebedee doesnt get this info quick enough

func Open(name string) (*models.Archive, []*models.HTMLFile, error) {
	fi, err := os.Stat(name)
	if err != nil {
		return nil, nil, err
	}

	zipReader, err := zip.OpenReader(name)
	if err != nil {
		return nil, nil, err
	}

	var hasHtmFile bool
	var htmlFiles []*models.HTMLFile
	for _, f := range zipReader.File {
		filename := filepath.Base(f.Name)
		if filename[0] == '.' {
			//skip hidden files
			continue
		}

		fileExt := filepath.Ext(f.Name)
		if strings.EqualFold(fileExt, ".html") || strings.EqualFold(fileExt, ".htm") {
			hasHtmFile = true
			//we only care about html files right now for preview
			//the patch from importer will overwrite with full details
			htmlFiles = append(htmlFiles, &models.HTMLFile{
				Name: filename,
				URI:  f.Name,
			})
		}
	}

	if !hasHtmFile {
		return nil, nil, ErrNoIndexHtml
	}

	return &models.Archive{Size: fi.Size()}, htmlFiles, nil
}
