package models_test

import (
	"testing"

	"github.com/ONSdigital/dp-interactives-api/models"

	. "github.com/smartystreets/goconvey/convey"
)

const domain = "domain"

func TestSetUrl(t *testing.T) {
	Convey("When we SetUrl on nil interactive", t, func() {
		var i *models.Interactive
		i.SetURL(domain)
	})

	Convey("When we SetUrl on an empty interactive", t, func() {
		i := &models.Interactive{}
		i.SetURL(domain)
		So(i.URL, ShouldEqual, "")
	})

	Convey("When we SetUrl on an valid interactive", t, func() {
		i := &models.Interactive{
			Metadata: &models.InteractiveMetadata{
				HumanReadableSlug: "slug",
				ResourceID:        "resource_id",
			},
		}
		i.SetURL(domain)
		So(i.URL, ShouldEqual, "domain/interactives/slug-resource_id/embed")
	})
}
