package mongo

import (
	"context"
	"errors"
	"fmt"
	"github.com/ONSdigital/dp-interactives-api/config"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/ONSdigital/dp-interactives-api/models"
	dpMongoDriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/ONSdigital/log.go/v2/log"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	Archive PatchAttribute = iota
	Publish
	LinkToCollection
)

type PatchAttribute int

func (a PatchAttribute) String() string {
	switch a {
	case Archive:
		return "Archive"
	case LinkToCollection:
		return "LinkToCollection"
	case Publish:
		return "Publish"
	default:
		return ""
	}
}

// GetActiveInteractiveFromSHA retrieves an active interactive by its SHA
func (m *Mongo) GetActiveInteractiveGivenSha(ctx context.Context, sha string) (*models.Interactive, error) {
	log.Info(ctx, "getting interactive by SHA", log.Data{"sha": sha})

	var vis models.Interactive
	err := m.Connection.Collection(m.ActualCollectionName(config.MetadataCollection)).
		FindOne(ctx, bson.M{"sha": sha, "active": true}, &vis)
	if err != nil {
		if errors.Is(err, dpMongoDriver.ErrNoDocumentFound) {
			return nil, ErrNoRecordFound
		}
		return nil, err
	}

	return &vis, nil
}

// GetActiveInteractiveGivenField retrieves an active interactive by its field
func (m *Mongo) GetActiveInteractiveGivenField(ctx context.Context, fieldName, fieldValue string) (*models.Interactive, error) {
	log.Info(ctx, "getting interactive by field", log.Data{fieldName: fieldValue})

	var vis models.Interactive
	err := m.Connection.Collection(m.ActualCollectionName(config.MetadataCollection)).
		FindOne(ctx, bson.M{fieldName: fieldValue, "active": true}, &vis)
	if err != nil {
		if errors.Is(err, dpMongoDriver.ErrNoDocumentFound) {
			return nil, ErrNoRecordFound
		}
		return nil, err
	}

	return &vis, nil
}

// GetInteractive retrieves an interactive by its id
func (m *Mongo) GetInteractive(ctx context.Context, id string) (*models.Interactive, error) {
	log.Info(ctx, "getting interactive by id", log.Data{"_id": id})

	var interactive models.Interactive
	err := m.Connection.Collection(m.ActualCollectionName(config.MetadataCollection)).
		FindOne(ctx, bson.M{"_id": id}, &interactive)
	if err != nil {
		if errors.Is(err, dpMongoDriver.ErrNoDocumentFound) {
			return nil, ErrNoRecordFound
		}
		return nil, err
	}

	//derived attributes
	interactive.SetURL(m.PreviewRootURL)
	if interactive.Archive != nil {
		var htmlFiles []models.HTMLFile
		for _, f := range interactive.Archive.Files {
			if f == nil {
				continue
			}
			filename := filepath.Base(f.URI)
			fileExt := filepath.Ext(f.URI)
			if strings.EqualFold(fileExt, ".html") || strings.EqualFold(fileExt, ".htm") {
				htmlFiles = append(htmlFiles, models.HTMLFile{
					Name: filename,
					URI:  fmt.Sprintf("%s/%s", interactive.URI, f.URI),
				})
			}
		}
		interactive.HTMLFiles = &htmlFiles
	}

	return &interactive, nil
}

func (m *Mongo) ListInteractives(ctx context.Context, modelFilter *models.InteractiveFilter) ([]*models.Interactive, error) {
	filter := generateFilter(modelFilter)

	var values []*models.Interactive
	_, err := m.Connection.Collection(m.ActualCollectionName(config.MetadataCollection)).
		Find(ctx, filter, &values, dpMongoDriver.Sort(bson.M{"_id": -1}))
	if err != nil {
		return values, err
	}

	for _, interactive := range values {
		interactive.SetURL(m.PreviewRootURL)
	}

	return values, nil
}

// Reflect into the metadata structure
// generate filter string depending on data type
// string eq value
// array in value(s)
func generateFilter(model *models.InteractiveFilter) bson.M {
	filter := bson.M{}
	filter["active"] = bson.M{"$eq": true}
	if model == nil || model.Metadata == nil {
		return filter
	}

	// if filter by collection-id
	if model.AssociateCollection {
		// collection_id == given
		// OR
		// no collection_id AND not published
		fil := bson.M{}
		fil["$and"] = []interface{}{bson.M{"published": true}}
		filter["$or"] = []interface{}{bson.M{"metadata.collection_id": model.Metadata.CollectionID}, fil}
		return filter
	}

	// if there is a resource_id set, filter using that
	if model.Metadata.ResourceID != "" {
		filter["metadata.resource_id"] = bson.M{"$eq": model.Metadata.ResourceID}
		return filter
	}

	// else filter using other metadata
	v := reflect.ValueOf(*(model.Metadata))
	typeOfS := v.Type()

	for i := 0; i < v.NumField(); i++ {
		tag := strings.Split(typeOfS.Field(i).Tag.Get("json"), ",")[0]
		valType := typeOfS.Field(i).Type.Kind()
		val := v.Field(i).Interface()

		switch valType {
		case reflect.String:
			if val != "" {
				filter["metadata."+tag] = bson.M{"$regex": val, "$options": "i"}
			}
		}
	}
	return filter
}

// UpsertInteractive adds or overides an existing interactive
func (m *Mongo) UpsertInteractive(ctx context.Context, id string, vis *models.Interactive) (err error) {
	log.Info(ctx, "upserting interactive", log.Data{"id": id})

	update := bson.M{
		"$set": vis,
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
func (m *Mongo) PatchInteractive(ctx context.Context, attribute PatchAttribute, i *models.Interactive) error {
	log.Info(ctx, "patching interactive", log.Data{"id": i.ID})

	var patch bson.M
	switch attribute {
	case Archive:
		patch = bson.M{"archive": i.Archive, "state": i.State}
	case Publish: // unlink from collection
		patch = bson.M{"published": i.Published, "metadata.collection_id": ""}
	case LinkToCollection:
		patch = bson.M{"metadata.collection_id": i.Metadata.CollectionID}
	default:
		return fmt.Errorf("unsupported attribute %s", attribute)
	}

	update := bson.M{
		"$set": patch,
		"$currentDate": bson.M{
			"last_updated": true,
		},
	}
	_, err := m.Connection.Collection(m.ActualCollectionName(config.MetadataCollection)).
		UpdateById(ctx, i.ID, update)
	return err
}
