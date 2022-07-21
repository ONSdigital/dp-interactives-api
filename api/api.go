package api

import (
	"context"
	"fmt"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-interactives-api/config"
	"github.com/ONSdigital/dp-interactives-api/event"
	"github.com/ONSdigital/dp-interactives-api/internal/data"
	"github.com/ONSdigital/dp-interactives-api/models"
	"github.com/ONSdigital/dp-interactives-api/schema"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-net/v2/responder"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gorilla/mux"
	"net/http"
	"os"
)

const (
	InteractivesCreatePermission string = "interactives:create"
	InteractivesReadPermission   string = "interactives:read"
	InteractivesUpdatePermission string = "interactives:update"
	InteractivesDeletePermission string = "interactives:delete"
)

type API struct {
	cfg           *config.Config
	Router        *mux.Router
	mongoDB       MongoServer
	filesService  FilesService
	auth          authorisation.Middleware
	producer      *event.AvroProducer
	s3            S3Interface
	newUUID       data.Generator
	newResourceID data.Generator
	newSlug       data.Generator
	respond       *responder.Responder
}

// Setup creates the API struct and its endpoints with corresponding handlers
func Setup(ctx context.Context,
	cfg *config.Config,
	r *mux.Router,
	auth authorisation.Middleware,
	mongoDB MongoServer,
	kafkaProducer kafka.IProducer,
	s3 S3Interface,
	filesService FilesService,
	newUUID data.Generator,
	newResourceID data.Generator,
	newSlug data.Generator,
	respond *responder.Responder) *API {

	var kProducer *event.AvroProducer
	if kafkaProducer != nil {
		kProducer = event.NewAvroProducer(kafkaProducer.Channels().Output, schema.InteractiveUploadedEvent)
	} else {
		log.Error(ctx, "api setup error - no kafka producer", nil)
	}

	api := &API{
		cfg:           cfg,
		Router:        r,
		mongoDB:       mongoDB,
		auth:          auth,
		s3:            s3,
		filesService:  filesService,
		producer:      kProducer,
		newUUID:       newUUID,
		newSlug:       newSlug,
		newResourceID: newResourceID,
		respond:       respond,
	}

	if r != nil {
		if cfg.PublishingEnabled {
			r.HandleFunc("/v1/interactives", auth.Require(InteractivesCreatePermission, api.UploadInteractivesHandler)).Methods(http.MethodPost)
			r.HandleFunc("/v1/interactives", auth.Require(InteractivesReadPermission, api.ListInteractivesHandler)).Methods(http.MethodGet)
			r.HandleFunc("/v1/interactives/{id}", auth.Require(InteractivesReadPermission, api.GetInteractiveHandler)).Methods(http.MethodGet)
			r.HandleFunc("/v1/interactives/{id}", auth.Require(InteractivesUpdatePermission, api.UpdateInteractiveHandler)).Methods(http.MethodPut)
			r.HandleFunc("/v1/interactives/{id}", auth.Require(InteractivesUpdatePermission, api.PatchInteractiveHandler)).Methods(http.MethodPatch)
			r.HandleFunc("/v1/interactives/{id}", auth.Require(InteractivesDeletePermission, api.DeleteInteractivesHandler)).Methods(http.MethodDelete)
			r.HandleFunc("/v1/collection/{id}", auth.Require(InteractivesUpdatePermission, api.PublishCollectionHandler)).Methods(http.MethodPatch)
		} else {
			r.HandleFunc("/v1/interactives", api.ListInteractivesHandler).Methods(http.MethodGet)
			r.HandleFunc("/v1/interactives/{id}", api.GetInteractiveHandler).Methods(http.MethodGet)
		}
	} else {
		log.Error(ctx, "api setup error - no router", nil)
	}

	return api
}

// Close is called during graceful shutdown to give the API an opportunity to perform any required disposal task
func (*API) Close(ctx context.Context) error {
	log.Info(ctx, "graceful shutdown of api complete")
	return nil
}

func (api *API) uploadFile(req *FormDataRequest) (string, error) {
	err := api.s3.ValidateBucket()
	if err != nil {
		return "", fmt.Errorf("invalid s3 bucket %w", err)
	}

	localFile, err := os.Open(req.TmpFileName)
	if err != nil {
		return "", fmt.Errorf("cannot open zipfile %w", err)
	}

	uniqueS3Key := fmt.Sprintf("%s/%s", api.newUUID(""), req.Name)
	_, err = api.s3.Upload(&s3manager.UploadInput{Body: localFile, Key: &uniqueS3Key})
	if err != nil {
		return "", fmt.Errorf("s3 upload error %w", err)
	}

	return uniqueS3Key, nil
}

func (api *API) blockAccess(i *models.Interactive) bool {
	if i == nil {
		return true
	}

	if i.Active == nil || !*i.Active {
		//block all access to deleted interactives
		return true
	}

	//all in publishing mode or only published interactives in web
	viewable := api.cfg.PublishingEnabled || *i.Published

	return !viewable
}
