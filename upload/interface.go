package upload

import (
	"context"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out mock/s3.go -pkg mock . S3Interface

type S3Interface interface {
	Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
	ValidateBucket() error
	Checker(ctx context.Context, state *health.CheckState) error
}
