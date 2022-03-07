package steps

import (
	"context"
	"encoding/json"
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
