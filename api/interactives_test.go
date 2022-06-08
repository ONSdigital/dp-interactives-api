package api_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/ONSdigital/dp-net/v2/responder"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	authorisation "github.com/ONSdigital/dp-authorisation/v2/authorisation/mock"
	"github.com/ONSdigital/dp-interactives-api/api"
	apiMock "github.com/ONSdigital/dp-interactives-api/api/mock"
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
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	respondr              = responder.New()
	on, off               = true, false
	noopGen               = func(string) string { return "" }
	validInteractiveIdGen = func(string) string { return "an-id" }
	getInteractiveFunc    = func(ctx context.Context, id string) (*models.Interactive, error) {
		if id != "" {
			b := &on
			if id == "inactive-id" {
				b = &off
			}
			return &models.Interactive{
				ID:        id,
				SHA:       "sha",
				State:     models.ImportSuccess.String(),
				Active:    b,
				Published: &off,
				Metadata: &models.InteractiveMetadata{
					Title: "title",
					Label: "label",
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
		fs            api.FilesService
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
				{"/v1/interactives", http.MethodPost, http.StatusInternalServerError},
				{"/v1/interactives/an-id", http.MethodPut, http.StatusInternalServerError},
			},
			title:    "WhenValidationPassButS3BucketNotExisting_ThenInternalServerError",
			formFile: "resources/single-interactive.zip",
			mongoServer: &apiMock.MongoServerMock{
				GetActiveInteractiveGivenShaFunc:   func(ctx context.Context, sha string) (*models.Interactive, error) { return nil, nil },
				GetActiveInteractiveGivenFieldFunc: func(ctx context.Context, field, title string) (*models.Interactive, error) { return nil, nil },
				GetInteractiveFunc:                 getInteractiveFunc,
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
			formFile: "resources/single-interactive.zip",
			mongoServer: &apiMock.MongoServerMock{
				GetActiveInteractiveGivenShaFunc:   func(ctx context.Context, sha string) (*models.Interactive, error) { return nil, nil },
				GetActiveInteractiveGivenFieldFunc: func(ctx context.Context, field, title string) (*models.Interactive, error) { return nil, nil },
				GetInteractiveFunc:                 getInteractiveFunc,
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
			formFile: "resources/single-interactive.zip",
			mongoServer: &apiMock.MongoServerMock{
				GetActiveInteractiveGivenShaFunc:   func(ctx context.Context, sha string) (*models.Interactive, error) { return nil, nil },
				GetActiveInteractiveGivenFieldFunc: func(ctx context.Context, field, title string) (*models.Interactive, error) { return nil, nil },
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
			formFile: "resources/single-interactive.zip",
			mongoServer: &apiMock.MongoServerMock{
				GetActiveInteractiveGivenShaFunc:   func(ctx context.Context, sha string) (*models.Interactive, error) { return nil, nil },
				GetActiveInteractiveGivenFieldFunc: func(ctx context.Context, field, title string) (*models.Interactive, error) { return nil, nil },
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

			api := api.Setup(ctx, &config.Config{PublishingEnabled: true}, mux.NewRouter(), newAuthMiddlwareMock(), tc.mongoServer, tc.kafkaProducer, tc.s3, tc.fs, validInteractiveIdGen, noopGen, noopGen, respondr)

			for _, testReq := range tc.requests {
				var req *http.Request
				if tc.formFile != "" {
					req = test_support.NewFileUploadRequest(testReq.method, testReq.uri, "attachment", tc.formFile, &models.Interactive{
						Metadata: &models.InteractiveMetadata{
							Label:      "label1",
							InternalID: "idValue",
							Title:      "title1",
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

func TestUploadInteractivesHandlers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	formFile := "resources/single-interactive.zip"
	s3 := &s3Mock.S3InterfaceMock{
		ValidateBucketFunc: func() error { return nil },
		UploadFunc: func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
			return nil, nil
		},
	}
	kafkaProducer := &kMock.IProducerMock{
		ChannelsFunc: func() *kafka.ProducerChannels { return &kafka.ProducerChannels{Output: nil} },
	}
	fs := &apiMock.FilesServiceMock{}

	mongoServer := &apiMock.MongoServerMock{
		GetActiveInteractiveGivenShaFunc:   func(ctx context.Context, sha string) (*models.Interactive, error) { return nil, nil },
		GetActiveInteractiveGivenFieldFunc: func(ctx context.Context, field, title string) (*models.Interactive, error) { return nil, nil },
		UpsertInteractiveFunc: func(ctx context.Context, id string, vis *models.Interactive) error {
			if id, _ := strconv.Atoi(vis.Metadata.ResourceID); id > 15 {
				return nil
			} else {
				return mongo.WriteException{
					WriteErrors: []mongo.WriteError{
						{
							Index:   1,
							Code:    11000,
							Message: "duplicate key error",
						},
					},
				}
			}
		},
		GetInteractiveFunc: getInteractiveFunc,
	}

	type test struct {
		uri                 string
		method              string
		responseCode        int
		startId             int
		resourceIdFuncCalls int
	}
	tests := []test{
		{"/v1/interactives", http.MethodPost, http.StatusInternalServerError, 0, api.MaxCollisions},
		{"/v1/interactives", http.MethodPost, http.StatusAccepted, 10, 16},
	}
	for _, testReq := range tests {
		callCount := testReq.startId
		resourceIdGen := func(string) string {
			callCount++
			return strconv.Itoa(callCount)
		}

		a := api.Setup(ctx, &config.Config{PublishingEnabled: true}, mux.NewRouter(), newAuthMiddlwareMock(), mongoServer, kafkaProducer, s3, fs, validInteractiveIdGen, resourceIdGen, noopGen, respondr)

		req := test_support.NewFileUploadRequest(testReq.method, testReq.uri, "attachment", formFile, &models.Interactive{
			Metadata: &models.InteractiveMetadata{
				Label:      "label1",
				InternalID: "idValue",
				Title:      "title1",
			},
		})

		resp := httptest.NewRecorder()

		a.Router.ServeHTTP(resp, req)
		require.Equal(t, testReq.responseCode, resp.Result().StatusCode)
		require.Equal(t, testReq.resourceIdFuncCalls, callCount)
	}
}

func TestGetInteractiveMetadataHandler(t *testing.T) {
	t.Parallel()
	interactiveID := "11-22-33-44"
	tests := []struct {
		title             string
		responseCode      int
		mongoServer       api.MongoServer
		publishingEnabled bool
	}{
		{
			title:        "WhenMissingInDatabase_ThenStatusNotFound",
			responseCode: http.StatusNotFound,
			mongoServer: &apiMock.MongoServerMock{
				GetInteractiveFunc: func(ctx context.Context, id string) (*models.Interactive, error) { return nil, nil },
			},
		},
		{
			title:        "WhenInteractiveIsDeleted_ThenStatusNotFound",
			responseCode: http.StatusNotFound,
			mongoServer: &apiMock.MongoServerMock{
				GetInteractiveFunc: func(ctx context.Context, id string) (*models.Interactive, error) {
					return &models.Interactive{Active: &off}, nil
				},
			},
		},
		{
			title:        "WhenAnyOtherDBError_ThenInternalError",
			responseCode: http.StatusInternalServerError,
			mongoServer: &apiMock.MongoServerMock{
				GetInteractiveFunc: func(ctx context.Context, id string) (*models.Interactive, error) {
					return nil, errors.New("db-error")
				},
			},
		},
		{
			title:        "WhenAllGood_ThenStatusOK",
			responseCode: http.StatusOK,
			mongoServer: &apiMock.MongoServerMock{
				GetInteractiveFunc: func(ctx context.Context, id string) (*models.Interactive, error) {
					return &models.Interactive{Active: &on, Published: &on, Metadata: &models.InteractiveMetadata{}}, nil
				},
			},
			publishingEnabled: true,
		},
		{
			title:        "WhenWebAndUnpublished_ThenStatusNotFound",
			responseCode: http.StatusNotFound,
			mongoServer: &apiMock.MongoServerMock{
				GetInteractiveFunc: func(ctx context.Context, id string) (*models.Interactive, error) {
					return &models.Interactive{Active: &on, Published: &off, Metadata: &models.InteractiveMetadata{}}, nil
				},
			},
			publishingEnabled: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.title, func(t *testing.T) {
			ctx := context.Background()
			api := api.Setup(ctx, &config.Config{PublishingEnabled: tc.publishingEnabled}, mux.NewRouter(), newAuthMiddlwareMock(), tc.mongoServer, nil, nil, nil, noopGen, noopGen, noopGen, respondr)
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
