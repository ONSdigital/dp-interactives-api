package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/ONSdigital/dp-interactives-api/event"
	"github.com/ONSdigital/dp-interactives-api/models"
	"github.com/ONSdigital/dp-interactives-api/mongo"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
)

const (
	MaxCollisions          = 10
	MarshallingErrorCode   = "ErrMarshalling"
	EventPublishErrorCode  = "ErrPublishingEvent"
	DbErrorCode            = "ErrDB"
	UploadErrorCode        = "ErrUpload"
	RequestErrorCode       = "ErrRequest"
	DownstreamAPIErrorCode = "ErrDownstreamAPI"
)

var (
	enabled, disabled        = true, false
	ErrInvalidBody           = errors.New("body has invalid format")
	ErrCantUpdateMeta        = errors.New("cannot update metadata for a published interactive")
	ErrCantDeletePublishedIn = errors.New("cannot delete a published interactive")
	ErrPubErrNoCollectionID  = errors.New("cannot publish interactive, no collection ID")
)

func (api *API) UploadInteractivesHandler(ctx context.Context, _ http.ResponseWriter, req *http.Request) (*models.SuccessResponse, *models.ErrorResponse) {
	// 1. Validate request
	formDataRequest, err := newFormDataRequest(req, api, WantOnlyOneAttachmentWithMetadata, true)
	if err != nil {
		responseErr := models.NewError(ctx, err, RequestErrorCode, fmt.Errorf("request validation failed %w", err).Error())
		return nil, models.NewErrorResponse(http.StatusBadRequest, nil, responseErr)
	}

	if api.cfg.ValidateSHAEnabled {
		// 2. Check if duplicate exists
		existing, _ := api.mongoDB.GetActiveInteractiveGivenSha(ctx, formDataRequest.Sha)
		if existing != nil {
			err = fmt.Errorf("archive already exists id (%s) with sha (%s)", existing.ID, existing.SHA)
			responseErr := models.NewError(ctx, err, RequestErrorCode, err.Error())
			return nil, models.NewErrorResponse(http.StatusBadRequest, nil, responseErr)
		}
	}

	// 3. Check "label + title are unique"
	update := formDataRequest.Update.Interactive
	existing, _ := api.mongoDB.GetActiveInteractiveGivenField(ctx, "metadata.label", update.Metadata.Label)
	if existing != nil {
		err = fmt.Errorf("archive with label (%s) already exists id (%s)", existing.Metadata.Label, existing.ID)
		responseErr := models.NewError(ctx, err, RequestErrorCode, err.Error())
		return nil, models.NewErrorResponse(http.StatusBadRequest, nil, responseErr)
	}
	existing, _ = api.mongoDB.GetActiveInteractiveGivenField(ctx, "metadata.title", update.Metadata.Title)
	if existing != nil {
		err = fmt.Errorf("archive with title (%s) already exists id (%s)", existing.Metadata.Title, existing.ID)
		responseErr := models.NewError(ctx, err, RequestErrorCode, err.Error())
		return nil, models.NewErrorResponse(http.StatusBadRequest, nil, responseErr)
	}

	// 4. upload to S3
	uri, err := api.uploadFile(formDataRequest.Sha, formDataRequest.FileName, formDataRequest.FileData)
	if err != nil {
		responseErr := models.NewError(ctx, err, UploadErrorCode, fmt.Errorf("unable to upload %w", err).Error())
		return nil, models.NewErrorResponse(http.StatusInternalServerError, nil, responseErr)
	}

	// 5. Write to DB
	id := api.newUUID("")
	interact := &models.Interactive{
		ID:        id,
		SHA:       formDataRequest.Sha,
		Active:    &enabled,
		Published: &disabled,
		State:     models.ArchiveUploaded.String(),
		Archive:   &models.Archive{Name: uri},
	}
	collisions := 0
	for {
		update.Metadata.ResourceID = api.newResourceID("")
		update.Metadata.HumanReadableSlug = api.newSlug(update.Metadata.Label)
		interact.Metadata = update.Metadata

		err = api.mongoDB.UpsertInteractive(ctx, id, interact)
		if err == nil {
			break
		}

		if mongoDriver.IsDuplicateKeyError(err) {
			collisions++
		} else {
			responseErr := models.NewError(ctx, err, DbErrorCode, fmt.Errorf("unable to write to DB %w", err).Error())
			return nil, models.NewErrorResponse(http.StatusInternalServerError, nil, responseErr)
		}

		if collisions == MaxCollisions {
			responseErr := models.NewError(ctx, err, DbErrorCode, fmt.Errorf("unable to write to DB - max collisions %w", err).Error())
			return nil, models.NewErrorResponse(http.StatusInternalServerError, nil, responseErr)
		}
	}

	interactive, err := api.mongoDB.GetInteractive(ctx, id)
	if err != nil {
		responseErr := models.NewError(ctx, err, DbErrorCode, fmt.Errorf("error fetching interactive %s %w", id, err).Error())
		return nil, models.NewErrorResponse(http.StatusInternalServerError, nil, responseErr)
	}

	// 5. send kafka message to importer
	err = api.producer.InteractiveUploaded(&event.InteractiveUploaded{
		ID:           id,
		FilePath:     uri,
		Title:        interactive.Metadata.Title,
		CurrentFiles: []string{""}, //need to send an empty val :(
	})
	if err != nil {
		responseErr := models.NewError(ctx, err, EventPublishErrorCode, fmt.Errorf("unable to notify importer %w", err).Error())
		return nil, models.NewErrorResponse(http.StatusInternalServerError, nil, responseErr)
	}

	jsonB, err := JSONify(interactive)
	if err != nil {
		responseErr := models.NewError(ctx, err, MarshallingErrorCode, err.Error())
		return nil, models.NewErrorResponse(http.StatusInternalServerError, nil, responseErr)
	}

	return models.NewSuccessResponse(jsonB, http.StatusAccepted), nil
}

func (api *API) GetInteractiveMetadataHandler(ctx context.Context, _ http.ResponseWriter, req *http.Request) (*models.SuccessResponse, *models.ErrorResponse) {
	// get id
	vars := mux.Vars(req)
	id := vars["id"]

	// fetch info from DB
	i, err := api.mongoDB.GetInteractive(ctx, id)
	if err != nil && err != mongo.ErrNoRecordFound {
		responseErr := models.NewError(ctx, err, DbErrorCode, fmt.Errorf("error fetching interactive %s %w", id, err).Error())
		return nil, models.NewErrorResponse(http.StatusInternalServerError, nil, responseErr)
	}

	//if mongo.ErrNoRecordFound then will blockAccess(i==nil)
	if api.blockAccess(i) {
		responseErr := models.NewError(ctx, err, DbErrorCode, fmt.Errorf("interactive-id (%s) is either deleted or does not exist", id).Error())
		return nil, models.NewErrorResponse(http.StatusNotFound, nil, responseErr)
	}

	jsonB, err := JSONify(i)
	if err != nil {
		responseErr := models.NewError(ctx, err, MarshallingErrorCode, err.Error())
		return nil, models.NewErrorResponse(http.StatusInternalServerError, nil, responseErr)
	}

	return models.NewSuccessResponse(jsonB, http.StatusOK), nil
}

// update rules
// if published - allow only file updates
// if unpublished - allow both file + metadata
func (api *API) UpdateInteractiveHandler(ctx context.Context, w http.ResponseWriter, req *http.Request) (*models.SuccessResponse, *models.ErrorResponse) {
	// Validate request
	formDataRequest, err := newFormDataRequest(req, api, WantAtleastMaxOneAttachmentAndOrMetadata, false)
	if err != nil {
		//if validationErrs, ok := err.(validator.ValidationErrors); ok {
		//	writeError(w, buildValidationErrors(validationErrs), http.StatusBadRequest)
		//	return
		//}

		responseErr := models.NewError(ctx, err, RequestErrorCode, fmt.Errorf("request validation failed %w", err).Error())
		return nil, models.NewErrorResponse(http.StatusBadRequest, nil, responseErr)
	}

	// Check that id exists and is not deleted
	vars := mux.Vars(req)
	id := vars["id"]
	existing, err := api.mongoDB.GetInteractive(ctx, id)
	if (existing == nil && err == nil) || err == mongo.ErrNoRecordFound || (existing != nil && !*existing.Active) {
		err = fmt.Errorf("interactive-id (%s) is either deleted or does not exist", id)
		responseErr := models.NewError(ctx, err, RequestErrorCode, err.Error())
		return nil, models.NewErrorResponse(http.StatusNotFound, nil, responseErr)
	}
	if err != nil {
		responseErr := models.NewError(ctx, err, DbErrorCode, fmt.Errorf("error fetching interactive %s %w", id, err).Error())
		return nil, models.NewErrorResponse(http.StatusInternalServerError, nil, responseErr)
	}

	// fail if attempting to update metadata for a published model
	if *existing.Published && formDataRequest.hasMetadata() {
		responseErr := models.NewError(ctx, err, DbErrorCode, ErrCantUpdateMeta.Error())
		return nil, models.NewErrorResponse(http.StatusForbidden, nil, responseErr)
	}

	// -- ALL GOOD ABOVE

	// prepare updated model
	updatedModel := &models.Interactive{
		ID:        id,
		Published: existing.Published,
		State:     existing.State,
	}
	updatedModel.Metadata = updatedModel.Metadata.Update(existing.Metadata, api.newSlug)

	var update models.Interactive
	if formDataRequest.Update == nil { // no metada update
		update.Metadata = update.Metadata.Update(existing.Metadata, api.newSlug)
		update.Published = existing.Published
	} else {
		update = formDataRequest.Update.Interactive

		if update.Metadata != nil {
			updatedModel.Metadata = updatedModel.Metadata.Update(update.Metadata, api.newSlug)
		}
		if update.Published != nil {
			updatedModel.Published = update.Published
		}
	}

	if formDataRequest.Update != nil {
		if formDataRequest.Update.ImportSuccessful != nil { // importer updates
			if *formDataRequest.Update.ImportSuccessful {
				updatedModel.State = models.ImportSuccess.String()
			} else {
				updatedModel.State = models.ImportFailure.String()
			}
			updatedModel.ImportMessage = &formDataRequest.Update.ImportMessage
		}
	}

	if update.Archive != nil {
		updatedModel.Archive = &models.Archive{
			Name: update.Archive.Name,
			Size: update.Archive.Size,
		}
		for _, f := range update.Archive.Files {
			updatedModel.Archive.Files = append(updatedModel.Archive.Files, &models.File{
				Name:     f.Name,
				Mimetype: f.Mimetype,
				Size:     f.Size,
			})
		}
	}

	// Finally check if file to be uploaded
	uri := ""
	if formDataRequest.FileData != nil {
		if api.cfg.ValidateSHAEnabled {
			// Check if duplicate SHA exists
			i, _ := api.mongoDB.GetActiveInteractiveGivenSha(ctx, formDataRequest.Sha)
			if i != nil {
				err = fmt.Errorf("archive already exists id (%s) with sha (%s)", i.ID, i.SHA)
				responseErr := models.NewError(ctx, err, RequestErrorCode, err.Error())
				return nil, models.NewErrorResponse(http.StatusBadRequest, nil, responseErr)
			}
		}

		// Process form data (S3)
		uri, err = api.uploadFile(formDataRequest.Sha, formDataRequest.FileName, formDataRequest.FileData)
		if err != nil {
			responseErr := models.NewError(ctx, err, UploadErrorCode, fmt.Errorf("unable to upload %w", err).Error())
			return nil, models.NewErrorResponse(http.StatusInternalServerError, nil, responseErr)
		}

		updatedModel.State = models.ArchiveUploaded.String()
		updatedModel.SHA = formDataRequest.Sha
	}

	// link with collection-id
	// a collectionID is present in the update message
	if update.Metadata != nil && update.Metadata.CollectionID != "" {
		colID := update.Metadata.CollectionID
		// update if empty OR different
		if existing.Metadata.CollectionID == "" || colID != existing.Metadata.CollectionID {
			// files/archive can be in the update or existing - check update first
			arch := update.Archive
			if arch == nil || len(arch.Files) == 0 {
				arch = existing.Archive
			}
			if arch != nil && len(arch.Files) > 0 {
				for _, file := range arch.Files {
					if err := api.filesService.SetCollectionID(ctx, file.Name, colID); err != nil {
						responseErr := models.NewError(ctx, err, DownstreamAPIErrorCode, fmt.Errorf("error setting collectionID %s %s %w", id, colID, err).Error())
						return nil, models.NewErrorResponse(http.StatusInternalServerError, nil, responseErr)
					}
				}
				updatedModel.Metadata.CollectionID = colID
			}
		}
	}

	// publish (if not already)
	if existing.Published != nil && !*(existing.Published) &&
		update.Published != nil && *(update.Published) {
		collID := update.Metadata.CollectionID
		if collID == "" {
			collID = existing.Metadata.CollectionID
		}

		if collID != "" {
			if err := api.filesService.PublishCollection(ctx, collID); err != nil {
				responseErr := models.NewError(ctx, err, DownstreamAPIErrorCode, fmt.Errorf("error publishing collectionID %s %s %w", id, collID, err).Error())
				return nil, models.NewErrorResponse(http.StatusInternalServerError, nil, responseErr)
			}
			updatedModel.Published = &enabled
		} else {
			log.Error(ctx, fmt.Sprintf("no collection id for interactive (%s)", existing.ID), ErrPubErrNoCollectionID)
		}
	}

	// write to DB
	err = api.mongoDB.UpsertInteractive(ctx, id, updatedModel)
	if err != nil {
		responseErr := models.NewError(ctx, err, DbErrorCode, fmt.Errorf("unable to write to DB %w", err).Error())
		return nil, models.NewErrorResponse(http.StatusInternalServerError, nil, responseErr)
	}

	// get updated model
	interactive, err := api.mongoDB.GetInteractive(ctx, id)
	if err != nil {
		responseErr := models.NewError(ctx, err, DbErrorCode, fmt.Errorf("error fetching interactive %s %w", id, err).Error())
		return nil, models.NewErrorResponse(http.StatusInternalServerError, nil, responseErr)
	}

	// send kafka message to importer (if file uploaded)
	if uri != "" {
		//need to send at least one value
		//https://github.com/go-avro/avro/pull/20
		//https://github.com/go-avro/avro/issues/33 (we should update tbh)
		currentFiles := []string{""}
		if interactive.Archive != nil {
			for _, f := range interactive.Archive.Files {
				currentFiles = append(currentFiles, f.Name)
			}
		}
		err = api.producer.InteractiveUploaded(&event.InteractiveUploaded{
			ID:           id,
			FilePath:     uri,
			Title:        interactive.Metadata.Title,
			CurrentFiles: currentFiles,
		})
		if err != nil {
			responseErr := models.NewError(ctx, err, EventPublishErrorCode, fmt.Errorf("unable to notify importer %w", err).Error())
			return nil, models.NewErrorResponse(http.StatusInternalServerError, nil, responseErr)
		}
	}

	jsonB, err := JSONify(interactive)
	if err != nil {
		responseErr := models.NewError(ctx, err, MarshallingErrorCode, err.Error())
		return nil, models.NewErrorResponse(http.StatusInternalServerError, nil, responseErr)
	}

	return models.NewSuccessResponse(jsonB, http.StatusOK), nil
}

func (api *API) ListInteractivesHandler(req *http.Request, limit int, offset int) (interface{}, int, error) {
	ctx := req.Context()
	var filter *models.InteractiveMetadata

	filterJson := req.URL.Query().Get("filter")
	if filterJson != "" {
		defer req.Body.Close()
		filter = &models.InteractiveMetadata{}

		if err := json.Unmarshal([]byte(filterJson), &filter); err != nil {
			return nil, 0, fmt.Errorf("error unmarshalling body %w", ErrInvalidBody)
		}
	}

	db, _, err := api.mongoDB.ListInteractives(ctx, offset, limit, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("api endpoint getDatasets datastore.GetDatasets returned an error %w", err)
	}

	response := make([]*models.Interactive, 0)
	for _, i := range db {
		if !api.blockAccess(i) {
			response = append(response, i)
		}
	}

	return response, len(response), nil
}

func (api *API) DeleteInteractivesHandler(ctx context.Context, w http.ResponseWriter, req *http.Request) (*models.SuccessResponse, *models.ErrorResponse) {
	// get id
	vars := mux.Vars(req)
	id := vars["id"]

	// error if it doesnt exist
	vis, err := api.mongoDB.GetInteractive(ctx, id)
	if (vis == nil && err == nil) || err == mongo.ErrNoRecordFound {
		err = fmt.Errorf("interactive-id (%s) does not exist", id)
		responseErr := models.NewError(ctx, err, RequestErrorCode, err.Error())
		return nil, models.NewErrorResponse(http.StatusNotFound, nil, responseErr)
	}
	if err != nil {
		responseErr := models.NewError(ctx, err, DbErrorCode, fmt.Errorf("error fetching interactive %s %w", id, err).Error())
		return nil, models.NewErrorResponse(http.StatusInternalServerError, nil, responseErr)
	}

	// must not delete published interactives
	if *vis.Published {
		responseErr := models.NewError(ctx, err, DbErrorCode, ErrCantDeletePublishedIn.Error())
		return nil, models.NewErrorResponse(http.StatusForbidden, nil, responseErr)
	}

	// set to inactive
	err = api.mongoDB.UpsertInteractive(ctx, id, &models.Interactive{
		Active: &disabled,
	})
	if err != nil {
		responseErr := models.NewError(ctx, err, DbErrorCode, fmt.Errorf("unable to unset active flag %s %w", id, err).Error())
		return nil, models.NewErrorResponse(http.StatusInternalServerError, nil, responseErr)
	}

	return models.NewSuccessResponse([]byte{}, http.StatusOK), nil
}
