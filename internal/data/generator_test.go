package data_test

import (
	"github.com/ONSdigital/dp-interactives-api/internal/data"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	resourceIdGenerator = data.GenerateResourceId()
	slugGenerator       = data.GenerateHumanReadableSlug()
)

func TestGetResourceId(t *testing.T) {
	Convey("When we GenerateResourceId", t, func() {
		got := resourceIdGenerator("")
		So(got, ShouldHaveLength, 8)
	})
}

func TestGenerateHumanReadableSlug(t *testing.T) {
	type test struct {
		title, expected string
	}
	tests := []test{
		{"A Simple Title With An Article or Two", "simple-title-with-article-or-two"},
		{"A An And Simple Title With An The Article or Two", "simple-title-with-article-or-two"},
		{"    An Simple Title    With    An Article or\n\n\t\nTwo And    Some Of The    Whitespace   \n\r\t\n", "simple-title-with-article-or-two-some-of-whitespace"},
		{"A An The", ""},
		{"", ""},
		//the slug can be updated before publishing - we will rerun it before saving
		{"simple-title-with-article-or-two", "simple-title-with-article-or-two"},
		{"simple-title-with- an-article-or-two", "simple-title-with-article-or-two"},
		{"simple-title-with- an-the-article-or-two", "simple-title-with-article-or-two"},
	}
	for _, tc := range tests {
		Convey("When we GenerateHumanReadableSlug with title["+tc.title+"]", t, func() {
			got := slugGenerator(tc.title)
			So(got, ShouldEqual, tc.expected)
		})
	}
}
