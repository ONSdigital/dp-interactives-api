package mongo

import (
	"context"
	"github.com/ONSdigital/dp-interactives-api/config"
	"github.com/ONSdigital/dp-interactives-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func (m *Mongo) ListArchiveFiles(ctx context.Context, interactiveId string) ([]*models.ArchiveFile, error) {
	var values []*models.ArchiveFile
	_, err := m.Connection.Collection(m.ActualCollectionName(config.ArchiveCollection)).
		Find(ctx, bson.M{"interactive_id": interactiveId}, &values)
	if err != nil {
		return values, err
	}
	return values, nil
}

// UpsertInteractive adds or overides an existing interactive
func (m *Mongo) UpsertArchiveFile(ctx context.Context, file *models.ArchiveFile) (err error) {
	update := bson.M{
		"$set": file,
		"$currentDate": bson.M{
			"last_updated": true,
		},
		"$setOnInsert": bson.M{
			"created": time.Now(),
		},
	}

	_, err = m.Connection.Collection(m.ActualCollectionName(config.ArchiveCollection)).
		UpsertById(ctx, file.ID, update)
	return
}
