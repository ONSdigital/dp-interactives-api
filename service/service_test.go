package service_test

import (
	"context"
	"fmt"
	"github.com/ONSdigital/dp-net/v2/responder"
	"net/http"
	"sync"
	"testing"

	"github.com/ONSdigital/dp-interactives-api/models"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	authorisationMock "github.com/ONSdigital/dp-authorisation/v2/authorisation/mock"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-interactives-api/api"
	apiMock "github.com/ONSdigital/dp-interactives-api/api/mock"
	"github.com/ONSdigital/dp-interactives-api/config"
	"github.com/ONSdigital/dp-interactives-api/service"
	serviceMock "github.com/ONSdigital/dp-interactives-api/service/mock"
	"github.com/ONSdigital/dp-interactives-api/upload"
	uploadMock "github.com/ONSdigital/dp-interactives-api/upload/mock"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	"github.com/pkg/errors"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	expectedChecks = 5
)

var (
	ctx              = context.Background()
	testBuildTime    = "BuildTime"
	testGitCommit    = "GitCommit"
	testVersion      = "Version"
	errServer        = errors.New("HTTP Server error")
	errMongoDB       = errors.New("mongoDB error")
	errKafkaProducer = errors.New("kafkaProducer error")
	errHealthcheck   = errors.New("healthCheck error")
	errS3            = errors.New("s3 error")

	noopGen            = func(string) string { return "" }
	funcDoGetGenerator = func() (models.Generator, models.Generator, models.Generator) {
		return noopGen, noopGen, noopGen
	}
	funcDoGetMongoDbErr = func(ctx context.Context, cfg *config.Config) (api.MongoServer, error) {
		return nil, errMongoDB
	}
	funcDoGetHealthcheckErr = func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
		return nil, errHealthcheck
	}
	funcDoGetHTTPServerNil = func(bindAddr string, router http.Handler) service.HTTPServer {
		return nil
	}

	funcDoGetResponder = func(ctx context.Context, cfg *config.Config) (*responder.Responder, error) {
		return responder.New(), nil
	}
)

func TestRun(t *testing.T) {

	Convey("Having a set of mocked dependencies", t, func() {

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		authorisationMiddleware := &authorisationMock.MiddlewareMock{
			RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
				return handlerFunc
			},
			CloseFunc: func(ctx context.Context) error {
				return nil
			},
		}

		mongoDbMock := &apiMock.MongoServerMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		kafkaProducerMock := kafkatest.NewMessageProducer(true)

		hcMock := &serviceMock.HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(ctx context.Context) {},
		}

		serverWg := &sync.WaitGroup{}
		serverMock := &serviceMock.HTTPServerMock{
			ListenAndServeFunc: func() error {
				serverWg.Done()
				return nil
			},
		}

		failingServerMock := &serviceMock.HTTPServerMock{
			ListenAndServeFunc: func() error {
				serverWg.Done()
				return errServer
			},
		}

		s3Mock := &uploadMock.S3InterfaceMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		fsMock := &apiMock.FilesServiceMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		funcDoGetMongoDbOk := func(ctx context.Context, cfg *config.Config) (api.MongoServer, error) {
			return mongoDbMock, nil
		}

		funcDoGetHealthcheckOk := func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
			return hcMock, nil
		}

		funcDoGetHTTPServer := func(bindAddr string, router http.Handler) service.HTTPServer {
			return serverMock
		}

		funcDoGetFailingHTTPSerer := func(bindAddr string, router http.Handler) service.HTTPServer {
			return failingServerMock
		}

		funcDoGetAuthOk := func(ctx context.Context, authorisationConfig *authorisation.Config) (authorisation.Middleware, error) {
			return authorisationMiddleware, nil
		}

		funcDoGetKafkaProducerOk := func(ctx context.Context, cfg *config.Config) (kafka.IProducer, error) {
			return kafkaProducerMock, nil
		}

		funcDoGetKafkaProducerErr := func(ctx context.Context, cfg *config.Config) (kafka.IProducer, error) {
			return nil, errKafkaProducer
		}

		funcDoGetS3Ok := func(ctx context.Context, cfg *config.Config) (upload.S3Interface, error) {
			return s3Mock, nil
		}

		funcDoGetS3Err := func(ctx context.Context, cfg *config.Config) (upload.S3Interface, error) {
			return nil, errS3
		}

		funcDoGetFilesServiceOk := func(ctx context.Context, cfg *config.Config) (api.FilesService, error) {
			return fsMock, nil
		}

		funcDoGetHealthClientOk := func(name string, url string) *health.Client {
			return &health.Client{
				URL:  url,
				Name: name,
			}
		}

		Convey("Given that initialising mongoDB returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:              funcDoGetHTTPServerNil,
				DoGetMongoDBFunc:                 funcDoGetMongoDbErr,
				DoGetKafkaProducerFunc:           funcDoGetKafkaProducerOk,
				DoGetHealthClientFunc:            funcDoGetHealthClientOk,
				DoGetS3ClientFunc:                funcDoGetS3Ok,
				DoGetAuthorisationMiddlewareFunc: funcDoGetAuthOk,
				DoGetGeneratorsFunc:              funcDoGetGenerator,
				DoGetResponderFunc:               funcDoGetResponder,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set. No further initialisations are attempted", func() {
				So(err, ShouldResemble, errMongoDB)
				So(svcList.MongoDB, ShouldBeFalse)
				So(svcList.KafkaProducer, ShouldBeFalse)
				So(svcList.HealthCheck, ShouldBeFalse)
				So(svcList.S3Client, ShouldBeFalse)
			})
		})

		Convey("Given that initialising kafka image-uploaded producer returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:              funcDoGetHTTPServerNil,
				DoGetMongoDBFunc:                 funcDoGetMongoDbOk,
				DoGetKafkaProducerFunc:           funcDoGetKafkaProducerErr,
				DoGetHealthClientFunc:            funcDoGetHealthClientOk,
				DoGetS3ClientFunc:                funcDoGetS3Ok,
				DoGetAuthorisationMiddlewareFunc: funcDoGetAuthOk,
				DoGetGeneratorsFunc:              funcDoGetGenerator,
				DoGetResponderFunc:               funcDoGetResponder,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set. No further initialisations are attempted", func() {
				So(err, ShouldResemble, errKafkaProducer)
				So(svcList.MongoDB, ShouldBeTrue)
				So(svcList.KafkaProducer, ShouldBeFalse)
				So(svcList.HealthCheck, ShouldBeFalse)
				So(svcList.S3Client, ShouldBeTrue)
			})
		})

		Convey("Given that initialising s3 returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:              funcDoGetHTTPServerNil,
				DoGetMongoDBFunc:                 funcDoGetMongoDbOk,
				DoGetKafkaProducerFunc:           funcDoGetKafkaProducerOk,
				DoGetHealthClientFunc:            funcDoGetHealthClientOk,
				DoGetS3ClientFunc:                funcDoGetS3Err,
				DoGetAuthorisationMiddlewareFunc: funcDoGetAuthOk,
				DoGetGeneratorsFunc:              funcDoGetGenerator,
				DoGetResponderFunc:               funcDoGetResponder,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set. No further initialisations are attempted", func() {
				So(err, ShouldResemble, errS3)
				So(svcList.MongoDB, ShouldBeTrue)
				So(svcList.KafkaProducer, ShouldBeFalse)
				So(svcList.HealthCheck, ShouldBeFalse)
				So(svcList.S3Client, ShouldBeFalse)
			})
		})

		Convey("Given that initialising healthcheck returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:              funcDoGetHTTPServerNil,
				DoGetMongoDBFunc:                 funcDoGetMongoDbOk,
				DoGetHealthCheckFunc:             funcDoGetHealthcheckErr,
				DoGetKafkaProducerFunc:           funcDoGetKafkaProducerOk,
				DoGetHealthClientFunc:            funcDoGetHealthClientOk,
				DoGetS3ClientFunc:                funcDoGetS3Ok,
				DoGetAuthorisationMiddlewareFunc: funcDoGetAuthOk,
				DoGetGeneratorsFunc:              funcDoGetGenerator,
				DoGetFilesServiceFunc:            funcDoGetFilesServiceOk,
				DoGetResponderFunc:               funcDoGetResponder,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set", func() {
				So(err, ShouldResemble, errHealthcheck)
				So(svcList.MongoDB, ShouldBeTrue)
				So(svcList.KafkaProducer, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeFalse)
				So(svcList.S3Client, ShouldBeTrue)
			})
		})

		Convey("Given that Checkers cannot be registered", func() {

			errAddheckFail := errors.New("Error(s) registering checkers for healthcheck")
			hcMockAddFail := &serviceMock.HealthCheckerMock{
				AddCheckFunc: func(name string, checker healthcheck.Checker) error { return errAddheckFail },
				StartFunc:    func(ctx context.Context) {},
			}

			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:    funcDoGetHTTPServerNil,
				DoGetMongoDBFunc:       funcDoGetMongoDbOk,
				DoGetKafkaProducerFunc: funcDoGetKafkaProducerOk,
				DoGetS3ClientFunc:      funcDoGetS3Ok,
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
					return hcMockAddFail, nil
				},
				DoGetHealthClientFunc:            funcDoGetHealthClientOk,
				DoGetAuthorisationMiddlewareFunc: funcDoGetAuthOk,
				DoGetGeneratorsFunc:              funcDoGetGenerator,
				DoGetFilesServiceFunc:            funcDoGetFilesServiceOk,
				DoGetResponderFunc:               funcDoGetResponder,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails, but all checks try to register", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, fmt.Sprintf("unable to register checkers: %s", errAddheckFail.Error()))
				So(svcList.MongoDB, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeTrue)
				So(hcMockAddFail.AddCheckCalls(), ShouldHaveLength, expectedChecks)
				So(hcMockAddFail.AddCheckCalls()[0].Name, ShouldResemble, "Mongo DB")
				So(hcMockAddFail.AddCheckCalls()[1].Name, ShouldResemble, "Uploaded Kafka Producer")
				So(hcMockAddFail.AddCheckCalls()[2].Name, ShouldResemble, "S3 checker")
				So(hcMockAddFail.AddCheckCalls()[3].Name, ShouldResemble, "FilesService checker")
				So(hcMockAddFail.AddCheckCalls()[4].Name, ShouldResemble, "permissions cache health check")
			})
		})

		Convey("Given that all dependencies are successfully initialised", func() {

			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:              funcDoGetHTTPServer,
				DoGetMongoDBFunc:                 funcDoGetMongoDbOk,
				DoGetKafkaProducerFunc:           funcDoGetKafkaProducerOk,
				DoGetHealthCheckFunc:             funcDoGetHealthcheckOk,
				DoGetHealthClientFunc:            funcDoGetHealthClientOk,
				DoGetS3ClientFunc:                funcDoGetS3Ok,
				DoGetAuthorisationMiddlewareFunc: funcDoGetAuthOk,
				DoGetGeneratorsFunc:              funcDoGetGenerator,
				DoGetFilesServiceFunc:            funcDoGetFilesServiceOk,
				DoGetResponderFunc:               funcDoGetResponder,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			serverWg.Add(1)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run succeeds and all the flags are set", func() {
				So(err, ShouldBeNil)
				So(svcList.MongoDB, ShouldBeTrue)
				So(svcList.KafkaProducer, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeTrue)
				So(svcList.S3Client, ShouldBeTrue)
			})

			Convey("The checkers are registered and the healthcheck and http server started", func() {
				So(hcMock.AddCheckCalls(), ShouldHaveLength, expectedChecks)
				So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "Mongo DB")
				So(hcMock.AddCheckCalls()[1].Name, ShouldResemble, "Uploaded Kafka Producer")
				So(hcMock.AddCheckCalls()[2].Name, ShouldResemble, "S3 checker")
				So(hcMock.AddCheckCalls()[3].Name, ShouldResemble, "FilesService checker")
				So(hcMock.AddCheckCalls()[4].Name, ShouldResemble, "permissions cache health check")
				So(initMock.DoGetHTTPServerCalls(), ShouldHaveLength, 1)
				So(initMock.DoGetHTTPServerCalls()[0].BindAddr, ShouldEqual, ":27500")
				So(hcMock.StartCalls(), ShouldHaveLength, 1)
				serverWg.Wait() // Wait for HTTP server go-routine to finish
				So(serverMock.ListenAndServeCalls(), ShouldHaveLength, 1)
			})
		})

		Convey("Given that all dependencies are successfully initialised but the http server fails", func() {

			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:              funcDoGetFailingHTTPSerer,
				DoGetMongoDBFunc:                 funcDoGetMongoDbOk,
				DoGetKafkaProducerFunc:           funcDoGetKafkaProducerOk,
				DoGetHealthCheckFunc:             funcDoGetHealthcheckOk,
				DoGetHealthClientFunc:            funcDoGetHealthClientOk,
				DoGetS3ClientFunc:                funcDoGetS3Ok,
				DoGetAuthorisationMiddlewareFunc: funcDoGetAuthOk,
				DoGetGeneratorsFunc:              funcDoGetGenerator,
				DoGetFilesServiceFunc:            funcDoGetFilesServiceOk,
				DoGetResponderFunc:               funcDoGetResponder,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			serverWg.Add(1)
			_, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)
			So(err, ShouldBeNil)

			Convey("Then the error is returned in the error channel", func() {
				sErr := <-svcErrors
				So(sErr.Error(), ShouldResemble, fmt.Sprintf("failure in http listen and serve: %s", errServer.Error()))
				So(failingServerMock.ListenAndServeCalls(), ShouldHaveLength, 1)
			})
		})

	})
}

func TestClose(t *testing.T) {

	Convey("Having a correctly initialised service", t, func() {

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		hcStopped := false
		serverStopped := false
		mongoStopped := false

		authorisationMiddleware := &authorisationMock.MiddlewareMock{
			RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
				return handlerFunc
			},
			CloseFunc: func(ctx context.Context) error {
				return nil
			},
		}

		// healthcheck Stop does not depend on any other service being closed/stopped
		hcMock := &serviceMock.HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(ctx context.Context) {},
			StopFunc:     func() { hcStopped = true },
		}

		// server Shutdown will fail if healthcheck is not stopped
		serverMock := &serviceMock.HTTPServerMock{
			ListenAndServeFunc: func() error { return nil },
			ShutdownFunc: func(ctx context.Context) error {
				if !hcStopped {
					return errors.New("Server stopped before healthcheck")
				}
				serverStopped = true
				return nil
			},
		}

		// mongoDB Close will fail if healthcheck and http server are not already closed
		mongoDbMock := &apiMock.MongoServerMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
			CloseFunc: func(ctx context.Context) error {
				if !hcStopped || !serverStopped {
					return errors.New("MongoDB closed before stopping healthcheck or HTTP server")
				}
				mongoStopped = true
				return nil
			},
		}

		s3Mock := &uploadMock.S3InterfaceMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		fsMock := &apiMock.FilesServiceMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		// kafkaProducerMock will fail if healthcheck, http server and mongo are not already closed
		kafkaProducerMock := &kafkatest.IProducerMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
			CloseFunc: func(ctx context.Context) error {
				if !hcStopped || !serverStopped || !mongoStopped {
					return errors.New("KafkaProducer closed before stopping healthcheck, MongoDB or HTTP server")
				}
				return nil
			},
			ChannelsFunc: func() *kafka.ProducerChannels { return kafka.CreateProducerChannels() },
		}

		Convey("Closing the service results in all the dependencies being closed in the expected order", func() {

			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:    func(bindAddr string, router http.Handler) service.HTTPServer { return serverMock },
				DoGetMongoDBFunc:       func(ctx context.Context, cfg *config.Config) (api.MongoServer, error) { return mongoDbMock, nil },
				DoGetKafkaProducerFunc: func(ctx context.Context, cfg *config.Config) (kafka.IProducer, error) { return kafkaProducerMock, nil },
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
					return hcMock, nil
				},
				DoGetHealthClientFunc: func(name, url string) *health.Client { return &health.Client{} },
				DoGetS3ClientFunc:     func(ctx context.Context, cfg *config.Config) (upload.S3Interface, error) { return s3Mock, nil },
				DoGetAuthorisationMiddlewareFunc: func(ctx context.Context, authorisationConfig *authorisation.Config) (authorisation.Middleware, error) {
					return authorisationMiddleware, nil
				},
				DoGetGeneratorsFunc:   funcDoGetGenerator,
				DoGetFilesServiceFunc: func(ctx context.Context, cfg *config.Config) (api.FilesService, error) { return fsMock, nil },
				DoGetResponderFunc:    funcDoGetResponder,
			}

			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)
			So(err, ShouldBeNil)

			err = svc.Close(context.Background())
			So(err, ShouldBeNil)
			So(hcMock.StopCalls(), ShouldHaveLength, 1)
			So(serverMock.ShutdownCalls(), ShouldHaveLength, 1)
			So(mongoDbMock.CloseCalls(), ShouldHaveLength, 1)
			So(kafkaProducerMock.CloseCalls(), ShouldHaveLength, 1)
		})

		Convey("If services fail to stop, the Close operation tries to close all dependencies and returns an error", func() {

			failingserverMock := &serviceMock.HTTPServerMock{
				ListenAndServeFunc: func() error { return nil },
				ShutdownFunc: func(ctx context.Context) error {
					return errors.New("Failed to stop http server")
				},
			}

			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:    func(bindAddr string, router http.Handler) service.HTTPServer { return failingserverMock },
				DoGetMongoDBFunc:       func(ctx context.Context, cfg *config.Config) (api.MongoServer, error) { return mongoDbMock, nil },
				DoGetKafkaProducerFunc: func(ctx context.Context, cfg *config.Config) (kafka.IProducer, error) { return kafkaProducerMock, nil },
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
					return hcMock, nil
				},
				DoGetHealthClientFunc: func(name, url string) *health.Client { return &health.Client{} },
				DoGetS3ClientFunc:     func(ctx context.Context, cfg *config.Config) (upload.S3Interface, error) { return s3Mock, nil },
				DoGetAuthorisationMiddlewareFunc: func(ctx context.Context, authorisationConfig *authorisation.Config) (authorisation.Middleware, error) {
					return authorisationMiddleware, nil
				},
				DoGetGeneratorsFunc:   funcDoGetGenerator,
				DoGetFilesServiceFunc: func(ctx context.Context, cfg *config.Config) (api.FilesService, error) { return fsMock, nil },
				DoGetResponderFunc:    funcDoGetResponder,
			}

			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc, err := service.Run(ctx, cfg, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)
			So(err, ShouldBeNil)

			err = svc.Close(context.Background())
			So(err, ShouldNotBeNil)
			So(hcMock.StopCalls(), ShouldHaveLength, 1)
			So(failingserverMock.ShutdownCalls(), ShouldHaveLength, 1)
			So(mongoDbMock.CloseCalls(), ShouldHaveLength, 1)
			So(kafkaProducerMock.CloseCalls(), ShouldHaveLength, 1)
		})
	})
}
