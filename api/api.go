package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ONSdigital/dp-interactives-api/config"
	"github.com/ONSdigital/dp-interactives-api/event"
	"github.com/ONSdigital/dp-interactives-api/schema"
	"github.com/ONSdigital/dp-interactives-api/upload"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

var (
	ErrNoBody = errors.New("no body in http request")
)

type API struct {
	Router   *mux.Router
	mongoDB  MongoServer
	auth     AuthHandler
	producer *event.AvroProducer
	s3       upload.S3Interface
}

// Setup creates the API struct and its endpoints with corresponding handlers
func Setup(ctx context.Context, cfg *config.Config, r *mux.Router, auth AuthHandler, mongoDB MongoServer, kafkaProducer kafka.IProducer, s3 upload.S3Interface) *API {

	/*r.HandleFunc("/interactives", auth.Require(dpauth.Permissions{Create: true}, api.UploadInteractivesHandler)).Methods(http.MethodPost)
	r.HandleFunc("/interactives/{id}", auth.Require(dpauth.Permissions{Read: true}, api.GetInteractiveInfoHandler)).Methods(http.MethodGet)
	r.HandleFunc("/interactives/{id}", auth.Require(dpauth.Permissions{Update: true}, api.UpdateInteractiveInfoHandler)).Methods(http.MethodPost)
	r.HandleFunc("/interactives", auth.Require(dpauth.Permissions{Read: true}, api.ListInteractivessHandler)).Methods(http.MethodGet)*/
	var kProducer *event.AvroProducer
	if kafkaProducer != nil {
		kProducer = event.NewAvroProducer(kafkaProducer.Channels().Output, schema.InteractiveUploadedEvent)
	} else {
		log.Error(ctx, "api setup error - no kafka producer", nil)
	}

	api := &API{
		Router:   r,
		mongoDB:  mongoDB,
		auth:     auth,
		s3:       s3,
		producer: kProducer,
	}

	if r != nil {
		r.HandleFunc("/interactives", api.UploadInteractivesHandler).Methods(http.MethodPost)
		r.HandleFunc("/interactives/{id}", api.GetInteractiveMetadataHandler).Methods(http.MethodGet)
		r.HandleFunc("/interactives/{id}", api.UpdateInteractiveInfoHandler).Methods(http.MethodPost)
		r.HandleFunc("/interactives", api.ListInteractivessHandler).Methods(http.MethodGet)
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

func WriteJSONBody(v interface{}, w http.ResponseWriter, httpStatus int) error {

	// Set headers
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
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
