package models_test

import (
	"testing"

	"github.com/ONSdigital/dp-interactives-api/models"

	. "github.com/smartystreets/goconvey/convey"
)

const domain = "domain"

func TestSetJSONAttribs(t *testing.T) {
	Convey("When we SetJSONAttribs on nil interactive", t, func() {
		var i *models.Interactive
		i.SetJSONAttribs(domain)
	})

	Convey("When we SetJSONAttribs on an empty interactive", t, func() {
		i := &models.Interactive{}
		i.SetJSONAttribs(domain)
		So(i.URL, ShouldEqual, "")
	})

	Convey("When we SetJSONAttribs on an valid interactive", t, func() {
		i := &models.Interactive{
			Metadata: &models.Metadata{
				HumanReadableSlug: "slug",
				ResourceID:        "resource_id",
			},
			HTMLFiles: []*models.HTMLFile{
				{Name: "one", URI: "one"},
				{Name: "two", URI: "one/two"},
			},
		}
		i.SetJSONAttribs(domain)
		So(i.URL, ShouldEqual, "domain/interactives/slug-resource_id/embed")
		So(i.URI, ShouldEqual, "/interactives/slug-resource_id")
		So(len(i.HTMLFiles), ShouldEqual, 2)
		So(i.HTMLFiles[0].URI, ShouldEqual, "/interactives/slug-resource_id/one")
		So(i.HTMLFiles[1].URI, ShouldEqual, "/interactives/slug-resource_id/one/two")
	})
}
