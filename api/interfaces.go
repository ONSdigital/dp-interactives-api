package api

import (
	"context"
	"net/http"

	dpauth "github.com/ONSdigital/dp-authorisation/auth"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-interactives-api/models"
)

//go:generate moq -out mock/mongo.go -pkg mock . MongoServer
//go:generate moq -out mock/auth.go -pkg mock . AuthHandler

type MongoServer interface {
	Close(ctx context.Context) error
	Checker(ctx context.Context, state *healthcheck.CheckState) (err error)
	UpsertInteractive(ctx context.Context, id string, vis *models.Interactive) (err error)
	GetInteractiveFromSHA(ctx context.Context, sha string) (*models.Interactive, error)
	GetInteractive(ctx context.Context, id string) (*models.Interactive, error)
	ListInteractives(ctx context.Context, offset, limit int) (interface{}, int, error)
}

// AuthHandler interface for adding auth to endpoints
type AuthHandler interface {
	Require(required dpauth.Permissions, handler http.HandlerFunc) http.HandlerFunc
}
