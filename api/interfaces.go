package api

import (
	"context"
	"github.com/ONSdigital/dp-interactives-api/mongo"
	"net/http"

	dpauth "github.com/ONSdigital/dp-authorisation/auth"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-interactives-api/models"
)

//go:generate moq -out mock/mongo.go -pkg mock . MongoServer
//go:generate moq -out mock/auth.go -pkg mock . AuthHandler
//go:generate moq -out mock/filesservice.go -pkg mock . FilesService

type MongoServer interface {
	Close(ctx context.Context) error
	Checker(ctx context.Context, state *healthcheck.CheckState) (err error)
	UpsertInteractive(ctx context.Context, id string, vis *models.Interactive) (err error)
	GetActiveInteractiveGivenSha(ctx context.Context, sha string) (*models.Interactive, error)
	GetActiveInteractiveGivenField(ctx context.Context, fieldName, fieldValue string) (*models.Interactive, error)
	GetInteractive(ctx context.Context, id string) (*models.Interactive, error)
	ListInteractives(ctx context.Context, offset, limit int, filter *models.InteractiveFilter) ([]*models.Interactive, int, error)
	PatchInteractive(context.Context, mongo.PatchAction, *models.Interactive) error
}

// AuthHandler interface for adding auth to endpoints
type AuthHandler interface {
	Require(required dpauth.Permissions, handler http.HandlerFunc) http.HandlerFunc
}

type FilesService interface {
	SetCollectionID(ctx context.Context, file, collectionID string) error
	PublishCollection(ctx context.Context, collectionID string) error
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}
