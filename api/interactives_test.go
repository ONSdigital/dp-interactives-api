package api_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ONSdigital/dp-interactives-api/api"
	mongoMock "github.com/ONSdigital/dp-interactives-api/api/mock"
	"github.com/ONSdigital/dp-interactives-api/config"
	"github.com/ONSdigital/dp-interactives-api/models"
	"github.com/ONSdigital/dp-interactives-api/upload"
	s3Mock "github.com/ONSdigital/dp-interactives-api/upload/mock"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	kMock "github.com/ONSdigital/dp-kafka/v3/kafkatest"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestUploadInteractivesHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		title         string
		req           *http.Request
		responseCode  int
		formFile      string
		mongoServer   api.MongoServer
		s3            upload.S3Interface
		kafkaProducer kafka.IProducer
	}{
		{
			title:        "WhenMissingAttachment_ThenStatusBadRequest",
			req:          httptest.NewRequest(http.MethodPost, "/interactives", nil),
			responseCode: http.StatusBadRequest,
		},
		{
			title:        "WhenUploadedFileIsNotZip_ThenStatusBadRequest",
			formFile:     "./mock/fortest.txt",
			responseCode: http.StatusBadRequest,
		},
		{
			title:        "WhenFileIsZipButAlreadyExists_ThenStatusBadRequest",
			formFile:     "./mock/interactives.zip",
			responseCode: http.StatusBadRequest,
			mongoServer: &mongoMock.MongoServerMock{
				GetInteractiveFromSHAFunc: func(ctx context.Context, sha string) (*models.Interactive, error) { return &models.Interactive{}, nil },
			},
		},
		{
			title:        "WhenValidationPassButS3BucketNotExisting_ThenInternalServerError",
			formFile:     "./mock/interactives.zip",
			responseCode: http.StatusInternalServerError,
			mongoServer: &mongoMock.MongoServerMock{
				GetInteractiveFromSHAFunc: func(ctx context.Context, sha string) (*models.Interactive, error) { return nil, nil },
			},
			s3: &s3Mock.S3InterfaceMock{
				ValidateBucketFunc: func() error { return errors.New("s3 error") },
			},
		},
		{
			title:        "WhenUploadError_ThenInternalServerError",
			formFile:     "./mock/interactives.zip",
			responseCode: http.StatusInternalServerError,
			mongoServer: &mongoMock.MongoServerMock{
				GetInteractiveFromSHAFunc: func(ctx context.Context, sha string) (*models.Interactive, error) { return nil, nil },
			},
			s3: &s3Mock.S3InterfaceMock{
				ValidateBucketFunc: func() error { return nil },
				UploadFunc: func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
					return nil, errors.New("upload error")
				},
			},
		},
		{
			title:        "WhenDbError_ThenInternalServerError",
			formFile:     "./mock/interactives.zip",
			responseCode: http.StatusInternalServerError,
			mongoServer: &mongoMock.MongoServerMock{
				GetInteractiveFromSHAFunc: func(ctx context.Context, sha string) (*models.Interactive, error) { return nil, nil },
				UpsertInteractiveFunc: func(ctx context.Context, id string, vis *models.Interactive) error {
					return errors.New("db upsert error")
				},
			},
			s3: &s3Mock.S3InterfaceMock{
				ValidateBucketFunc: func() error { return nil },
				UploadFunc: func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
					return nil, nil
				},
			},
		},
		{
			title:        "WhenAllSuccess_ThenStatusAccepted",
			formFile:     "./mock/interactives.zip",
			responseCode: http.StatusAccepted,
			mongoServer: &mongoMock.MongoServerMock{
				GetInteractiveFromSHAFunc: func(ctx context.Context, sha string) (*models.Interactive, error) { return nil, nil },
				UpsertInteractiveFunc: func(ctx context.Context, id string, vis *models.Interactive) error {
					return nil
				},
			},
			s3: &s3Mock.S3InterfaceMock{
				ValidateBucketFunc: func() error { return nil },
				UploadFunc: func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
					return nil, nil
				},
			},
			kafkaProducer: &kMock.IProducerMock{
				ChannelsFunc: func() *kafka.ProducerChannels { return &kafka.ProducerChannels{Output: nil} },
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.title, func(t *testing.T) {
			ctx := context.Background()
			api := api.Setup(ctx, &config.Config{}, nil, nil, tc.mongoServer, tc.kafkaProducer, tc.s3)
			resp := httptest.NewRecorder()
			if tc.formFile != "" {
				tc.req, _ = newfileUploadRequest("/interactives", map[string]string{"metadata1": "value1"}, "attachment", tc.formFile)
			}

			api.UploadInteractivesHandler(resp, tc.req)

			require.Equal(t, tc.responseCode, resp.Result().StatusCode)
		})
	}
}

func TestGetInteractiveMetadataHandler(t *testing.T) {
	t.Parallel()
	interactiveID := "11-22-33-44"
	tests := []struct {
		title        string
		responseCode int
		mongoServer  api.MongoServer
	}{
		{
			title:        "WhenMissingInDatabase_ThenStatusNotFound",
			responseCode: http.StatusNotFound,
			mongoServer: &mongoMock.MongoServerMock{
				GetInteractiveFunc: func(ctx context.Context, id string) (*models.Interactive, error) { return nil, nil },
			},
		},
		{
			title:        "WhenInteractiveIsDeleted_ThenStatusNotFound",
			responseCode: http.StatusNotFound,
			mongoServer: &mongoMock.MongoServerMock{
				GetInteractiveFunc: func(ctx context.Context, id string) (*models.Interactive, error) {
					return &models.Interactive{State: models.IsDeleted.String()}, nil
				},
			},
		},
		{
			title:        "WhenAnyOtherDBError_ThenInternalError",
			responseCode: http.StatusInternalServerError,
			mongoServer: &mongoMock.MongoServerMock{
				GetInteractiveFunc: func(ctx context.Context, id string) (*models.Interactive, error) {
					return &models.Interactive{}, errors.New("db-error")
				},
			},
		},
		{
			title:        "WhenAllGood_ThenStatusOK",
			responseCode: http.StatusOK,
			mongoServer: &mongoMock.MongoServerMock{
				GetInteractiveFunc: func(ctx context.Context, id string) (*models.Interactive, error) {
					return &models.Interactive{}, nil
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.title, func(t *testing.T) {
			ctx := context.Background()
			api := api.Setup(ctx, &config.Config{}, mux.NewRouter(), nil, tc.mongoServer, nil, nil)
			resp := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:27050/interactives/%s", interactiveID), nil)
			api.Router.ServeHTTP(resp, req)

			require.Equal(t, tc.responseCode, resp.Result().StatusCode)
		})
	}
}

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, _ := os.Open(path)
	fileContents, _ := ioutil.ReadAll(file)
	fi, _ := file.Stat()
	file.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile(paramName, fi.Name())
	part.Write(fileContents)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	request, err := http.NewRequest("POST", uri, body)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	writer.Close()

	return request, err
}
