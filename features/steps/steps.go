package steps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"

	test_support "github.com/ONSdigital/dp-interactives-api/internal/test-support"
	"github.com/stretchr/testify/assert"

	"github.com/ONSdigital/dp-authorisation/v2/authorisationtest"
	"github.com/ONSdigital/dp-interactives-api/models"

	"github.com/cucumber/godog"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	expectedContentType = "application/json; charset=utf-8"
)

var (
	WellKnownTestTime, _ = time.Parse("2006-01-02T15:04:05Z", "2021-01-01T00:00:00Z")
)

func (c *InteractivesApiComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I have these interactives:$`, c.iHaveTheseInteractives)
	//todo probabaly should move these to dp-component-test library
	ctx.Step(`^I POST file "([^"]*)" with form-data "([^"]*)"$`, c.IPostToWithFormData)
	ctx.Step(`^I PUT file "([^"]*)" with form-data "([^"]*)"$`, c.iPUTFileWithFormdata)
	ctx.Step(`^I PUT no file with form-data "([^"]*)"$`, c.iPUTNoFileWithFormdata)
	ctx.Step(`^I should receive the following model response with status "([^"]*)":$`, c.IShouldReceiveTheFollowingModelResponse)
	ctx.Step(`^I should receive the following list\(model\) response with status "([^"]*)":$`, c.iShouldReceiveTheFollowingListmodelResponseWithStatus)
	ctx.Step(`^I am an interactives user`, c.adminJWTToken)
	ctx.Step(`^As an interactives user I POST file "([^"]*)" with form-data "([^"]*)"$`, c.IPostToWithFormDataAsAdmin)
	ctx.Step(`^As an interactives user I PUT file "([^"]*)" with form-data "([^"]*)"$`, c.iPUTFileWithFormdataAsAdmin)
	ctx.Step(`^As an interactives user I PUT no file with form-data "([^"]*)"$`, c.iPUTNoFileWithFormdataAsAdmin)
	ctx.Step(`^As an interactives user with filter I GET '(.*)'$`, c.IGetWithFilterString)
}

func (c *InteractivesApiComponent) adminJWTToken() error {
	err := c.ApiFeature.ISetTheHeaderTo("Authorization", authorisationtest.AdminJWTToken)
	return err
}

func (c *InteractivesApiComponent) iHaveTheseInteractives(datasetsJson *godog.DocString) error {
	var componentTestData []struct {
		ID          string                      `json:"id,omitempty"`
		ArchiveName string                      `json:"file_name,omitempty"`
		State       string                      `json:"state,omitempty"`
		Active      bool                        `json:"active,omitempty"`
		Published   bool                        `json:"published,omitempty"`
		MetaData    *models.InteractiveMetadata `json:"metadata,omitempty"`
	}

	err := json.Unmarshal([]byte(datasetsJson.Content), &componentTestData)
	if err != nil {
		return err
	}

	for timeOffset, testData := range componentTestData {
		mongoInteractive := &models.Interactive{
			ID: "0d77a889-abb2-4432-ad22-9c23cf7ee796",
			Archive: &models.Archive{
				Name: "kqA7qPo1GeOJeff69lByWLbPiZM=/docker-vernemq-master.zip",
			},
			SHA:       "kqA7qPo1GeOJeff69lByWLbPiZM=",
			State:     testData.State,
			Active:    &testData.Active,
			Published: &testData.Published,
			Metadata:  testData.MetaData,
		}
		if testData.ID != "" {
			mongoInteractive.ID = testData.ID
		}
		if testData.ArchiveName != "" {
			mongoInteractive.Archive.Name = testData.ArchiveName
		}

		if err := c.putDocumentInDatabase(mongoInteractive, mongoInteractive.ID, "metadata", timeOffset); err != nil {
			return err
		}
	}

	return nil
}

func (c *InteractivesApiComponent) putDocumentInDatabase(document interface{}, id, collectionName string, timeOffset int) error {
	update := bson.M{
		"$set": document,
		"$setOnInsert": bson.M{
			"last_updated": WellKnownTestTime.Add(time.Second * time.Duration(timeOffset)),
		},
	}

	_, err := c.MongoClient.Connection.C(collectionName).UpsertById(context.Background(), id, update)

	if err != nil {
		return err
	}
	return nil
}

func (c *InteractivesApiComponent) IPostToWithFormData(formFile, path string, body *godog.DocString) error {
	return c.makeRequest(http.MethodPost, path, formFile, []byte(body.Content), false)
}

func (c *InteractivesApiComponent) iPUTNoFileWithFormdata(path string, body *godog.DocString) error {
	return c.makeRequest(http.MethodPut, path, "-", []byte(body.Content), false)
}

func (c *InteractivesApiComponent) iPUTFileWithFormdata(formFile, path string, body *godog.DocString) error {
	return c.makeRequest(http.MethodPut, path, formFile, []byte(body.Content), false)
}

func (c *InteractivesApiComponent) IPostToWithFormDataAsAdmin(formFile, path string, body *godog.DocString) error {
	return c.makeRequest(http.MethodPost, path, formFile, []byte(body.Content), true)
}

func (c *InteractivesApiComponent) IGetWithFilterString(path string) error {
	return c.makeRequest(http.MethodGet, path, "-", nil, true)
}

func (c *InteractivesApiComponent) iPUTNoFileWithFormdataAsAdmin(path string, body *godog.DocString) error {
	return c.makeRequest(http.MethodPut, path, "-", []byte(body.Content), true)
}

func (c *InteractivesApiComponent) iPUTFileWithFormdataAsAdmin(formFile, path string, body *godog.DocString) error {
	return c.makeRequest(http.MethodPut, path, formFile, []byte(body.Content), true)
}

func (c *InteractivesApiComponent) makeRequest(method, path, formFile string, data []byte, admin bool) error {
	handler, err := c.InitialiseService()
	if err != nil {
		return err
	}

	var req *http.Request
	var i *models.Interactive
	if data != nil {
		err = json.Unmarshal(data, &i)
		if err != nil {
			return err
		}

		req = test_support.NewFileUploadRequest(method, "http://foo"+path, "attachment", formFile, i)
	} else {
		req = httptest.NewRequest(method, "http://foo"+path, bytes.NewReader(data))
	}
	if admin {
		req.Header.Set("Authorization", authorisationtest.AdminJWTToken)
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	c.ApiFeature.HttpResponse = w.Result()
	return c.StepError()
}

func (c *InteractivesApiComponent) iShouldReceiveTheFollowingListmodelResponseWithStatus(expectedCodeStr string, expectedAPIResponse *godog.DocString) error {
	var expected, actual []models.Interactive
	err := c.toModel(expectedCodeStr, expectedAPIResponse, &expected, &actual)
	if err != nil {
		return err
	}
	assert.Equal(c, expected, actual)
	return c.StepError()
}

func (c *InteractivesApiComponent) IShouldReceiveTheFollowingModelResponse(expectedCodeStr string, expectedAPIResponse *godog.DocString) error {
	var expected, actual models.Interactive
	err := c.toModel(expectedCodeStr, expectedAPIResponse, &expected, &actual)
	if err != nil {
		return err
	}

	if expected.LastUpdated == nil {
		//expect a date today
		now := time.Now()
		a, b, c := now.Date()
		x, y, z := actual.LastUpdated.Date()
		if a != x && b != y && c != z {
			return fmt.Errorf("last_updated  not as expected %s", actual.LastUpdated)
		}
		expected.LastUpdated, actual.LastUpdated = &now, &now
	}

	assert.Equal(c, expected, actual)

	return c.StepError()
}

func (c *InteractivesApiComponent) toModel(expectedCodeStr string, expectedAPIResponse *godog.DocString, expected, actual interface{}) error {
	if err := c.ApiFeature.TheHTTPStatusCodeShouldBe(expectedCodeStr); err != nil {
		return err
	}

	if err := c.ApiFeature.TheResponseHeaderShouldBe("Content-Type", expectedContentType); err != nil {
		return err
	}

	err := json.Unmarshal([]byte(expectedAPIResponse.Content), expected)
	if err != nil {
		return err
	}

	responseBody := c.ApiFeature.HttpResponse.Body
	body, _ := ioutil.ReadAll(responseBody)
	err = json.Unmarshal(body, actual)
	if err != nil {
		return err
	}

	return nil
}
