package steps

import (
	"context"
	"encoding/json"
	test_support "github.com/ONSdigital/dp-interactives-api/internal/test-support"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/ONSdigital/dp-interactives-api/models"

	"github.com/cucumber/godog"
	"go.mongodb.org/mongo-driver/bson"
)

var WellKnownTestTime time.Time

func init() {
	WellKnownTestTime, _ = time.Parse("2006-01-02T15:04:05Z", "2021-01-01T00:00:00Z")
}

func (c *InteractivesApiComponent) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I have these interactives:$`, c.iHaveTheseInteractives)
	//todo probabaly should move these to dp-component-test library
	ctx.Step(`^I POST file "([^"]*)" with form-data "([^"]*)"$`, c.IPostToWithFormData)
	ctx.Step(`^I PUT file "([^"]*)" with form-data "([^"]*)"$`, c.iPUTFileWithFormdata)
	ctx.Step(`^I PUT no file with form-data "([^"]*)"$`, c.iPUTNoFileWithFormdata)
	ctx.Step(`^I should receive the following JSON response:$`, c.ApiFeature.IShouldReceiveTheFollowingJSONResponse)
}

func (c *InteractivesApiComponent) iHaveTheseInteractives(datasetsJson *godog.DocString) error {
	var componentTestData []struct {
		ID          string                      `json:"id,omitempty"`
		ArchiveName string                      `json:"file_name,omitempty"`
		State       string                      `json:"state,omitempty"`
		Active      bool                        `json:"active,omitempty"`
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
			SHA:      "kqA7qPo1GeOJeff69lByWLbPiZM=",
			State:    testData.State,
			Active:   &testData.Active,
			Metadata: testData.MetaData,
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
	return c.makeRequest(http.MethodPost, path, formFile, []byte(body.Content))
}

func (c *InteractivesApiComponent) iPUTNoFileWithFormdata(path string, body *godog.DocString) error {
	return c.makeRequest(http.MethodPut, path, "-", []byte(body.Content))
}

func (c *InteractivesApiComponent) iPUTFileWithFormdata(formFile, path string, body *godog.DocString) error {
	return c.makeRequest(http.MethodPut, path, formFile, []byte(body.Content))
}

func (c *InteractivesApiComponent) makeRequest(method, path, formFile string, data []byte) error {
	handler, err := c.InitialiseService()
	if err != nil {
		return err
	}

	var update *models.InteractiveUpdate
	err = json.Unmarshal(data, &update)
	if err != nil {
		return err
	}

	req := test_support.NewFileUploadRequest(method, "http://foo"+path, "attachment", formFile, update)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	c.ApiFeature.HttpResponse = w.Result()
	return c.ApiFeature.StepError()
}
