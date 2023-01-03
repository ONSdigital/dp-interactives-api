package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/interactives"
	"github.com/ONSdigital/dp-interactives-api/config"

	"github.com/ONSdigital/dp-interactives-api/models"
	dpMongoDriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	State           string = "State"
)

// GetInteractive retrieves an interactive by its id
func (m *Mongo) GetInteractive(ctx context.Context, id string) (*models.Interactive, error) {
	var interactive *models.Interactive
	err := m.Connection.Collection(m.ActualCollectionName(config.MetadataCollection)).
		FindOne(ctx, bson.M{"_id": id}, &interactive)
	if err != nil {
		if errors.Is(err, dpMongoDriver.ErrNoDocumentFound) {
			return nil, ErrNoRecordFound
		}
		return nil, err
	}

	interactive.SetJSONAttribs(m.PreviewRootURL)

	return interactive, nil
}

func (m *Mongo) ListInteractives(ctx context.Context, modelFilter *models.Filter) ([]*models.Interactive, error) {
	filter := generateFilter(modelFilter)

	var values []*models.Interactive
	_, err := m.Connection.Collection(m.ActualCollectionName(config.MetadataCollection)).
		Find(ctx, filter, &values, dpMongoDriver.Sort(bson.M{"_id": -1}))
	if err != nil {
		return values, err
	}

	for _, interactive := range values {
		interactive.SetJSONAttribs(m.PreviewRootURL)
	}

	return values, nil
}

// UpsertInteractive adds or overides an existing interactive
func (m *Mongo) UpsertInteractive(ctx context.Context, id string, i *models.Interactive) (err error) {
	update := bson.M{
		"$set": i,
		"$currentDate": bson.M{
			"last_updated": true,
		},
		"$setOnInsert": bson.M{
			"created": time.Now(),
		},
	}

	_, err = m.Connection.Collection(m.ActualCollectionName(config.MetadataCollection)).
		UpsertById(ctx, id, update)
	return
}

// PatchInteractive patches an existing interactive
func (m *Mongo) PatchInteractive(ctx context.Context, attribute interactives.PatchAttribute, i *models.Interactive) error {
	collection := m.ActualCollectionName(config.MetadataCollection)

	var patch bson.M
	switch attribute {
	case interactives.PatchArchive:
		patch = bson.M{"archive": i.Archive, "state": i.State}
	case interactives.Publish: // unlink from collection
		patch = bson.M{"published": i.Published, "metadata.collection_id": ""}
	case interactives.LinkToCollection:
		patch = bson.M{"metadata.collection_id": i.Metadata.CollectionID}
	case interactives.PatchAttribute(State):
		patch = bson.M{"state": i.State}
	default:
		return fmt.Errorf("unsupported attribute %s", attribute)
	}

	update := bson.M{
		"$set": patch,
		"$currentDate": bson.M{
			"last_updated": true,
		},
	}

	_, err := m.Connection.Collection(collection).UpdateById(ctx, i.ID, update)
	return err
}
