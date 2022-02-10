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

	interactives := []models.Interactive{}

	err := json.Unmarshal([]byte(datasetsJson.Content), &interactives)
	if err != nil {
		return err
	}

	for timeOffset, datasetDoc := range interactives {
		datasetID := datasetDoc.ID

		if err := c.putDocumentInDatabase(datasetDoc, datasetID, "interactives", timeOffset); err != nil {
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