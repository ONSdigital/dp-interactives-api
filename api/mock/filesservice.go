// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-interactives-api/api"
	"sync"
)

// Ensure, that FilesServiceMock does implement api.FilesService.
// If this is not the case, regenerate this file with moq.
var _ api.FilesService = &FilesServiceMock{}

// FilesServiceMock is a mock implementation of api.FilesService.
//
// 	func TestSomethingThatUsesFilesService(t *testing.T) {
//
// 		// make and configure a mocked api.FilesService
// 		mockedFilesService := &FilesServiceMock{
// 			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error {
// 				panic("mock out the Checker method")
// 			},
// 			PublishCollectionFunc: func(ctx context.Context, collectionID string) error {
// 				panic("mock out the PublishCollection method")
// 			},
// 			SetCollectionIDFunc: func(ctx context.Context, file string, collectionID string) error {
// 				panic("mock out the SetCollectionID method")
// 			},
// 		}
//
// 		// use mockedFilesService in code that requires api.FilesService
// 		// and then make assertions.
//
// 	}
type FilesServiceMock struct {
	// CheckerFunc mocks the Checker method.
	CheckerFunc func(ctx context.Context, state *healthcheck.CheckState) error

	// PublishCollectionFunc mocks the PublishCollection method.
	PublishCollectionFunc func(ctx context.Context, collectionID string) error

	// SetCollectionIDFunc mocks the SetCollectionID method.
	SetCollectionIDFunc func(ctx context.Context, file string, collectionID string) error

	// calls tracks calls to the methods.
	calls struct {
		// Checker holds details about calls to the Checker method.
		Checker []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// State is the state argument value.
			State *healthcheck.CheckState
		}
		// PublishCollection holds details about calls to the PublishCollection method.
		PublishCollection []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// CollectionID is the collectionID argument value.
			CollectionID string
		}
		// SetCollectionID holds details about calls to the SetCollectionID method.
		SetCollectionID []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// File is the file argument value.
			File string
			// CollectionID is the collectionID argument value.
			CollectionID string
		}
	}
	lockChecker           sync.RWMutex
	lockPublishCollection sync.RWMutex
	lockSetCollectionID   sync.RWMutex
}

// Checker calls CheckerFunc.
func (mock *FilesServiceMock) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	if mock.CheckerFunc == nil {
		panic("FilesServiceMock.CheckerFunc: method is nil but FilesService.Checker was just called")
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
//     len(mockedFilesService.CheckerCalls())
func (mock *FilesServiceMock) CheckerCalls() []struct {
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

// PublishCollection calls PublishCollectionFunc.
func (mock *FilesServiceMock) PublishCollection(ctx context.Context, collectionID string) error {
	if mock.PublishCollectionFunc == nil {
		panic("FilesServiceMock.PublishCollectionFunc: method is nil but FilesService.PublishCollection was just called")
	}
	callInfo := struct {
		Ctx          context.Context
		CollectionID string
	}{
		Ctx:          ctx,
		CollectionID: collectionID,
	}
	mock.lockPublishCollection.Lock()
	mock.calls.PublishCollection = append(mock.calls.PublishCollection, callInfo)
	mock.lockPublishCollection.Unlock()
	return mock.PublishCollectionFunc(ctx, collectionID)
}

// PublishCollectionCalls gets all the calls that were made to PublishCollection.
// Check the length with:
//     len(mockedFilesService.PublishCollectionCalls())
func (mock *FilesServiceMock) PublishCollectionCalls() []struct {
	Ctx          context.Context
	CollectionID string
} {
	var calls []struct {
		Ctx          context.Context
		CollectionID string
	}
	mock.lockPublishCollection.RLock()
	calls = mock.calls.PublishCollection
	mock.lockPublishCollection.RUnlock()
	return calls
}

// SetCollectionID calls SetCollectionIDFunc.
func (mock *FilesServiceMock) SetCollectionID(ctx context.Context, file string, collectionID string) error {
	if mock.SetCollectionIDFunc == nil {
		panic("FilesServiceMock.SetCollectionIDFunc: method is nil but FilesService.SetCollectionID was just called")
	}
	callInfo := struct {
		Ctx          context.Context
		File         string
		CollectionID string
	}{
		Ctx:          ctx,
		File:         file,
		CollectionID: collectionID,
	}
	mock.lockSetCollectionID.Lock()
	mock.calls.SetCollectionID = append(mock.calls.SetCollectionID, callInfo)
	mock.lockSetCollectionID.Unlock()
	return mock.SetCollectionIDFunc(ctx, file, collectionID)
}

// SetCollectionIDCalls gets all the calls that were made to SetCollectionID.
// Check the length with:
//     len(mockedFilesService.SetCollectionIDCalls())
func (mock *FilesServiceMock) SetCollectionIDCalls() []struct {
	Ctx          context.Context
	File         string
	CollectionID string
} {
	var calls []struct {
		Ctx          context.Context
		File         string
		CollectionID string
	}
	mock.lockSetCollectionID.RLock()
	calls = mock.calls.SetCollectionID
	mock.lockSetCollectionID.RUnlock()
	return calls
}