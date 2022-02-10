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

// GetInteractiveFromSHA retrieves a interactive by its SHA
func (m *Mongo) GetInteractiveFromSHA(ctx context.Context, sha string) (*models.Interactive, error) {
	log.Info(ctx, "getting interactive by SHA", log.Data{"sha": sha})

	var vis models.Interactive
	err := m.Connection.GetConfiguredCollection().FindOne(ctx, bson.M{"sha": sha}, &vis)
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
