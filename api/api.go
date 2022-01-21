package api

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-interactives-api/config"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

type API struct {
	Router  *mux.Router
	mongoDB MongoServer
	//auth               AuthHandler

	//uploadProducer     *event.AvroProducer
	//publishedProducer  *event.AvroProducer
	//urlBuilder         *url.Builder
	//downloadServiceURL string
}

// Setup creates the API struct and its endpoints with corresponding handlers
func Setup(ctx context.Context, cfg *config.Config, r *mux.Router, auth AuthHandler, mongoDB MongoServer, kafkaProducer kafka.IProducer, kafkaConsumer kafka.IConsumerGroup) *API {

	api := &API{
		Router:  r,
		mongoDB: mongoDB,
	}

	//r.HandleFunc("/interactives", auth.Require(dpauth.Permissions{Read: true}, api.UploadVisualisationHandler)).Methods(http.MethodPut)
	r.HandleFunc("/interactives", api.UploadVisualisationHandler).Methods(http.MethodPut)

	return api
}

// Close is called during graceful shutdown to give the API an opportunity to perform any required disposal task
func (*API) Close(ctx context.Context) error {
	log.Info(ctx, "graceful shutdown of api complete")
	return nil
}
