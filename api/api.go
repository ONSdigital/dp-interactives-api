package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ONSdigital/dp-interactives-api/models"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-interactives-api/config"
	"github.com/ONSdigital/dp-interactives-api/event"
	"github.com/ONSdigital/dp-interactives-api/pagination"
	"github.com/ONSdigital/dp-interactives-api/schema"
	"github.com/ONSdigital/dp-interactives-api/upload"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
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
	s3            upload.S3Interface
	newUUID       models.Generator
	newResourceID models.Generator
	newSlug       models.Generator
}

// Setup creates the API struct and its endpoints with corresponding handlers
func Setup(ctx context.Context,
	cfg *config.Config,
	r *mux.Router,
	auth authorisation.Middleware,
	mongoDB MongoServer,
	kafkaProducer kafka.IProducer,
	s3 upload.S3Interface,
	filesService FilesService,
	newUUID models.Generator,
	newResourceID models.Generator,
	newSlug models.Generator) *API {

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
	}

	paginator := pagination.NewPaginator(cfg.DefaultLimit, cfg.DefaultOffset, cfg.DefaultMaxLimit)

	if r != nil {
		if cfg.PublishingEnabled {
			r.HandleFunc("/v1/interactives", auth.Require(InteractivesCreatePermission, api.UploadInteractivesHandler)).Methods(http.MethodPost)
			r.HandleFunc("/v1/interactives", auth.Require(InteractivesReadPermission, paginator.Paginate(api.ListInteractivesHandler))).Methods(http.MethodGet)
			r.HandleFunc("/v1/interactives/{id}", auth.Require(InteractivesReadPermission, api.GetInteractiveMetadataHandler)).Methods(http.MethodGet)
			r.HandleFunc("/v1/interactives/{id}", auth.Require(InteractivesUpdatePermission, api.UpdateInteractiveHandler)).Methods(http.MethodPut)
			r.HandleFunc("/v1/interactives/{id}", auth.Require(InteractivesDeletePermission, api.DeleteInteractivesHandler)).Methods(http.MethodDelete)
		} else {
			r.HandleFunc("/v1/interactives", paginator.Paginate(api.ListInteractivesHandler)).Methods(http.MethodGet)
			r.HandleFunc("/v1/interactives/{id}", api.GetInteractiveMetadataHandler).Methods(http.MethodGet)
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

func (api *API) uploadFile(sha, filename string, data []byte) (string, error) {
	err := api.s3.ValidateBucket()
	if err != nil {
		return "", fmt.Errorf("invalid s3 bucket %w", err)
	}

	fileWithPath := fmt.Sprintf("%s/%s", sha, filename)
	_, err = api.s3.Upload(&s3manager.UploadInput{Body: bytes.NewReader(data), Key: &fileWithPath})
	if err != nil {
		return "", fmt.Errorf("s3 upload error %w", err)
	}

	return fileWithPath, nil
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

func WriteJSONBody(v interface{}, w http.ResponseWriter, httpStatus int) error {

	// Set headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	// Marshal provided model
	payload, err := json.Marshal(v)
	if err != nil {
		return err
	}

	// Write payload to body
	if _, err := w.Write(payload); err != nil {
		return err
	}
	return nil
}
