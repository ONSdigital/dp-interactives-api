package api_test

import (
	"archive/zip"
	"github.com/ONSdigital/dp-interactives-api/api"
	"io"
	"os"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestArchive(t *testing.T) {

	Convey("Given an invalid zip file", t, func() {
		archive, err := os.CreateTemp("", "test-zip_*.zip")
		So(err, ShouldBeNil)
		defer os.Remove(archive.Name())
		Convey("Then there should an error returned when attempt to open", func() {
			a, err := api.Open(archive.Name())
			So(err, ShouldBeError, zip.ErrFormat)
			So(a, ShouldBeNil)
		})
	})

	Convey("Given a valid zip file", t, func() {
		archiveName, _, err := createTestZip("root.css", "root.html", "root.js", "index.html")
		defer os.Remove(archiveName)
		So(err, ShouldBeNil)
		So(archiveName, ShouldNotBeEmpty)

		Convey("Then open should run successfully", func() {
			a, err := api.Open(archiveName)
			So(err, ShouldBeNil)

			Convey("And files in archive should be 4", func() {
				So(len(a.Files), ShouldEqual, 2)
			})
		})
	})

	Convey("Given an invalid zip file (no .html file present)", t, func() {
		archiveName, _, err := createTestZip("root.css", "root.js")
		defer os.Remove(archiveName)
		So(err, ShouldBeNil)
		So(archiveName, ShouldNotBeEmpty)

		Convey("Then open should run successfully", func() {
			a, err := api.Open(archiveName)
			So(err, ShouldEqual, api.ErrNoIndexHtml)
			So(a, ShouldBeNil)
		})
	})
}

func createTestZip(filenames ...string) (string, []byte, error) {
	archive, err := os.CreateTemp("", "test-zip_*.zip")
	if err != nil {
		return "", nil, err
	}

	zipWriter := zip.NewWriter(archive)
	for _, f := range filenames {
		w, err := zipWriter.Create(f)
		if err != nil {
			return "", nil, err
		}
		if _, err = io.Copy(w, strings.NewReader(f)); err != nil {
			return "", nil, err
		}
	}

	if err = zipWriter.Flush(); err != nil {
		return "", nil, err
	}
	if err = zipWriter.Close(); err != nil {
		return "", nil, err
	}
	if err = archive.Close(); err != nil {
		return "", nil, err
	}

	archive, err = os.Open(archive.Name())
	if err != nil {
		return "", nil, err
	}

	b, err := io.ReadAll(archive)
	if err != nil {
		return "", nil, err
	}

	if err = archive.Close(); err != nil {
		return "", nil, err
	}

	return archive.Name(), b, nil
}
