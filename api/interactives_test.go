package api_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/ONSdigital/dp-interactives-api/api"
	authorisation "github.com/ONSdigital/dp-authorisation/v2/authorisation/mock"
	mongoMock "github.com/ONSdigital/dp-interactives-api/api/mock"
	"github.com/ONSdigital/dp-interactives-api/config"
	test_support "github.com/ONSdigital/dp-interactives-api/internal/test-support"
	"github.com/ONSdigital/dp-interactives-api/models"
	"github.com/ONSdigital/dp-interactives-api/upload"
	s3Mock "github.com/ONSdigital/dp-interactives-api/upload/mock"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	kMock "github.com/ONSdigital/dp-kafka/v3/kafkatest"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	t, f               = true, false
	getInteractiveFunc = func(ctx context.Context, id string) (*models.Interactive, error) {
		if id != "" {
			b := &t
			if id == "inactive-id" {
				b = &f
			}
			return &models.Interactive{
				ID:        id,
				SHA:       "sha",
				State:     models.ImportSuccess.String(),
				Active:    b,
				Published: &f,
				Metadata: &models.InteractiveMetadata{
					Title: "title",
				},
			}, nil
		}
		return nil, nil
	}
)

func TestUploadAndUpdateInteractivesHandlers(t *testing.T) {
	t.Parallel()

	type request struct {
		uri          string
		method       string
		responseCode int
	}
	tests := []struct {
		requests      []request
		title         string
		formFile      string
		mongoServer   api.MongoServer
		s3            upload.S3Interface
		kafkaProducer kafka.IProducer
	}{
		{
			requests: []request{
				{"/v1/interactives", http.MethodPost, http.StatusBadRequest},
				{"/v1/interactives/an-id", http.MethodPut, http.StatusBadRequest},
			},
			title: "WhenMissingAttachment_ThenStatusBadRequest",
		},
		{
			requests: []request{
				{"/v1/interactives", http.MethodPost, http.StatusBadRequest},
				{"/v1/interactives/an-id", http.MethodPut, http.StatusBadRequest},
			},
			title:    "WhenUploadedFileIsNotZip_ThenStatusBadRequest",
			formFile: "resources/fortest.txt",
		},
		{
			requests: []request{
				{"/v1/interactives", http.MethodPost, http.StatusBadRequest},
				{"/v1/interactives/an-id", http.MethodPut, http.StatusBadRequest},
			},
			title:    "WhenFileIsZipButAlreadyExists_ThenStatusBadRequest",
			formFile: "resources/interactives.zip",
			mongoServer: &mongoMock.MongoServerMock{
				GetActiveInteractiveGivenShaFunc: func(ctx context.Context, sha string) (*models.Interactive, error) { return &models.Interactive{}, nil },
			},
		},
		{
			requests: []request{
				{"/v1/interactives", http.MethodPost, http.StatusInternalServerError},
				{"/v1/interactives/an-id", http.MethodPut, http.StatusInternalServerError},
			},
			title:    "WhenValidationPassButS3BucketNotExisting_ThenInternalServerError",
			formFile: "resources/interactives.zip",
			mongoServer: &mongoMock.MongoServerMock{
				GetActiveInteractiveGivenShaFunc:   func(ctx context.Context, sha string) (*models.Interactive, error) { return nil, nil },
				GetActiveInteractiveGivenTitleFunc: func(ctx context.Context, title string) (*models.Interactive, error) { return nil, nil },
			},
			s3: &s3Mock.S3InterfaceMock{
				ValidateBucketFunc: func() error { return errors.New("s3 error") },
			},
		},
		{
			requests: []request{
				{"/v1/interactives", http.MethodPost, http.StatusInternalServerError},
				{"/v1/interactives/an-id", http.MethodPut, http.StatusInternalServerError},
			},
			title:    "WhenUploadError_ThenInternalServerError",
			formFile: "resources/interactives.zip",
			mongoServer: &mongoMock.MongoServerMock{
				GetActiveInteractiveGivenShaFunc:   func(ctx context.Context, sha string) (*models.Interactive, error) { return nil, nil },
				GetActiveInteractiveGivenTitleFunc: func(ctx context.Context, title string) (*models.Interactive, error) { return nil, nil },
			},
			s3: &s3Mock.S3InterfaceMock{
				ValidateBucketFunc: func() error { return nil },
				UploadFunc: func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
					return nil, errors.New("upload error")
				},
			},
		},
		{
			requests: []request{
				{"/v1/interactives", http.MethodPost, http.StatusInternalServerError},
				{"/v1/interactives/an-id", http.MethodPut, http.StatusInternalServerError},
			},
			title:    "WhenDbError_ThenInternalServerError",
			formFile: "resources/interactives.zip",
			mongoServer: &mongoMock.MongoServerMock{
				GetActiveInteractiveGivenShaFunc:   func(ctx context.Context, sha string) (*models.Interactive, error) { return nil, nil },
				GetActiveInteractiveGivenTitleFunc: func(ctx context.Context, title string) (*models.Interactive, error) { return nil, nil },
				UpsertInteractiveFunc: func(ctx context.Context, id string, vis *models.Interactive) error {
					return errors.New("db upsert error")
				},
				GetInteractiveFunc: getInteractiveFunc,
			},
			s3: &s3Mock.S3InterfaceMock{
				ValidateBucketFunc: func() error { return nil },
				UploadFunc: func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
					return nil, nil
				},
			},
		},
		{
			requests: []request{
				{"/v1/interactives", http.MethodPost, http.StatusAccepted},
				{"/v1/interactives/an-id", http.MethodPut, http.StatusOK},
				{"/v1/interactives/inactive-id", http.MethodPut, http.StatusNotFound},
			},
			title:    "WhenAllSuccess_ThenStatus20x",
			formFile: "resources/interactives.zip",
			mongoServer: &mongoMock.MongoServerMock{
				GetActiveInteractiveGivenShaFunc:   func(ctx context.Context, sha string) (*models.Interactive, error) { return nil, nil },
				GetActiveInteractiveGivenTitleFunc: func(ctx context.Context, title string) (*models.Interactive, error) { return nil, nil },
				UpsertInteractiveFunc: func(ctx context.Context, id string, vis *models.Interactive) error {
					return nil
				},
				GetInteractiveFunc: getInteractiveFunc,
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

			api := api.Setup(ctx, &config.Config{}, mux.NewRouter(), newAuthMiddlwareMock(), tc.mongoServer, tc.kafkaProducer, tc.s3)

			for _, testReq := range tc.requests {
				var req *http.Request
				if tc.formFile != "" {
					req = test_support.NewFileUploadRequest(testReq.method, testReq.uri, "attachment", tc.formFile, &models.InteractiveUpdate{
						Interactive: models.Interactive{
							Metadata: &models.InteractiveMetadata{
								Title: "value1",
								Uri:   "value2",
							},
						},
					})
				} else {
					req = httptest.NewRequest(testReq.method, testReq.uri, nil)
				}

				resp := httptest.NewRecorder()

				api.Router.ServeHTTP(resp, req)
				require.Equal(t, testReq.responseCode, resp.Result().StatusCode)
			}
		})
	}
}

func TestGetInteractiveMetadataHandler(t *testing.T) {
	t.Parallel()
	activeFlag := true
	inactiveFlag := false
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
					return &models.Interactive{Active: &inactiveFlag}, nil
				},
			},
		},
		{
			title:        "WhenAnyOtherDBError_ThenInternalError",
			responseCode: http.StatusInternalServerError,
			mongoServer: &mongoMock.MongoServerMock{
				GetInteractiveFunc: func(ctx context.Context, id string) (*models.Interactive, error) {
					return &models.Interactive{Active: &activeFlag}, errors.New("db-error")
				},
			},
		},
		{
			title:        "WhenAllGood_ThenStatusOK",
			responseCode: http.StatusOK,
			mongoServer: &mongoMock.MongoServerMock{
				GetInteractiveFunc: func(ctx context.Context, id string) (*models.Interactive, error) {
					return &models.Interactive{Active: &activeFlag, Metadata: &models.InteractiveMetadata{}}, nil
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.title, func(t *testing.T) {
			ctx := context.Background()
			api := api.Setup(ctx, &config.Config{}, mux.NewRouter(), newAuthMiddlwareMock(), tc.mongoServer, nil, nil)
			resp := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:27050/v1/interactives/%s", interactiveID), nil)
			api.Router.ServeHTTP(resp, req)

			require.Equal(t, tc.responseCode, resp.Result().StatusCode)
		})
	}
}

func newAuthMiddlwareMock() *authorisation.MiddlewareMock {
	return &authorisation.MiddlewareMock{
		RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
			return handlerFunc
		},
	}
}