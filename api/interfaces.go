package api

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/v2/interactives"
	dpauth "github.com/ONSdigital/dp-authorisation/auth"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-interactives-api/models"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"net/http"
)

//go:generate moq -out mock/mongo.go -pkg mock . MongoServer
//go:generate moq -out mock/auth.go -pkg mock . AuthHandler
//go:generate moq -out mock/filesservice.go -pkg mock . FilesService
//go:generate moq -out mock/s3.go -pkg mock . S3Interface

type MongoServer interface {
	Close(ctx context.Context) error
	Checker(ctx context.Context, state *healthcheck.CheckState) (err error)
	UpsertInteractive(ctx context.Context, id string, vis *models.Interactive) (err error)
	GetInteractive(ctx context.Context, id string) (*models.Interactive, error)
	ListInteractives(ctx context.Context, filter *models.Filter) ([]*models.Interactive, error)
	PatchInteractive(context.Context, interactives.PatchAttribute, *models.Interactive) error
}

// AuthHandler interface for adding auth to endpoints
type AuthHandler interface {
	Require(required dpauth.Permissions, handler http.HandlerFunc) http.HandlerFunc
}

type FilesService interface {
	SetCollectionID(ctx context.Context, file, collectionID string) error
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}

type S3Interface interface {
	Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
	ValidateBucket() error
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}
