package steps

import (
	"context"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-authorisation/v2/authorisationtest"
	"github.com/ONSdigital/dp-authorisation/v2/permissions"
	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-component-test/utils"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-interactives-api/api"
	apiMock "github.com/ONSdigital/dp-interactives-api/api/mock"
	"github.com/ONSdigital/dp-interactives-api/config"
	"github.com/ONSdigital/dp-interactives-api/internal/data"
	"github.com/ONSdigital/dp-interactives-api/mongo"
	"github.com/ONSdigital/dp-interactives-api/service"
	serviceMock "github.com/ONSdigital/dp-interactives-api/service/mock"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/ONSdigital/dp-net/v2/responder"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	uuid "github.com/satori/go.uuid"
	"net/http"
)

type InteractivesApiComponent struct {
	componenttest.ErrorFeature
	ApiFeature     *componenttest.APIFeature
	svc            *service.Service
	errorChan      chan error
	MongoClient    *mongo.Mongo
	Config         *config.Config
	HTTPServer     *http.Server
	filesService   api.FilesService
	ServiceRunning bool
	initialiser    service.Initialiser
}

func setupFakePermissionsAPI() *authorisationtest.FakePermissionsAPI {
	fakePermissionsAPI := authorisationtest.NewFakePermissionsAPI()
	bundle := getPermissionsBundle()
	fakePermissionsAPI.Reset()
	fakePermissionsAPI.UpdatePermissionsBundleResponse(bundle)
	return fakePermissionsAPI
}

func getPermissionsBundle() *permissions.Bundle {
	return &permissions.Bundle{
		"interactives:create": { // role
			"groups/role-admin": { // group
				{
					ID: "1", // policy
				},
			},
		},
		"interactives:read": { // role
			"groups/role-admin": { // group
				{
					ID: "2", // policy
				},
			},
		},
		"interactives:update": { // role
			"groups/role-admin": { // group
				{
					ID: "2", // policy
				},
			},
		},
		"interactives:delete": { // role
			"groups/role-admin": { // group
				{
					ID: "2", // policy
				},
			},
		},
	}
}

func NewInteractivesApiComponent(mongoURI string) (*InteractivesApiComponent, error) {
	c := &InteractivesApiComponent{
		HTTPServer:     &http.Server{},
		errorChan:      make(chan error),
		ServiceRunning: false,
	}

	var err error

	c.Config, err = config.Get()
	if err != nil {
		return nil, err
	}

	log.Info(context.Background(), "configuration for component test", log.Data{"config": c.Config})

	cfg, _ := config.Get()

	mongodb := &mongo.Mongo{
		PreviewRootURL: "http://preview_url",
		MongoConfig: config.MongoConfig{
			MongoDriverConfig: mongodriver.MongoDriverConfig{
				ClusterEndpoint: mongoURI,
				Database:        utils.RandomDatabase(),
				Collections:     cfg.MongoConfig.Collections,
				ConnectTimeout:  cfg.MongoConfig.ConnectTimeout,
				QueryTimeout:    cfg.MongoConfig.QueryTimeout,
			},
		}}

	if err := mongodb.Init(context.Background()); err != nil {
		return nil, err
	}

	c.MongoClient = mongodb

	cfg.AuthorisationConfig.PermissionsAPIURL = setupFakePermissionsAPI().URL()

	c.filesService = &apiMock.FilesServiceMock{
		SetCollectionIDFunc: func(ctx context.Context, file string, collectionID string) error { return nil },
	}

	return c, nil
}

func (c *InteractivesApiComponent) Reset() error {
	ctx := context.Background()

	if err := c.MongoClient.Init(ctx); err != nil {
		log.Warn(ctx, "error initialising MongoClient during Reset", log.Data{"err": err.Error()})
	}
	c.setInitialiserMock()
	return nil
}

func (c *InteractivesApiComponent) Close() error {
	ctx := context.Background()
	if c.svc != nil && c.ServiceRunning {
		if err := c.MongoClient.Connection.DropDatabase(ctx); err != nil {
			log.Warn(ctx, "error dropping database on Close", log.Data{"err": err.Error()})
		}
		if err := c.svc.Close(ctx); err != nil {
			return err
		}
		c.ServiceRunning = false
	}
	return nil
}

func (c *InteractivesApiComponent) InitialiseService() (http.Handler, error) {
	var svc *service.Service
	var err error

	if svc, err = service.Run(context.Background(), c.Config, service.NewServiceList(c.initialiser), "1", "", "", c.errorChan); err != nil {
		return nil, err
	}
	c.svc = svc
	c.ServiceRunning = true
	return c.HTTPServer.Handler, nil
}

func (f *InteractivesApiComponent) DoGetHealthcheckOk(_ *config.Config, _ string, _ string, _ string) (service.HealthChecker, error) {
	return &serviceMock.HealthCheckerMock{
		AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
		StartFunc:    func(ctx context.Context) {},
		StopFunc:     func() {},
	}, nil
}

func (c *InteractivesApiComponent) DoGetHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	c.HTTPServer.Addr = bindAddr
	c.HTTPServer.Handler = router
	return c.HTTPServer
}

func (c *InteractivesApiComponent) DoGetMockedKafkaProducerOk(ctx context.Context, cfg *config.Config) (kafka.IProducer, error) {
	return &kafkatest.IProducerMock{
		ChannelsFunc: func() *kafka.ProducerChannels {
			return &kafka.ProducerChannels{}
		},
		CloseFunc: func(ctx context.Context) error { return nil },
	}, nil
}

func (f *InteractivesApiComponent) DoGetMongoDB(_ context.Context, _ *config.Config) (api.MongoServer, error) {
	return f.MongoClient, nil
}

func (f *InteractivesApiComponent) DoS3Client(_ context.Context, _ *config.Config) (api.S3Interface, error) {
	return &apiMock.S3InterfaceMock{
		CheckerFunc:        func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		ValidateBucketFunc: func() error { return nil },
		UploadFunc: func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
			return nil, nil
		},
	}, nil
}

func (f *InteractivesApiComponent) DoGetAuthorisationMiddleware(ctx context.Context, cfg *authorisation.Config) (authorisation.Middleware, error) {
	middleware, err := authorisation.NewMiddlewareFromConfig(ctx, cfg, cfg.JWTVerificationPublicKeys)
	if err != nil {
		return nil, err
	}
	return middleware, nil
}

func (f *InteractivesApiComponent) DoGetGenerators() (data.Generator, data.Generator, data.Generator) {
	emptyUUID := func(string) string { return uuid.Nil.String() }
	fakeResourceID := func(string) string { return "AbcdE123" }
	noopSlug := func(in string) string { return in }
	return emptyUUID, fakeResourceID, noopSlug
}

func (f *InteractivesApiComponent) DoGetFSClient(ctx context.Context, cfg *config.Config) (api.FilesService, error) {
	return f.filesService, nil
}

func (f *InteractivesApiComponent) DoGetResponder(_ context.Context, _ *config.Config) (*responder.Responder, error) {
	return responder.New(), nil
}

func (c *InteractivesApiComponent) setInitialiserMock() {
	c.initialiser = &serviceMock.InitialiserMock{
		DoGetMongoDBFunc:                 c.DoGetMongoDB,
		DoGetKafkaProducerFunc:           c.DoGetMockedKafkaProducerOk,
		DoGetHealthCheckFunc:             c.DoGetHealthcheckOk,
		DoGetHTTPServerFunc:              c.DoGetHTTPServer,
		DoGetS3ClientFunc:                c.DoS3Client,
		DoGetAuthorisationMiddlewareFunc: c.DoGetAuthorisationMiddleware,
		DoGetGeneratorsFunc:              c.DoGetGenerators,
		DoGetFilesServiceFunc:            c.DoGetFSClient,
		DoGetResponderFunc:               c.DoGetResponder,
	}
}
