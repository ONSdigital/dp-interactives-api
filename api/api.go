package api

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-interactives-api/config"
	"github.com/ONSdigital/dp-interactives-api/upload"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

type API struct {
	Router   *mux.Router
	mongoDB  MongoServer
	auth     AuthHandler
	producer kafka.IProducer
	consumer kafka.IConsumerGroup
	s3       upload.S3Interface
}

// Setup creates the API struct and its endpoints with corresponding handlers
func Setup(ctx context.Context, cfg *config.Config, r *mux.Router, auth AuthHandler, mongoDB MongoServer, kafkaProducer kafka.IProducer, kafkaConsumer kafka.IConsumerGroup, s3 upload.S3Interface) *API {

	api := &API{
		Router:   r,
		mongoDB:  mongoDB,
		auth:     auth,
		producer: kafkaProducer,
		consumer: kafkaConsumer,
		s3:       s3,
	}

	//r.HandleFunc("/interactives", auth.Require(dpauth.Permissions{Read: true}, api.UploadVisualisationHandler)).Methods(http.MethodPut)
	r.HandleFunc("/interactives", api.UploadVisualisationHandler).Methods(http.MethodPut)
	r.HandleFunc("/interactives/{id}", api.GetVisualisationInfoHandler).Methods(http.MethodGet)
	r.HandleFunc("/interactives/{id}", api.DeleteVisualisationHandler).Methods(http.MethodDelete)

	return api
}

// Close is called during graceful shutdown to give the API an opportunity to perform any required disposal task
func (*API) Close(ctx context.Context) error {
	log.Info(ctx, "graceful shutdown of api complete")
	return nil
}
