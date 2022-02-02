package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/ONSdigital/dp-interactives-api/config"
	"github.com/ONSdigital/dp-interactives-api/event"
	"github.com/ONSdigital/dp-interactives-api/schema"
	"github.com/ONSdigital/dp-interactives-api/upload"
	kafka "github.com/ONSdigital/dp-kafka/v2"
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

	api := &API{
		Router:  r,
		mongoDB: mongoDB,
		auth:    auth,
		s3:      s3,
	}

	/*r.HandleFunc("/interactives", auth.Require(dpauth.Permissions{Create: true}, api.UploadVisualisationHandler)).Methods(http.MethodPost)
	r.HandleFunc("/interactives/{id}", auth.Require(dpauth.Permissions{Read: true}, api.GetVisualisationInfoHandler)).Methods(http.MethodGet)
	r.HandleFunc("/interactives/{id}", auth.Require(dpauth.Permissions{Update: true}, api.UpdateVisualisationInfoHandler)).Methods(http.MethodPost)
	r.HandleFunc("/interactives", auth.Require(dpauth.Permissions{Read: true}, api.GetAllVisualisationsHandler)).Methods(http.MethodGet)*/
	api.producer = event.NewAvroProducer(kafkaProducer.Channels().Output, schema.InteractiveUploadedEvent)
	r.HandleFunc("/interactives", api.UploadVisualisationHandler).Methods(http.MethodPost)
	r.HandleFunc("/interactives/{id}", api.GetVisualisationInfoHandler).Methods(http.MethodGet)
	r.HandleFunc("/interactives/{id}", api.UpdateVisualisationInfoHandler).Methods(http.MethodPost)
	r.HandleFunc("/interactives", api.ListVisualisationsHandler).Methods(http.MethodGet)

	return api
}

// Close is called during graceful shutdown to give the API an opportunity to perform any required disposal task
func (*API) Close(ctx context.Context) error {
	log.Info(ctx, "graceful shutdown of api complete")
	return nil
}
