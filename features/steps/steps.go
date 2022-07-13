package steps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/dp-authorisation/v2/authorisationtest"
	"github.com/ONSdigital/dp-interactives-api/config"
	test_support "github.com/ONSdigital/dp-interactives-api/internal/test-support"
	"github.com/ONSdigital/dp-interactives-api/models"
	"github.com/cucumber/godog"
	"github.com/rdumont/assistdog"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

const (
	expectedContentType = "application/json; charset=utf-8"
)

var (
	WellKnownTestTime, _ = time.Parse("2006-01-02T15:04:05Z", "2021-01-01T00:00:00Z")
)

type assistArchiveFiles struct {
	Name, InteractiveID, Mimetype, URI string
	Size                               int
}

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
	ctx.Step(`^I should have these archive files:$`, c.iShouldHaveTheseArchiveFiles)
}

func (c *InteractivesApiComponent) adminJWTToken() error {
	err := c.ApiFeature.ISetTheHeaderTo("Authorization", authorisationtest.AdminJWTToken)
	return err
}

func (c *InteractivesApiComponent) iHaveTheseInteractives(datasetsJson *godog.DocString) error {
	var componentTestData []struct {
		ID          string           `json:"id,omitempty"`
		ArchiveName string           `json:"file_name,omitempty"`
		State       string           `json:"state,omitempty"`
		Active      bool             `json:"active,omitempty"`
		Published   bool             `json:"published,omitempty"`
		MetaData    *models.Metadata `json:"metadata,omitempty"`
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

	_, err := c.MongoClient.Connection.Collection(collectionName).UpsertById(context.Background(), id, update)

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

func (c *InteractivesApiComponent) iShouldHaveTheseArchiveFiles(table *godog.Table) error {
	assist := assistdog.NewDefault()
	slice, err := assist.CreateSlice(&assistArchiveFiles{}, table)
	if err != nil {
		return err
	}

	files := slice.([]*assistArchiveFiles)
	interactiveId := files[0].InteractiveID
	collection := c.MongoClient.ActualCollectionName(config.ArchiveCollection)

	var db []*models.ArchiveFile
	_, err = c.MongoClient.Connection.Collection(collection).Find(context.TODO(), bson.M{"interactive_id": interactiveId}, &db)
	if err != nil {
		return err
	}

	assert.Nil(c, err)
	assert.Equal(c, len(files), len(db))
	for i := 0; i < len(files); i++ {
		assert.Equal(c, files[i].Name, db[i].Name)
		assert.Equal(c, files[i].InteractiveID, db[i].InteractiveID)
		assert.Equal(c, files[i].Mimetype, db[i].Mimetype)
		assert.Equal(c, files[i].URI, db[i].URI)
		assert.EqualValues(c, files[i].Size, db[i].Size)
	}

	return c.StepError()
}

func (c *InteractivesApiComponent) makeRequest(method, path, formFile string, data []byte, admin bool) error {
	handler, err := c.InitialiseService()
	if err != nil {
		return err
	}

	var req *http.Request
	var i *models.Interactive
	if data != nil {
		sanitisedData := strings.ReplaceAll(string(data), "\n", "")
		sanitisedData = strings.ReplaceAll(sanitisedData, " ", "")
		if sanitisedData != "{}" {
			err = json.Unmarshal(data, &i)
			if err != nil {
				return err
			}
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
