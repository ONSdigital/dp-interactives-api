package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-interactives-api/models"
	dpMongoLock "github.com/ONSdigital/dp-mongodb/v3/dplock"
	dpMongoHealth "github.com/ONSdigital/dp-mongodb/v3/health"
	dpMongoDriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/ONSdigital/log.go/v2/log"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	connectTimeoutInSeconds = 5
	queryTimeoutInSeconds   = 15
	interactivesCol         = "interactives"
)

var (
	ErrNoRecordFound = errors.New("no record exists")
)

type Mongo struct {
	Collection   string
	Database     string
	Connection   *dpMongoDriver.MongoConnection
	URI          string
	Username     string
	Password     string
	healthClient *dpMongoHealth.CheckMongoClient
	lockClient   *dpMongoLock.Lock
	IsSSL        bool
}

func (m *Mongo) getConnectionConfig(shouldEnableReadConcern, shouldEnableWriteConcern bool) *dpMongoDriver.MongoConnectionConfig {
	return &dpMongoDriver.MongoConnectionConfig{
		TLSConnectionConfig: dpMongoDriver.TLSConnectionConfig{
			IsSSL: m.IsSSL,
		},
		ConnectTimeoutInSeconds: connectTimeoutInSeconds,
		QueryTimeoutInSeconds:   queryTimeoutInSeconds,

		Username:                      m.Username,
		Password:                      m.Password,
		ClusterEndpoint:               m.URI,
		Database:                      m.Database,
		Collection:                    m.Collection,
		IsWriteConcernMajorityEnabled: shouldEnableWriteConcern,
		IsStrongReadConcernEnabled:    shouldEnableReadConcern,
	}
}

// Init creates a new mongodb.MongoConnection with a strong consistency and a write mode of "majority".
func (m *Mongo) Init(ctx context.Context, shouldEnableReadConcern, shouldEnableWriteConcern bool) (err error) {
	if m.Connection != nil {
		return errors.New("interactives connection already exists")
	}
	mongoConnection, err := dpMongoDriver.Open(m.getConnectionConfig(shouldEnableReadConcern, shouldEnableWriteConcern))
	if err != nil {
		return err
	}
	m.Connection = mongoConnection

	databaseCollectionBuilder := make(map[dpMongoHealth.Database][]dpMongoHealth.Collection)
	databaseCollectionBuilder[(dpMongoHealth.Database)(m.Database)] = []dpMongoHealth.Collection{(dpMongoHealth.Collection)(m.Collection)}
	// Create health-client from session
	m.healthClient = dpMongoHealth.NewClientWithCollections(m.Connection, databaseCollectionBuilder)

	// Create MongoDB lock client, which also starts the purger loop
	m.lockClient = dpMongoLock.New(ctx, m.Connection, interactivesCol)
	return nil
}

// GetActiveInteractiveFromSHA retrieves an active interactive by its SHA
func (m *Mongo) GetActiveInteractiveFromSHA(ctx context.Context, sha string) (*models.Interactive, error) {
	log.Info(ctx, "getting interactive by SHA", log.Data{"sha": sha})

	var vis models.Interactive
	err := m.Connection.GetConfiguredCollection().FindOne(ctx, bson.M{"sha": sha, "active": true}, &vis)
	if err != nil {
		if dpMongoDriver.IsErrNoDocumentFound(err) {
			return nil, ErrNoRecordFound
		}
		return nil, err
	}

	return &vis, nil
}

// GetInteractive retrieves an interactive by its id
func (m *Mongo) GetInteractive(ctx context.Context, id string) (*models.Interactive, error) {
	log.Info(ctx, "getting interactive by id", log.Data{"_id": id})

	var vis models.Interactive
	err := m.Connection.GetConfiguredCollection().FindOne(ctx, bson.M{"_id": id}, &vis)
	if err != nil {
		if dpMongoDriver.IsErrNoDocumentFound(err) {
			return nil, ErrNoRecordFound
		}
		return nil, err
	}

	return &vis, nil
}

func (m *Mongo) ListInteractives(ctx context.Context, offset, limit int) ([]*models.Interactive, int, error) {

	selector := bson.M{}
	selector["active"] = bson.M{"$eq": true}
	f := m.Connection.GetConfiguredCollection().Find(selector).Sort(bson.M{"_id": -1})

	// get total count and paginated values according to provided offset and limit
	var values []*models.Interactive
	totalCount, err := QueryPage(ctx, f, offset, limit, &values)
	if err != nil {
		return values, 0, err
	}

	return values, totalCount, nil
}

func QueryPage(ctx context.Context, f *dpMongoDriver.Find, offset, limit int, result *[]*models.Interactive) (totalCount int, err error) {

	// get total count of items for the provided query
	totalCount, err = f.Count(ctx)
	if err != nil {
		log.Error(ctx, "error counting items", err)
		return 0, err
	}

	// query the items corresponding to the provided offset and limit (only if necessary)
	// guaranteeing at least one document will be found
	if totalCount > 0 && limit > 0 && offset < totalCount {
		return totalCount, f.Skip(offset).Limit(limit).IterAll(ctx, result)
	}

	return totalCount, nil
}

// UpsertInteractive adds or overides an existing interactive
func (m *Mongo) UpsertInteractive(ctx context.Context, id string, vis *models.Interactive) (err error) {
	log.Info(ctx, "upserting interactive", log.Data{"id": id})

	update := bson.M{
		"$set": vis,
		"$setOnInsert": bson.M{
			"last_updated": time.Now(),
		},
	}

	_, err = m.Connection.GetConfiguredCollection().UpsertById(ctx, id, update)
	return
}

// Close closes the mongo session and returns any error
func (m *Mongo) Close(ctx context.Context) error {
	m.lockClient.Close(ctx)
	return m.Connection.Close(ctx)
}

// Checker is called by the healthcheck library to check the health state of this mongoDB instance
func (m *Mongo) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	return m.healthClient.Checker(ctx, state)
}
