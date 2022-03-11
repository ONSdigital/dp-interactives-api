package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-interactives-api/api"
	"github.com/ONSdigital/dp-interactives-api/config"
	"github.com/ONSdigital/dp-interactives-api/upload"
	kafka "github.com/ONSdigital/dp-kafka/v3"
)

//go:generate moq -out mock/initialiser.go -pkg mock . Initialiser
//go:generate moq -out mock/server.go -pkg mock . HTTPServer
//go:generate moq -out mock/healthcheck.go -pkg mock . HealthChecker

// Initialiser defines the methods to initialise external services
type Initialiser interface {
	DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer
	DoGetMongoDB(ctx context.Context, cfg *config.Config) (api.MongoServer, error)
	DoGetKafkaProducer(ctx context.Context, cfg *config.Config) (kafka.IProducer, error)
	DoGetHealthClient(name, url string) *health.Client
	DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error)
	DoGetS3Client(ctx context.Context, cfg *config.Config) (upload.S3Interface, error)
	DoGetAuthorisationMiddleware(ctx context.Context, authorisationConfig *authorisation.Config) (authorisation.Middleware, error)
}

// HTTPServer defines the required methods from the HTTP server
type HTTPServer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

// HealthChecker defines the required methods from Healthcheck
type HealthChecker interface {
	Handler(w http.ResponseWriter, req *http.Request)
	Start(ctx context.Context)
	Stop()
	AddCheck(name string, checker healthcheck.Checker) (err error)
}
