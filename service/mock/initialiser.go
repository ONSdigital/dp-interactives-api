// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-interactives-api/api"
	"github.com/ONSdigital/dp-interactives-api/config"
	"github.com/ONSdigital/dp-interactives-api/models"
	"github.com/ONSdigital/dp-interactives-api/service"
	"github.com/ONSdigital/dp-interactives-api/upload"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"net/http"
	"sync"
)

// Ensure, that InitialiserMock does implement service.Initialiser.
// If this is not the case, regenerate this file with moq.
var _ service.Initialiser = &InitialiserMock{}

// InitialiserMock is a mock implementation of service.Initialiser.
//
// 	func TestSomethingThatUsesInitialiser(t *testing.T) {
//
// 		// make and configure a mocked service.Initialiser
// 		mockedInitialiser := &InitialiserMock{
// 			DoGetAuthorisationMiddlewareFunc: func(ctx context.Context, authorisationConfig *authorisation.Config) (authorisation.Middleware, error) {
// 				panic("mock out the DoGetAuthorisationMiddleware method")
// 			},
// 			DoGetFilesServiceFunc: func(ctx context.Context, cfg *config.Config) (api.FilesService, error) {
// 				panic("mock out the DoGetFilesService method")
// 			},
// 			DoGetGeneratorsFunc: func() (models.Generator, models.Generator, models.Generator) {
// 				panic("mock out the DoGetGenerators method")
// 			},
// 			DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.HTTPServer {
// 				panic("mock out the DoGetHTTPServer method")
// 			},
// 			DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
// 				panic("mock out the DoGetHealthCheck method")
// 			},
// 			DoGetHealthClientFunc: func(name string, url string) *health.Client {
// 				panic("mock out the DoGetHealthClient method")
// 			},
// 			DoGetKafkaProducerFunc: func(ctx context.Context, cfg *config.Config) (kafka.IProducer, error) {
// 				panic("mock out the DoGetKafkaProducer method")
// 			},
// 			DoGetMongoDBFunc: func(ctx context.Context, cfg *config.Config) (api.MongoServer, error) {
// 				panic("mock out the DoGetMongoDB method")
// 			},
// 			DoGetS3ClientFunc: func(ctx context.Context, cfg *config.Config) (upload.S3Interface, error) {
// 				panic("mock out the DoGetS3Client method")
// 			},
// 		}
//
// 		// use mockedInitialiser in code that requires service.Initialiser
// 		// and then make assertions.
//
// 	}
type InitialiserMock struct {
	// DoGetAuthorisationMiddlewareFunc mocks the DoGetAuthorisationMiddleware method.
	DoGetAuthorisationMiddlewareFunc func(ctx context.Context, authorisationConfig *authorisation.Config) (authorisation.Middleware, error)

	// DoGetFilesServiceFunc mocks the DoGetFilesService method.
	DoGetFilesServiceFunc func(ctx context.Context, cfg *config.Config) (api.FilesService, error)

	// DoGetGeneratorsFunc mocks the DoGetGenerators method.
	DoGetGeneratorsFunc func() (models.Generator, models.Generator, models.Generator)

	// DoGetHTTPServerFunc mocks the DoGetHTTPServer method.
	DoGetHTTPServerFunc func(bindAddr string, router http.Handler) service.HTTPServer

	// DoGetHealthCheckFunc mocks the DoGetHealthCheck method.
	DoGetHealthCheckFunc func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error)

	// DoGetHealthClientFunc mocks the DoGetHealthClient method.
	DoGetHealthClientFunc func(name string, url string) *health.Client

	// DoGetKafkaProducerFunc mocks the DoGetKafkaProducer method.
	DoGetKafkaProducerFunc func(ctx context.Context, cfg *config.Config) (kafka.IProducer, error)

	// DoGetMongoDBFunc mocks the DoGetMongoDB method.
	DoGetMongoDBFunc func(ctx context.Context, cfg *config.Config) (api.MongoServer, error)

	// DoGetS3ClientFunc mocks the DoGetS3Client method.
	DoGetS3ClientFunc func(ctx context.Context, cfg *config.Config) (upload.S3Interface, error)

	// calls tracks calls to the methods.
	calls struct {
		// DoGetAuthorisationMiddleware holds details about calls to the DoGetAuthorisationMiddleware method.
		DoGetAuthorisationMiddleware []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// AuthorisationConfig is the authorisationConfig argument value.
			AuthorisationConfig *authorisation.Config
		}
		// DoGetFilesService holds details about calls to the DoGetFilesService method.
		DoGetFilesService []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Cfg is the cfg argument value.
			Cfg *config.Config
		}
		// DoGetGenerators holds details about calls to the DoGetGenerators method.
		DoGetGenerators []struct {
		}
		// DoGetHTTPServer holds details about calls to the DoGetHTTPServer method.
		DoGetHTTPServer []struct {
			// BindAddr is the bindAddr argument value.
			BindAddr string
			// Router is the router argument value.
			Router http.Handler
		}
		// DoGetHealthCheck holds details about calls to the DoGetHealthCheck method.
		DoGetHealthCheck []struct {
			// Cfg is the cfg argument value.
			Cfg *config.Config
			// BuildTime is the buildTime argument value.
			BuildTime string
			// GitCommit is the gitCommit argument value.
			GitCommit string
			// Version is the version argument value.
			Version string
		}
		// DoGetHealthClient holds details about calls to the DoGetHealthClient method.
		DoGetHealthClient []struct {
			// Name is the name argument value.
			Name string
			// URL is the url argument value.
			URL string
		}
		// DoGetKafkaProducer holds details about calls to the DoGetKafkaProducer method.
		DoGetKafkaProducer []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Cfg is the cfg argument value.
			Cfg *config.Config
		}
		// DoGetMongoDB holds details about calls to the DoGetMongoDB method.
		DoGetMongoDB []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Cfg is the cfg argument value.
			Cfg *config.Config
		}
		// DoGetS3Client holds details about calls to the DoGetS3Client method.
		DoGetS3Client []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Cfg is the cfg argument value.
			Cfg *config.Config
		}
	}
	lockDoGetAuthorisationMiddleware sync.RWMutex
	lockDoGetFilesService            sync.RWMutex
	lockDoGetGenerators              sync.RWMutex
	lockDoGetHTTPServer              sync.RWMutex
	lockDoGetHealthCheck             sync.RWMutex
	lockDoGetHealthClient            sync.RWMutex
	lockDoGetKafkaProducer           sync.RWMutex
	lockDoGetMongoDB                 sync.RWMutex
	lockDoGetS3Client                sync.RWMutex
}

// DoGetAuthorisationMiddleware calls DoGetAuthorisationMiddlewareFunc.
func (mock *InitialiserMock) DoGetAuthorisationMiddleware(ctx context.Context, authorisationConfig *authorisation.Config) (authorisation.Middleware, error) {
	if mock.DoGetAuthorisationMiddlewareFunc == nil {
		panic("InitialiserMock.DoGetAuthorisationMiddlewareFunc: method is nil but Initialiser.DoGetAuthorisationMiddleware was just called")
	}
	callInfo := struct {
		Ctx                 context.Context
		AuthorisationConfig *authorisation.Config
	}{
		Ctx:                 ctx,
		AuthorisationConfig: authorisationConfig,
	}
	mock.lockDoGetAuthorisationMiddleware.Lock()
	mock.calls.DoGetAuthorisationMiddleware = append(mock.calls.DoGetAuthorisationMiddleware, callInfo)
	mock.lockDoGetAuthorisationMiddleware.Unlock()
	return mock.DoGetAuthorisationMiddlewareFunc(ctx, authorisationConfig)
}

// DoGetAuthorisationMiddlewareCalls gets all the calls that were made to DoGetAuthorisationMiddleware.
// Check the length with:
//     len(mockedInitialiser.DoGetAuthorisationMiddlewareCalls())
func (mock *InitialiserMock) DoGetAuthorisationMiddlewareCalls() []struct {
	Ctx                 context.Context
	AuthorisationConfig *authorisation.Config
} {
	var calls []struct {
		Ctx                 context.Context
		AuthorisationConfig *authorisation.Config
	}
	mock.lockDoGetAuthorisationMiddleware.RLock()
	calls = mock.calls.DoGetAuthorisationMiddleware
	mock.lockDoGetAuthorisationMiddleware.RUnlock()
	return calls
}

// DoGetFilesService calls DoGetFilesServiceFunc.
func (mock *InitialiserMock) DoGetFilesService(ctx context.Context, cfg *config.Config) (api.FilesService, error) {
	if mock.DoGetFilesServiceFunc == nil {
		panic("InitialiserMock.DoGetFilesServiceFunc: method is nil but Initialiser.DoGetFilesService was just called")
	}
	callInfo := struct {
		Ctx context.Context
		Cfg *config.Config
	}{
		Ctx: ctx,
		Cfg: cfg,
	}
	mock.lockDoGetFilesService.Lock()
	mock.calls.DoGetFilesService = append(mock.calls.DoGetFilesService, callInfo)
	mock.lockDoGetFilesService.Unlock()
	return mock.DoGetFilesServiceFunc(ctx, cfg)
}

// DoGetFilesServiceCalls gets all the calls that were made to DoGetFilesService.
// Check the length with:
//     len(mockedInitialiser.DoGetFilesServiceCalls())
func (mock *InitialiserMock) DoGetFilesServiceCalls() []struct {
	Ctx context.Context
	Cfg *config.Config
} {
	var calls []struct {
		Ctx context.Context
		Cfg *config.Config
	}
	mock.lockDoGetFilesService.RLock()
	calls = mock.calls.DoGetFilesService
	mock.lockDoGetFilesService.RUnlock()
	return calls
}

// DoGetGenerators calls DoGetGeneratorsFunc.
func (mock *InitialiserMock) DoGetGenerators() (models.Generator, models.Generator, models.Generator) {
	if mock.DoGetGeneratorsFunc == nil {
		panic("InitialiserMock.DoGetGeneratorsFunc: method is nil but Initialiser.DoGetGenerators was just called")
	}
	callInfo := struct {
	}{}
	mock.lockDoGetGenerators.Lock()
	mock.calls.DoGetGenerators = append(mock.calls.DoGetGenerators, callInfo)
	mock.lockDoGetGenerators.Unlock()
	return mock.DoGetGeneratorsFunc()
}

// DoGetGeneratorsCalls gets all the calls that were made to DoGetGenerators.
// Check the length with:
//     len(mockedInitialiser.DoGetGeneratorsCalls())
func (mock *InitialiserMock) DoGetGeneratorsCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockDoGetGenerators.RLock()
	calls = mock.calls.DoGetGenerators
	mock.lockDoGetGenerators.RUnlock()
	return calls
}

// DoGetHTTPServer calls DoGetHTTPServerFunc.
func (mock *InitialiserMock) DoGetHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	if mock.DoGetHTTPServerFunc == nil {
		panic("InitialiserMock.DoGetHTTPServerFunc: method is nil but Initialiser.DoGetHTTPServer was just called")
	}
	callInfo := struct {
		BindAddr string
		Router   http.Handler
	}{
		BindAddr: bindAddr,
		Router:   router,
	}
	mock.lockDoGetHTTPServer.Lock()
	mock.calls.DoGetHTTPServer = append(mock.calls.DoGetHTTPServer, callInfo)
	mock.lockDoGetHTTPServer.Unlock()
	return mock.DoGetHTTPServerFunc(bindAddr, router)
}

// DoGetHTTPServerCalls gets all the calls that were made to DoGetHTTPServer.
// Check the length with:
//     len(mockedInitialiser.DoGetHTTPServerCalls())
func (mock *InitialiserMock) DoGetHTTPServerCalls() []struct {
	BindAddr string
	Router   http.Handler
} {
	var calls []struct {
		BindAddr string
		Router   http.Handler
	}
	mock.lockDoGetHTTPServer.RLock()
	calls = mock.calls.DoGetHTTPServer
	mock.lockDoGetHTTPServer.RUnlock()
	return calls
}

// DoGetHealthCheck calls DoGetHealthCheckFunc.
func (mock *InitialiserMock) DoGetHealthCheck(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
	if mock.DoGetHealthCheckFunc == nil {
		panic("InitialiserMock.DoGetHealthCheckFunc: method is nil but Initialiser.DoGetHealthCheck was just called")
	}
	callInfo := struct {
		Cfg       *config.Config
		BuildTime string
		GitCommit string
		Version   string
	}{
		Cfg:       cfg,
		BuildTime: buildTime,
		GitCommit: gitCommit,
		Version:   version,
	}
	mock.lockDoGetHealthCheck.Lock()
	mock.calls.DoGetHealthCheck = append(mock.calls.DoGetHealthCheck, callInfo)
	mock.lockDoGetHealthCheck.Unlock()
	return mock.DoGetHealthCheckFunc(cfg, buildTime, gitCommit, version)
}

// DoGetHealthCheckCalls gets all the calls that were made to DoGetHealthCheck.
// Check the length with:
//     len(mockedInitialiser.DoGetHealthCheckCalls())
func (mock *InitialiserMock) DoGetHealthCheckCalls() []struct {
	Cfg       *config.Config
	BuildTime string
	GitCommit string
	Version   string
} {
	var calls []struct {
		Cfg       *config.Config
		BuildTime string
		GitCommit string
		Version   string
	}
	mock.lockDoGetHealthCheck.RLock()
	calls = mock.calls.DoGetHealthCheck
	mock.lockDoGetHealthCheck.RUnlock()
	return calls
}

// DoGetHealthClient calls DoGetHealthClientFunc.
func (mock *InitialiserMock) DoGetHealthClient(name string, url string) *health.Client {
	if mock.DoGetHealthClientFunc == nil {
		panic("InitialiserMock.DoGetHealthClientFunc: method is nil but Initialiser.DoGetHealthClient was just called")
	}
	callInfo := struct {
		Name string
		URL  string
	}{
		Name: name,
		URL:  url,
	}
	mock.lockDoGetHealthClient.Lock()
	mock.calls.DoGetHealthClient = append(mock.calls.DoGetHealthClient, callInfo)
	mock.lockDoGetHealthClient.Unlock()
	return mock.DoGetHealthClientFunc(name, url)
}

// DoGetHealthClientCalls gets all the calls that were made to DoGetHealthClient.
// Check the length with:
//     len(mockedInitialiser.DoGetHealthClientCalls())
func (mock *InitialiserMock) DoGetHealthClientCalls() []struct {
	Name string
	URL  string
} {
	var calls []struct {
		Name string
		URL  string
	}
	mock.lockDoGetHealthClient.RLock()
	calls = mock.calls.DoGetHealthClient
	mock.lockDoGetHealthClient.RUnlock()
	return calls
}

// DoGetKafkaProducer calls DoGetKafkaProducerFunc.
func (mock *InitialiserMock) DoGetKafkaProducer(ctx context.Context, cfg *config.Config) (kafka.IProducer, error) {
	if mock.DoGetKafkaProducerFunc == nil {
		panic("InitialiserMock.DoGetKafkaProducerFunc: method is nil but Initialiser.DoGetKafkaProducer was just called")
	}
	callInfo := struct {
		Ctx context.Context
		Cfg *config.Config
	}{
		Ctx: ctx,
		Cfg: cfg,
	}
	mock.lockDoGetKafkaProducer.Lock()
	mock.calls.DoGetKafkaProducer = append(mock.calls.DoGetKafkaProducer, callInfo)
	mock.lockDoGetKafkaProducer.Unlock()
	return mock.DoGetKafkaProducerFunc(ctx, cfg)
}

// DoGetKafkaProducerCalls gets all the calls that were made to DoGetKafkaProducer.
// Check the length with:
//     len(mockedInitialiser.DoGetKafkaProducerCalls())
func (mock *InitialiserMock) DoGetKafkaProducerCalls() []struct {
	Ctx context.Context
	Cfg *config.Config
} {
	var calls []struct {
		Ctx context.Context
		Cfg *config.Config
	}
	mock.lockDoGetKafkaProducer.RLock()
	calls = mock.calls.DoGetKafkaProducer
	mock.lockDoGetKafkaProducer.RUnlock()
	return calls
}

// DoGetMongoDB calls DoGetMongoDBFunc.
func (mock *InitialiserMock) DoGetMongoDB(ctx context.Context, cfg *config.Config) (api.MongoServer, error) {
	if mock.DoGetMongoDBFunc == nil {
		panic("InitialiserMock.DoGetMongoDBFunc: method is nil but Initialiser.DoGetMongoDB was just called")
	}
	callInfo := struct {
		Ctx context.Context
		Cfg *config.Config
	}{
		Ctx: ctx,
		Cfg: cfg,
	}
	mock.lockDoGetMongoDB.Lock()
	mock.calls.DoGetMongoDB = append(mock.calls.DoGetMongoDB, callInfo)
	mock.lockDoGetMongoDB.Unlock()
	return mock.DoGetMongoDBFunc(ctx, cfg)
}

// DoGetMongoDBCalls gets all the calls that were made to DoGetMongoDB.
// Check the length with:
//     len(mockedInitialiser.DoGetMongoDBCalls())
func (mock *InitialiserMock) DoGetMongoDBCalls() []struct {
	Ctx context.Context
	Cfg *config.Config
} {
	var calls []struct {
		Ctx context.Context
		Cfg *config.Config
	}
	mock.lockDoGetMongoDB.RLock()
	calls = mock.calls.DoGetMongoDB
	mock.lockDoGetMongoDB.RUnlock()
	return calls
}

// DoGetS3Client calls DoGetS3ClientFunc.
func (mock *InitialiserMock) DoGetS3Client(ctx context.Context, cfg *config.Config) (upload.S3Interface, error) {
	if mock.DoGetS3ClientFunc == nil {
		panic("InitialiserMock.DoGetS3ClientFunc: method is nil but Initialiser.DoGetS3Client was just called")
	}
	callInfo := struct {
		Ctx context.Context
		Cfg *config.Config
	}{
		Ctx: ctx,
		Cfg: cfg,
	}
	mock.lockDoGetS3Client.Lock()
	mock.calls.DoGetS3Client = append(mock.calls.DoGetS3Client, callInfo)
	mock.lockDoGetS3Client.Unlock()
	return mock.DoGetS3ClientFunc(ctx, cfg)
}

// DoGetS3ClientCalls gets all the calls that were made to DoGetS3Client.
// Check the length with:
//     len(mockedInitialiser.DoGetS3ClientCalls())
func (mock *InitialiserMock) DoGetS3ClientCalls() []struct {
	Ctx context.Context
	Cfg *config.Config
} {
	var calls []struct {
		Ctx context.Context
		Cfg *config.Config
	}
	mock.lockDoGetS3Client.RLock()
	calls = mock.calls.DoGetS3Client
	mock.lockDoGetS3Client.RUnlock()
	return calls
}
