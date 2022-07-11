package mongo

import (
	"context"
	"errors"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-interactives-api/config"
	mongohealth "github.com/ONSdigital/dp-mongodb/v3/health"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
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
			mongohealth.Collection(m.ActualCollectionName(config.ArchiveCollection)),
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
