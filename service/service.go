package service

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/health"
	dpauth "github.com/ONSdigital/dp-authorisation/auth"
	"github.com/ONSdigital/dp-interactives-api/api"
	"github.com/ONSdigital/dp-interactives-api/config"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-net/handlers"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
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
	interactivesKafkaConsumer kafka.IConsumerGroup
}

func Run(ctx context.Context, cfg *config.Config, serviceList *ExternalServiceList, buildTime, gitCommit, version string, svcErrors chan error) (*Service, error) {
	log.Info(ctx, "running service")
	var zc *health.Client
	var auth api.AuthHandler

	// Get Health client for Zebedee and permissions
	zc = serviceList.GetHealthClient("Zebedee", cfg.ZebedeeURL)
	auth = getAuthorisationHandlers(zc)

	// Get HTTP Server with collectionID checkHeader middleware
	r := mux.NewRouter()
	middleware := alice.New(handlers.CheckHeader(handlers.CollectionID))
	s := serviceList.GetHTTPServer(cfg.BindAddr, middleware.Then(r))

	mongoDB, err := serviceList.GetMongoDB(ctx, cfg)
	if err != nil {
		log.Fatal(ctx, "failed to initialise mongo DB", err)
		return nil, err
	}

	// Get Kafka consumer
	consumer, err := serviceList.GetKafkaConsumer(ctx, cfg)
	if err != nil {
		log.Fatal(ctx, "failed to initialise kafka consumer", err)
		return nil, err
	}
	producer, err := serviceList.GetKafkaProducer(ctx, cfg)
	if err != nil {
		log.Fatal(ctx, "failed to initialise kafka producer", err)
		return nil, err
	}

	// Get S3Uploaded client
	s3Client, err := serviceList.GetS3Client(ctx, cfg)
	if err != nil {
		log.Fatal(ctx, "failed to initialise S3 client for uploaded bucket", err)
		return nil, err
	}

	a := api.Setup(ctx, cfg, r, auth, mongoDB, producer, consumer, s3Client)

	//heathcheck - start
	hc, err := serviceList.GetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		log.Fatal(ctx, "could not instantiate healthcheck", err)
		return nil, err
	}
	if err := registerCheckers(ctx, cfg, hc, mongoDB, producer, consumer); err != nil {
		return nil, errors.Wrap(err, "unable to register checkers")
	}

	r.StrictSlash(true).Path("/health").HandlerFunc(hc.Handler)
	hc.Start(ctx)
	//healthcheck - end

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
		healthCheck:               nil,
		mongoDB:                   mongoDB,
		interactivesKafkaProducer: producer,
		interactivesKafkaConsumer: consumer,
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

		if svc.serviceList.KafkaConsumer {
			if err := svc.interactivesKafkaConsumer.Close(ctx); err != nil {
				log.Error(ctx, "error closing Kafka consumer", err)
				hasShutdownError = true
			}
		}

		// stop any incoming requests before closing any outbound connections
		if err := svc.server.Shutdown(ctx); err != nil {
			log.Error(ctx, "failed to shutdown http server", err)
			hasShutdownError = true
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
	consumer kafka.IConsumerGroup) (err error) {

	hasErrors := false

	if err = hc.AddCheck("Mongo DB", mongoDB.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for mongo db", err)
	}

	if err = hc.AddCheck("Uploaded Kafka Producer", producer.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for uploaded kafka producer", err, log.Data{"topic": cfg.InteractivesTopic})
	}

	if err = hc.AddCheck("Published Kafka Consumer", consumer.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for published kafka consumer", err, log.Data{"group": cfg.InteractivesGroup, "topic": cfg.InteractivesTopic})
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}

// generate permissions from dp-auth-api, using the provided health client, reusing its http Client
func getAuthorisationHandlers(zc *health.Client) api.AuthHandler {
	log.Info(context.Background(), "getting Authorisation Handlers", log.Data{"zc_url": zc.URL})

	authClient := dpauth.NewPermissionsClient(zc.Client)
	authVerifier := dpauth.DefaultPermissionsVerifier()

	// for checking caller permissions when we only have a user/service token
	permissions := dpauth.NewHandler(
		dpauth.NewPermissionsRequestBuilder(zc.URL),
		authClient,
		authVerifier,
	)

	return permissions
}