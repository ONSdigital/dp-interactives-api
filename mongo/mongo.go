package mongo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-interactives-api/config"
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

	ImportArchive PatchAction = iota
)

var (
	ErrNoRecordFound = errors.New("no record exists")
)

type PatchAction int64

func (a PatchAction) String() string {
	switch a {
	case ImportArchive:
		return "ImportArchive"
	}
	return "unknown"
}

type Mongo struct {
	Config       *config.Config
	Connection   *dpMongoDriver.MongoConnection
	healthClient *dpMongoHealth.CheckMongoClient
	lockClient   *dpMongoLock.Lock
}

func (m *Mongo) Database() string {
	return m.Config.MongoConfig.Database
}

func (m *Mongo) Collection() string {
	return m.Config.MongoConfig.Collection
}

func (m *Mongo) getConnectionConfig(shouldEnableReadConcern, shouldEnableWriteConcern bool) *dpMongoDriver.MongoConnectionConfig {
	return &dpMongoDriver.MongoConnectionConfig{
		TLSConnectionConfig: dpMongoDriver.TLSConnectionConfig{
			IsSSL: m.Config.MongoConfig.IsSSL,
		},
		ConnectTimeoutInSeconds: connectTimeoutInSeconds,
		QueryTimeoutInSeconds:   queryTimeoutInSeconds,

		Username:                      m.Config.MongoConfig.Username,
		Password:                      m.Config.MongoConfig.Password,
		ClusterEndpoint:               m.Config.MongoConfig.BindAddr,
		Database:                      m.Config.MongoConfig.Database,
		Collection:                    m.Config.MongoConfig.Collection,
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
	databaseCollectionBuilder[(dpMongoHealth.Database)(m.Database())] = []dpMongoHealth.Collection{(dpMongoHealth.Collection)(m.Collection())}
	// Create health-client from session
	m.healthClient = dpMongoHealth.NewClientWithCollections(m.Connection, databaseCollectionBuilder)

	// Create MongoDB lock client, which also starts the purger loop
	m.lockClient = dpMongoLock.New(ctx, m.Connection, interactivesCol)
	return nil
}

// GetActiveInteractiveFromSHA retrieves an active interactive by its SHA
func (m *Mongo) GetActiveInteractiveGivenSha(ctx context.Context, sha string) (*models.Interactive, error) {
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

// GetActiveInteractiveGivenField retrieves an active interactive by its field
func (m *Mongo) GetActiveInteractiveGivenField(ctx context.Context, fieldName, fieldValue string) (*models.Interactive, error) {
	log.Info(ctx, "getting interactive by field", log.Data{fieldName: fieldValue})

	var vis models.Interactive
	err := m.Connection.GetConfiguredCollection().FindOne(ctx, bson.M{fieldName: fieldValue, "active": true}, &vis)
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

	var interactive models.Interactive
	err := m.Connection.GetConfiguredCollection().FindOne(ctx, bson.M{"_id": id}, &interactive)
	if err != nil {
		if dpMongoDriver.IsErrNoDocumentFound(err) {
			return nil, ErrNoRecordFound
		}
		return nil, err
	}

	interactive.SetURL(m.Config.PreviewRootURL)

	return &interactive, nil
}

func (m *Mongo) ListInteractives(ctx context.Context, offset, limit int, modelFilter *models.InteractiveFilter) ([]*models.Interactive, int, error) {

	filter := generateFilter(modelFilter)
	f := m.Connection.GetConfiguredCollection().Find(filter).Sort(bson.M{"_id": -1})

	// get total count and paginated values according to provided offset and limit
	var values []*models.Interactive
	totalCount, err := QueryPage(ctx, f, offset, limit, &values)
	if err != nil {
		return values, 0, err
	}

	for _, interactive := range values {
		interactive.SetURL(m.Config.PreviewRootURL)
	}

	return values, totalCount, nil
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
		fil["$and"] = []interface{}{bson.M{"metadata.collection_id": ""}, bson.M{"published": false}}
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
		"$currentDate": bson.M{
			"last_updated": true,
		},
		"$setOnInsert": bson.M{
			"created": time.Now(),
		},
	}

	_, err = m.Connection.GetConfiguredCollection().UpsertById(ctx, id, update)
	return
}

// PatchInteractive patches an existing interactive
func (m *Mongo) PatchInteractive(ctx context.Context, a PatchAction, i *models.Interactive) error {
	log.Info(ctx, "patching interactive", log.Data{"id": i.ID})

	var patch bson.M
	switch a {
	case ImportArchive:
		patch = bson.M{"archive": i.Archive, "import_message": i.ImportMessage, "state": i.State}
	default:
		return fmt.Errorf("unsupported patch action %s", a)
	}

	update := bson.M{
		"$set": patch,
		"$currentDate": bson.M{
			"last_updated": true,
		},
	}
	_, err := m.Connection.GetConfiguredCollection().UpdateById(ctx, i.ID, update)
	return err
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
