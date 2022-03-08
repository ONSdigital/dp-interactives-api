package steps

import (
	"context"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"net/http"
	"strings"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-interactives-api/api"
	"github.com/ONSdigital/dp-interactives-api/config"
	"github.com/ONSdigital/dp-interactives-api/mongo"
	"github.com/ONSdigital/dp-interactives-api/service"
	serviceMock "github.com/ONSdigital/dp-interactives-api/service/mock"
	"github.com/ONSdigital/dp-interactives-api/upload"
	uploadMock "github.com/ONSdigital/dp-interactives-api/upload/mock"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	"github.com/ONSdigital/log.go/v2/log"
)

type InteractivesApiComponent struct {
	componenttest.ErrorFeature
	ApiFeature     *componenttest.APIFeature
	svc            *service.Service
	errorChan      chan error
	MongoClient    *mongo.Mongo
	Config         *config.Config
	HTTPServer     *http.Server
	ServiceRunning bool
	initialiser    service.Initialiser
	s3             uploadMock.S3InterfaceMock
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
		URI:        strings.Replace(mongoURI, "mongodb://", "", 1),
		Database:   cfg.MongoConfig.Database,
		Collection: cfg.MongoConfig.Collection,
	}

	if err := mongodb.Init(context.Background(), false, true); err != nil {
		return nil, err
	}

	c.MongoClient = mongodb

	return c, nil
}

func (c *InteractivesApiComponent) Reset() error {
	ctx := context.Background()

	c.MongoClient.Database = "interactives-api"
	if err := c.MongoClient.Init(ctx, false, true); err != nil {
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

func (f *InteractivesApiComponent) DoS3Client(_ context.Context, _ *config.Config) (upload.S3Interface, error) {
	return &uploadMock.S3InterfaceMock{
		CheckerFunc:        func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		ValidateBucketFunc: func() error { return nil },
		UploadFunc: func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
			return nil, nil
		},
	}, nil
}

func (c *InteractivesApiComponent) setInitialiserMock() {
	c.initialiser = &serviceMock.InitialiserMock{
		DoGetMongoDBFunc:       c.DoGetMongoDB,
		DoGetKafkaProducerFunc: c.DoGetMockedKafkaProducerOk,
		DoGetHealthCheckFunc:   c.DoGetHealthcheckOk,
		DoGetHTTPServerFunc:    c.DoGetHTTPServer,
		DoGetS3ClientFunc:      c.DoS3Client,
	}
}
