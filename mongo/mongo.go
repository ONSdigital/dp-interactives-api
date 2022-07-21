package mongo

import (
	"context"
	"errors"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-interactives-api/config"
	"github.com/ONSdigital/dp-interactives-api/models"
	mongohealth "github.com/ONSdigital/dp-mongodb/v3/health"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
	"strings"
)

var (
	ErrNoRecordFound = errors.New("no record exists")
)

type Mongo struct {
	config.MongoConfig
	PreviewRootURL string
	Connection     *mongodriver.MongoConnection
	healthClient   *mongohealth.CheckMongoClient
}

// Init returns an initialised Mongo object encapsulating a connection to the mongo server/cluster with the given configuration,
// a health client to check the health of the mongo server/cluster, and a lock client
func (m *Mongo) Init(ctx context.Context) (err error) {
	m.Connection, err = mongodriver.Open(&m.MongoDriverConfig)
	if err != nil {
		return err
	}

	databaseCollectionBuilder := map[mongohealth.Database][]mongohealth.Collection{
		(mongohealth.Database)(m.Database): {
			mongohealth.Collection(m.ActualCollectionName(config.MetadataCollection)),
		},
	}
	m.healthClient = mongohealth.NewClientWithCollections(m.Connection, databaseCollectionBuilder)

	return nil
}

// Close represents mongo session closing within the context deadline
func (m *Mongo) Close(ctx context.Context) error {
	return m.Connection.Close(ctx)
}

// Checker is called by the healthcheck library to check the health state of this mongoDB instance
func (m *Mongo) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	return m.healthClient.Checker(ctx, state)
}

// Reflect into the metadata structure
// generate filter string depending on data type
// string eq value
// array in value(s)
func generateFilter(model *models.Filter) bson.M {
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
