package service

import (
	"context"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-interactives-api/api"
	"github.com/ONSdigital/dp-interactives-api/config"
	"github.com/ONSdigital/dp-interactives-api/mongo"
	"github.com/ONSdigital/dp-interactives-api/upload"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	dphttp "github.com/ONSdigital/dp-net/http"
	dps3 "github.com/ONSdigital/dp-s3/v2"
)

type ExternalServiceList struct {
	MongoDB       bool
	HealthCheck   bool
	KafkaProducer bool
	S3Client      bool
	Init          Initialiser
}

func NewServiceList(initialiser Initialiser) *ExternalServiceList {
	return &ExternalServiceList{
		MongoDB:       false,
		HealthCheck:   false,
		KafkaProducer: false,
		S3Client:      false,
		Init:          initialiser,
	}
}

type Init struct{}

// GetHTTPServer creates an http server and sets the Server flag to true
func (e *ExternalServiceList) GetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := e.Init.DoGetHTTPServer(bindAddr, router)
	return s
}

// GetMongoDB creates a mongoDB client and sets the Mongo flag to true
func (e *ExternalServiceList) GetMongoDB(ctx context.Context, cfg *config.Config) (api.MongoServer, error) {
	mongoDB, err := e.Init.DoGetMongoDB(ctx, cfg)
	if err != nil {
		return nil, err
	}
	e.MongoDB = true
	return mongoDB, nil
}

// GetKafkaProducer returns a kafka producer
func (e *ExternalServiceList) GetKafkaProducer(ctx context.Context, cfg *config.Config) (producer kafka.IProducer, err error) {
	producer, err = e.Init.DoGetKafkaProducer(ctx, cfg)
	if err != nil {
		return nil, err
	}
	e.KafkaProducer = true
	return producer, nil
}

// GetS3Uploaded creates a S3 client and sets the S3Uploaded flag to true
func (e *ExternalServiceList) GetS3Client(ctx context.Context, cfg *config.Config) (upload.S3Interface, error) {
	s3, err := e.Init.DoGetS3Client(ctx, cfg)
	if err != nil {
		return nil, err
	}
	e.S3Client = true
	return s3, nil
}

// GetHealthClient returns a healthclient for the provided URL
func (e *ExternalServiceList) GetHealthClient(name, url string) *health.Client {
	return e.Init.DoGetHealthClient(name, url)
}

// GetHealthCheck creates a healthcheck with versionInfo and sets teh HealthCheck flag to true
func (e *ExternalServiceList) GetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	hc, err := e.Init.DoGetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	e.HealthCheck = true
	return hc, nil
}

// GetAuthorisationMiddleware creates a new instance of authorisation.Middlware
func (e *ExternalServiceList) GetAuthorisationMiddleware(ctx context.Context, authorisationConfig *authorisation.Config) (authorisation.Middleware, error) {
	return e.Init.DoGetAuthorisationMiddleware(ctx, authorisationConfig)
}

// DoGetHTTPServer creates an HTTP Server with the provided bind address and router
func (e *Init) DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := dphttp.NewServer(bindAddr, router)
	s.HandleOSSignals = false
	return s
}

// DoGetMongoDB returns a MongoDB
func (e *Init) DoGetMongoDB(ctx context.Context, cfg *config.Config) (api.MongoServer, error) {
	mongodb := &mongo.Mongo{
		Collection: cfg.MongoConfig.Collection,
		Database:   cfg.MongoConfig.Database,
		URI:        cfg.MongoConfig.BindAddr,
		Username:   cfg.MongoConfig.Username,
		Password:   cfg.MongoConfig.Password,
		IsSSL:      cfg.MongoConfig.IsSSL,
	}
	if err := mongodb.Init(ctx, false, true); err != nil {
		return nil, err
	}
	return mongodb, nil
}

// DoGetKafkaProducer creates a kafka producer for the provided broker addresses, topic and envMax values in config
func (e *Init) DoGetKafkaProducer(ctx context.Context, cfg *config.Config) (kafka.IProducer, error) {
	pConfig := &kafka.ProducerConfig{
		KafkaVersion:    &cfg.KafkaVersion,
		MaxMessageBytes: &cfg.KafkaMaxBytes,
		BrokerAddrs:     cfg.Brokers,
		Topic:           cfg.InteractivesWriteTopic,
	}
	if cfg.MinBrokers > 0 {
		pConfig.MinBrokersHealthy = &cfg.MinBrokers
	}
	if cfg.KafkaSecProtocol == "TLS" {
		pConfig.SecurityConfig = kafka.GetSecurityConfig(
			cfg.KafkaSecCACerts,
			cfg.KafkaSecClientCert,
			cfg.KafkaSecClientKey,
			cfg.KafkaSecSkipVerify,
		)
	}
	return kafka.NewProducer(ctx, pConfig)
}

// DoGetS3Uploaded returns a S3Client
func (e *Init) DoGetS3Client(ctx context.Context, cfg *config.Config) (upload.S3Interface, error) {
	if cfg.AwsEndpoint != "" {
		//for local development only - set env var to initialise
		s, err := session.NewSession(&aws.Config{
			Endpoint:         aws.String(cfg.AwsEndpoint),
			Region:           aws.String(cfg.AwsRegion),
			S3ForcePathStyle: aws.Bool(true),
			Credentials:      credentials.NewStaticCredentials("n/a", "n/a", ""),
		})

		if err != nil {
			return nil, err
		}

		return dps3.NewClientWithSession(cfg.UploadBucketName, s), nil
	}

	s3Client, err := dps3.NewClient(cfg.AwsRegion, cfg.UploadBucketName)
	if err != nil {
		return nil, err
	}
	return s3Client, nil
}

// DoGetHealthClient creates a new Health Client for the provided name and url
func (e *Init) DoGetHealthClient(name, url string) *health.Client {
	return health.NewClient(name, url)
}

// DoGetHealthCheck creates a healthcheck with versionInfo
func (e *Init) DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return &hc, nil
}

// DoGetAuthorisationMiddleware creates authorisation middleware for the given config
func (e *Init) DoGetAuthorisationMiddleware(ctx context.Context, authorisationConfig *authorisation.Config) (authorisation.Middleware, error) {
	return authorisation.NewFeatureFlaggedMiddleware(ctx, authorisationConfig)
}
