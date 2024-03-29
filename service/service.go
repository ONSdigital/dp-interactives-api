package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-interactives-api/api"
	"github.com/ONSdigital/dp-interactives-api/config"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Service contains all the configs, server and clients to run the interactives API
type Service struct {
	config                    *config.Config
	server                    HTTPServer
	router                    *mux.Router
	api                       *api.API
	serviceList               *ExternalServiceList
	healthCheck               HealthChecker
	mongoDB                   api.MongoServer
	interactivesKafkaProducer kafka.IProducer
	authorisationMiddleware   authorisation.Middleware
	filesService              api.FilesService
}

func Run(ctx context.Context, cfg *config.Config, serviceList *ExternalServiceList, buildTime, gitCommit, version string, svcErrors chan error) (*Service, error) {
	log.Info(ctx, "running service")

	r := mux.NewRouter()
	s := serviceList.GetHTTPServer(cfg.BindAddr, r)

	mongoDB, err := serviceList.GetMongoDB(ctx, cfg)
	if err != nil {
		log.Fatal(ctx, "failed to initialise mongo DB", err)
		return nil, err
	}

	var s3Client api.S3Interface
	var producer kafka.IProducer
	var filesService api.FilesService
	var authorisationMiddleware authorisation.Middleware
	if cfg.PublishingEnabled {
		// Get S3Uploaded client
		s3Client, err = serviceList.GetS3Client(ctx, cfg)
		if err != nil {
			log.Fatal(ctx, "failed to initialise S3 client for uploaded bucket", err)
			return nil, err
		}

		// Get Kafka producer
		producer, err = serviceList.GetKafkaProducer(ctx, cfg)
		if err != nil {
			log.Fatal(ctx, "failed to initialise kafka producer", err)
			return nil, err
		}

		filesService, err = serviceList.GetFilesService(ctx, cfg)
		if err != nil {
			log.Fatal(ctx, "failed to initialise files service", err)
			return nil, err
		}

		// Auth - only needed in publish
		authorisationMiddleware, err = serviceList.GetAuthorisationMiddleware(ctx, cfg.AuthorisationConfig)
		if err != nil {
			log.Fatal(ctx, "could not instantiate authorisation middleware", err)
			return nil, err
		}
	}

	uuidGen, resourceIdGen, slugGen := serviceList.GetGenerators()
	responder, _ := serviceList.GetResponder(ctx, cfg)
	a := api.Setup(ctx, cfg, r, authorisationMiddleware, mongoDB, producer, s3Client, filesService, uuidGen, resourceIdGen, slugGen, responder)

	//heathcheck
	hc, err := serviceList.GetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		log.Fatal(ctx, "could not instantiate healthcheck", err)
		return nil, err
	}
	err = registerCheckers(ctx, cfg, hc, mongoDB, producer, s3Client, authorisationMiddleware, filesService)
	if err != nil {
		return nil, errors.Wrap(err, "unable to register checkers")
	}

	r.StrictSlash(true).Path("/health").Methods(http.MethodGet).HandlerFunc(hc.Handler)
	hc.Start(ctx)

	// Run the http server in a new go-routine
	go func() {
		if err := s.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()

	return &Service{
		config:                    cfg,
		server:                    s,
		router:                    r,
		api:                       a,
		serviceList:               serviceList,
		healthCheck:               hc,
		mongoDB:                   mongoDB,
		interactivesKafkaProducer: producer,
		authorisationMiddleware:   authorisationMiddleware,
		filesService:              filesService,
	}, nil
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.config.GracefulShutdownTimeout
	log.Info(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout})
	ctx, cancel := context.WithTimeout(ctx, timeout)

	// track shutown gracefully closes up
	var gracefulShutdown bool

	go func() {
		defer cancel()
		var hasShutdownError bool

		// stop healthcheck first, as it depends on everything else
		if svc.serviceList.HealthCheck {
			svc.healthCheck.Stop()
		}

		// stop any incoming requests before closing any outbound connections
		if err := svc.server.Shutdown(ctx); err != nil {
			log.Error(ctx, "failed to shutdown http server", err)
			hasShutdownError = true
		}

		// close API
		if err := svc.api.Close(ctx); err != nil {
			log.Error(ctx, "error closing API", err)
			hasShutdownError = true
		}

		// close mongoDB
		if svc.serviceList.MongoDB {
			if err := svc.mongoDB.Close(ctx); err != nil {
				log.Error(ctx, "error closing mongoDB", err)
				hasShutdownError = true
			}
		}

		if svc.serviceList.KafkaProducer {
			if err := svc.interactivesKafkaProducer.Close(ctx); err != nil {
				log.Error(ctx, "error closing Kafka producer", err)
				hasShutdownError = true
			}
		}

		if svc.config.PublishingEnabled {
			if err := svc.authorisationMiddleware.Close(ctx); err != nil {
				log.Error(ctx, "failed to close authorisation middleware", err)
				hasShutdownError = true
			}
		}

		if !hasShutdownError {
			gracefulShutdown = true
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	if !gracefulShutdown {
		err := errors.New("failed to shutdown gracefully")
		log.Error(ctx, "failed to shutdown gracefully ", err)
		return err
	}

	log.Info(ctx, "graceful shutdown was successful")
	return nil
}

func registerCheckers(ctx context.Context,
	cfg *config.Config,
	hc HealthChecker,
	mongoDB api.MongoServer,
	producer kafka.IProducer,
	s3 api.S3Interface,
	authorisationMiddleware authorisation.Middleware,
	filesService api.FilesService) (err error) {

	hasErrors := false

	if err = hc.AddCheck("Mongo DB", mongoDB.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for mongo db", err)
	}

	if cfg.PublishingEnabled {
		if err = hc.AddCheck("Uploaded Kafka Producer", producer.Checker); err != nil {
			hasErrors = true
			log.Error(ctx, "error adding check for uploaded kafka producer", err, log.Data{"topic": cfg.InteractivesWriteTopic})
		}

		if err = hc.AddCheck("S3 checker", s3.Checker); err != nil {
			hasErrors = true
			log.Error(ctx, "error adding check for s3", err)
		}

		if err = hc.AddCheck("FilesService checker", filesService.Checker); err != nil {
			hasErrors = true
			log.Error(ctx, "error adding check for filesService", err)
		}

		if err := hc.AddCheck("permissions cache health check", authorisationMiddleware.HealthCheck); err != nil {
			hasErrors = true
			log.Error(ctx, "error adding check for permissions cache", err)
		}
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}
