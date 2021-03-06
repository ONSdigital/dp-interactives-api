// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-interactives-api/api"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"sync"
)

// Ensure, that S3InterfaceMock does implement api.S3Interface.
// If this is not the case, regenerate this file with moq.
var _ api.S3Interface = &S3InterfaceMock{}

// S3InterfaceMock is a mock implementation of api.S3Interface.
//
// 	func TestSomethingThatUsesS3Interface(t *testing.T) {
//
// 		// make and configure a mocked api.S3Interface
// 		mockedS3Interface := &S3InterfaceMock{
// 			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error {
// 				panic("mock out the Checker method")
// 			},
// 			UploadFunc: func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
// 				panic("mock out the Upload method")
// 			},
// 			ValidateBucketFunc: func() error {
// 				panic("mock out the ValidateBucket method")
// 			},
// 		}
//
// 		// use mockedS3Interface in code that requires api.S3Interface
// 		// and then make assertions.
//
// 	}
type S3InterfaceMock struct {
	// CheckerFunc mocks the Checker method.
	CheckerFunc func(ctx context.Context, state *healthcheck.CheckState) error

	// UploadFunc mocks the Upload method.
	UploadFunc func(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)

	// ValidateBucketFunc mocks the ValidateBucket method.
	ValidateBucketFunc func() error

	// calls tracks calls to the methods.
	calls struct {
		// Checker holds details about calls to the Checker method.
		Checker []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// State is the state argument value.
			State *healthcheck.CheckState
		}
		// Upload holds details about calls to the Upload method.
		Upload []struct {
			// Input is the input argument value.
			Input *s3manager.UploadInput
			// Options is the options argument value.
			Options []func(*s3manager.Uploader)
		}
		// ValidateBucket holds details about calls to the ValidateBucket method.
		ValidateBucket []struct {
		}
	}
	lockChecker        sync.RWMutex
	lockUpload         sync.RWMutex
	lockValidateBucket sync.RWMutex
}

// Checker calls CheckerFunc.
func (mock *S3InterfaceMock) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	if mock.CheckerFunc == nil {
		panic("S3InterfaceMock.CheckerFunc: method is nil but S3Interface.Checker was just called")
	}
	callInfo := struct {
		Ctx   context.Context
		State *healthcheck.CheckState
	}{
		Ctx:   ctx,
		State: state,
	}
	mock.lockChecker.Lock()
	mock.calls.Checker = append(mock.calls.Checker, callInfo)
	mock.lockChecker.Unlock()
	return mock.CheckerFunc(ctx, state)
}

// CheckerCalls gets all the calls that were made to Checker.
// Check the length with:
//     len(mockedS3Interface.CheckerCalls())
func (mock *S3InterfaceMock) CheckerCalls() []struct {
	Ctx   context.Context
	State *healthcheck.CheckState
} {
	var calls []struct {
		Ctx   context.Context
		State *healthcheck.CheckState
	}
	mock.lockChecker.RLock()
	calls = mock.calls.Checker
	mock.lockChecker.RUnlock()
	return calls
}

// Upload calls UploadFunc.
func (mock *S3InterfaceMock) Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	if mock.UploadFunc == nil {
		panic("S3InterfaceMock.UploadFunc: method is nil but S3Interface.Upload was just called")
	}
	callInfo := struct {
		Input   *s3manager.UploadInput
		Options []func(*s3manager.Uploader)
	}{
		Input:   input,
		Options: options,
	}
	mock.lockUpload.Lock()
	mock.calls.Upload = append(mock.calls.Upload, callInfo)
	mock.lockUpload.Unlock()
	return mock.UploadFunc(input, options...)
}

// UploadCalls gets all the calls that were made to Upload.
// Check the length with:
//     len(mockedS3Interface.UploadCalls())
func (mock *S3InterfaceMock) UploadCalls() []struct {
	Input   *s3manager.UploadInput
	Options []func(*s3manager.Uploader)
} {
	var calls []struct {
		Input   *s3manager.UploadInput
		Options []func(*s3manager.Uploader)
	}
	mock.lockUpload.RLock()
	calls = mock.calls.Upload
	mock.lockUpload.RUnlock()
	return calls
}

// ValidateBucket calls ValidateBucketFunc.
func (mock *S3InterfaceMock) ValidateBucket() error {
	if mock.ValidateBucketFunc == nil {
		panic("S3InterfaceMock.ValidateBucketFunc: method is nil but S3Interface.ValidateBucket was just called")
	}
	callInfo := struct {
	}{}
	mock.lockValidateBucket.Lock()
	mock.calls.ValidateBucket = append(mock.calls.ValidateBucket, callInfo)
	mock.lockValidateBucket.Unlock()
	return mock.ValidateBucketFunc()
}

// ValidateBucketCalls gets all the calls that were made to ValidateBucket.
// Check the length with:
//     len(mockedS3Interface.ValidateBucketCalls())
func (mock *S3InterfaceMock) ValidateBucketCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockValidateBucket.RLock()
	calls = mock.calls.ValidateBucket
	mock.lockValidateBucket.RUnlock()
	return calls
}
