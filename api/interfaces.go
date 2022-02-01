package api

import (
	"context"
	"net/http"

	dpauth "github.com/ONSdigital/dp-authorisation/auth"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-interactives-api/models"
)

//go:generate moq -out mock/mongo.go -pkg mock . MongoServer

type MongoServer interface {
	Close(ctx context.Context) error
	Checker(ctx context.Context, state *healthcheck.CheckState) (err error)
	UpsertVisualisation(ctx context.Context, id string, vis *models.Visualisation) (err error)
	GetVisualisationFromSHA(ctx context.Context, sha string) (*models.Visualisation, error)
}

// AuthHandler interface for adding auth to endpoints
type AuthHandler interface {
	Require(required dpauth.Permissions, handler http.HandlerFunc) http.HandlerFunc
}
