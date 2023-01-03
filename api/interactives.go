package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/ONSdigital/dp-api-clients-go/v2/interactives"
	"github.com/ONSdigital/dp-interactives-api/event"
	"github.com/ONSdigital/dp-interactives-api/internal/zip"
	"github.com/ONSdigital/dp-interactives-api/models"
	"github.com/ONSdigital/dp-interactives-api/mongo"
	"github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
)

const (
	MaxCollisions = 10
)

var (
	enabled, disabled        = true, false
	ErrInvalidBody           = errors.New("body has invalid format")
	ErrCantDeletePublishedIn = errors.New("cannot delete a published interactive")
)

func (api *API) UploadInteractivesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.Info(ctx, "upload interactives")

	// Validate request
	formDataRequest, errs := newFormDataRequest(r, api, WantOnlyOneAttachmentWithMetadata, true)
	if errs != nil {
		api.respond.Errors(ctx, w, http.StatusBadRequest, errs)
		return
	}
	archive, htmlFiles, err := zip.Open(formDataRequest.TmpFileName)
	if err != nil {
		api.respond.Error(ctx, w, http.StatusBadRequest, fmt.Errorf("unable to open file %w", err))
		return
	}

	update := formDataRequest.Interactive

	// Write to DB
	id := api.newUUID("")
	interact := &models.Interactive{
		ID:        id,
		Active:    &enabled,
		Published: &disabled,
		State:     models.ArchiveUploading.String(),
		Archive:   archive,
		HTMLFiles: htmlFiles,
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
			api.respond.Error(ctx, w, http.StatusInternalServerError, fmt.Errorf("unable to write to DB %w", err))
			return
		}

		if collisions == MaxCollisions {
			api.respond.Error(ctx, w, http.StatusInternalServerError, fmt.Errorf("unable to write to DB - max collisions %w", err))
			return
		}
	}

	interactive, err := api.mongoDB.GetInteractive(ctx, id)
	if err != nil {
		api.respond.Error(ctx, w, http.StatusInternalServerError, fmt.Errorf("error fetching interactive %s %w", id, err))
		return
	}

	// dont hang on to the old context
	requestID := request.GetRequestId(ctx)
	newCtx := request.WithRequestId(context.Background(), requestID)
	go api.uploadAsync(newCtx, interactive, formDataRequest.TmpFileName, formDataRequest.Name)

	api.respond.JSON(ctx, w, http.StatusAccepted, interactive)
}

func (api *API) GetInteractiveHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	interactive, status, err := api.GetInteractive(ctx, r)
	if err != nil {
		api.respond.Error(ctx, w, status, err)
		return
	}

	api.respond.JSON(ctx, w, http.StatusOK, interactive)
}

// update rules
// if published - allow only file updates
// if unpublished - allow both file + metadata
func (api *API) UpdateInteractiveHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]
	log.Info(ctx, "list interactives", log.Data{"_id": id})

	// Validate request
	formDataRequest, errs := newFormDataRequest(r, api, WantAtleastMaxOneAttachmentAndOrMetadata, false)
	if errs != nil {
		api.respond.Errors(ctx, w, http.StatusBadRequest, errs)
		return
	}

	// Check that id exists and is not deleted
	existing, err := api.mongoDB.GetInteractive(ctx, id)
	if (existing == nil && err == nil) || err == mongo.ErrNoRecordFound || (existing != nil && !*existing.Active) {
		api.respond.Error(ctx, w, http.StatusNotFound, fmt.Errorf("interactive-id (%s) is either deleted or does not exist", id))
		return
	}
	if err != nil {
		api.respond.Error(ctx, w, http.StatusInternalServerError, fmt.Errorf("error fetching interactive %s %w", id, err))
		return
	}

	// prepare updated model
	updatedModel := &models.Interactive{
		ID:        id,
		Published: existing.Published,
		State:     existing.State,
		Archive:   existing.Archive,
		Metadata:  existing.Metadata,
	}

	update := formDataRequest.Interactive
	if update != nil {
		if update.Metadata != nil {
			updatedModel.Metadata = updatedModel.Metadata.Update(update.Metadata, api.newSlug)
		}
	}

	// Finally check if file to be uploaded
	if formDataRequest.TmpFileName != "" {
		archive, htmlFiles, err := zip.Open(formDataRequest.TmpFileName)
		if err != nil {
			api.respond.Error(ctx, w, http.StatusBadRequest, fmt.Errorf("unable to open file %w", err))
			return
		}

		updatedModel.Archive = archive
		updatedModel.HTMLFiles = htmlFiles
		updatedModel.State = models.ArchiveUploading.String()
	}

	// write to DB
	err = api.mongoDB.UpsertInteractive(ctx, id, updatedModel)
	if err != nil {
		api.respond.Error(ctx, w, http.StatusInternalServerError, fmt.Errorf("unable to write to DB %w", err))
		return
	}

	// get updated model
	interactive, err := api.mongoDB.GetInteractive(ctx, id)
	if err != nil {
		api.respond.Error(ctx, w, http.StatusInternalServerError, fmt.Errorf("error fetching interactive %s %w", id, err))
		return
	}

	// upload (async) if file is present
	if formDataRequest.TmpFileName != "" {
		requestID := request.GetRequestId(ctx)
		newCtx := request.WithRequestId(context.Background(), requestID)
		go api.uploadAsync(newCtx, interactive, formDataRequest.TmpFileName, formDataRequest.Name)
	}

	api.respond.JSON(ctx, w, http.StatusOK, interactive)
}

// dedicated publish collection as multiple interactives can be a part of a single collection
func (api *API) PublishCollectionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	collectionId := vars["id"]
	log.Info(ctx, "publish collection", log.Data{"collection_id": collectionId})
	publish := true

	ix, err := api.mongoDB.ListInteractives(ctx, &models.Filter{Metadata: &models.Metadata{CollectionID: collectionId}})
	if err != nil {
		api.respond.Error(ctx, w, http.StatusInternalServerError, err)
		return
	}
	if len(ix) <= 0 { // There are no interactives in the collection, just return
		api.respond.JSON(ctx, w, http.StatusOK, nil)
		return
	}
	errInteractives := ""
	for _, inter := range ix {
		if !inter.CanPublish() {
			errInteractives = errInteractives + inter.ID + ", "
		}
	}
	if errInteractives != "" {
		api.respond.Error(ctx, w, http.StatusConflict, fmt.Errorf("interactive(s) not in correct state %s", errInteractives))
		return
	}

	for _, inter := range ix {
		inter.Published = &publish

		if err := api.mongoDB.PatchInteractive(ctx, interactives.Publish, inter); err != nil {
			errInteractives = errInteractives + inter.ID + ", "
			log.Error(ctx, fmt.Sprintf("error setting publish state for interactive [%s]", inter.ID), err)
		}
	}
	if errInteractives != "" {
		api.respond.Error(ctx, w, http.StatusInternalServerError, fmt.Errorf("failed to update publish state interactive(s) [%s]", errInteractives))
		return
	}

	api.respond.JSON(ctx, w, http.StatusOK, nil)
}

func (api *API) PatchInteractiveHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.Info(ctx, "patch interactive")
	i, status, err := api.GetInteractive(ctx, r)
	if err != nil {
		api.respond.Error(ctx, w, status, err)
		return
	}

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		api.respond.Error(ctx, w, http.StatusBadRequest, fmt.Errorf("cannot read request body %w", err))
		return
	}

	var patchReq interactives.PatchRequest
	if err := json.Unmarshal(bytes, &patchReq); err != nil {
		api.respond.Error(ctx, w, http.StatusBadRequest, fmt.Errorf("cannot unmarshal request body %w", err))
		return
	}

	switch patchReq.Attribute {
	case interactives.PatchArchive:
		if patchReq.Interactive.Archive == nil {
			api.respond.Error(ctx, w, http.StatusBadRequest, fmt.Errorf("no archive to patch"))
		}

		i.State = models.ImportFailure.String()
		if patchReq.Interactive.Archive.ImportSuccessful {
			i.State = models.ImportSuccess.String()
		}

		i.Archive = &models.Archive{
			Name:                patchReq.Interactive.Archive.Name,
			Size:                patchReq.Interactive.Archive.Size,
			ImportMessage:       patchReq.Interactive.Archive.ImportMessage,
			UploadRootDirectory: patchReq.Interactive.Archive.UploadRootDirectory,
		}

		err = api.mongoDB.PatchInteractive(ctx, patchReq.Attribute, i)
		if err != nil {
			api.respond.Error(ctx, w, http.StatusInternalServerError, fmt.Errorf("error patching interactive %s %w", i.ID, err))
			return
		}
	case interactives.LinkToCollection:
		if patchReq.Interactive.Metadata != nil && patchReq.Interactive.Metadata.CollectionID != "" {
			i.Metadata.CollectionID = patchReq.Interactive.Metadata.CollectionID
		}

		err = api.mongoDB.PatchInteractive(ctx, patchReq.Attribute, i)
		if err != nil {
			api.respond.Error(ctx, w, http.StatusInternalServerError, fmt.Errorf("error patching interactive %s %w", i.ID, err))
			return
		}
	default:
		api.respond.Error(ctx, w, http.StatusBadRequest, fmt.Errorf("unsuppported attribute %s", patchReq.Attribute))
		return
	}

	api.GetInteractiveHandler(w, r)
}

func (api *API) ListInteractivesHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	var filter *models.Filter

	filterJson := req.URL.Query().Get("filter")
	log.Info(ctx, "list interactives", log.Data{"filter": filterJson})
	if filterJson != "" {
		defer req.Body.Close()
		filter = &models.Filter{}

		if err := json.Unmarshal([]byte(filterJson), &filter); err != nil {
			api.respond.Error(ctx, w, http.StatusInternalServerError, fmt.Errorf("error unmarshalling body %w", ErrInvalidBody))
			return
		}
	}

	db, err := api.mongoDB.ListInteractives(ctx, filter)
	if err != nil {
		api.respond.Error(ctx, w, http.StatusInternalServerError, fmt.Errorf("api endpoint getDatasets datastore.GetDatasets returned an error %w", err))
		return
	}

	response := make([]*models.Interactive, 0)
	for _, i := range db {
		if !api.blockAccess(i) {
			response = append(response, i)
		}
	}

	api.respond.JSON(ctx, w, http.StatusOK, response)
}

func (api *API) DeleteInteractivesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]

	// error if it doesnt exist
	vis, err := api.mongoDB.GetInteractive(ctx, id)
	if (vis == nil && err == nil) || err == mongo.ErrNoRecordFound {
		api.respond.Error(ctx, w, http.StatusNotFound, fmt.Errorf("interactive-id (%s) does not exist", id))
		return
	}
	if err != nil {
		api.respond.Error(ctx, w, http.StatusInternalServerError, fmt.Errorf("error fetching interactive %s %w", id, err))
		return
	}

	// must not delete published interactives
	if *vis.Published {
		api.respond.Error(ctx, w, http.StatusForbidden, ErrCantDeletePublishedIn)
		return
	}

	// set to inactive
	err = api.mongoDB.UpsertInteractive(ctx, id, &models.Interactive{
		Active: &disabled,
	})
	if err != nil {
		api.respond.Error(ctx, w, http.StatusInternalServerError, fmt.Errorf("unable to unset active flag %s %w", id, err))
		return
	}

	api.respond.JSON(ctx, w, http.StatusNoContent, nil)
}

func (api *API) GetInteractive(ctx context.Context, req *http.Request) (*models.Interactive, int, error) {
	vars := mux.Vars(req)
	id := vars["id"]
	log.Info(ctx, "GetInteractive", log.Data{"_id": id})
	// fetch info from DB
	i, err := api.mongoDB.GetInteractive(ctx, id)
	if err != nil && err != mongo.ErrNoRecordFound {
		return nil, http.StatusInternalServerError, fmt.Errorf("error fetching interactive %s %w", id, err)
	}

	//if mongo.ErrNoRecordFound then will blockAccess(i==nil)
	if api.blockAccess(i) {
		return nil, http.StatusNotFound, fmt.Errorf("interactive either deleted or does not exist %s %w", id, err)
	}

	return i, http.StatusOK, nil
}

func (api *API) uploadAsync(ctx context.Context, ix *models.Interactive, tmpFileName, name string) {
	defer os.Remove(tmpFileName)
	// Upload to S3
	uri, err := api.uploadFile(tmpFileName, name)
	if err != nil {
		log.Error(ctx, fmt.Sprintf("error uploading [%s] to s3 bucket", tmpFileName), err)
		ix.State = models.ArchiveUploadFailed.String()
		err = api.mongoDB.PatchInteractive(ctx, interactives.PatchAttribute(mongo.State), ix)
		if err != nil {
			log.Error(ctx, fmt.Sprintf("error updating mongo for interactive [%s], State [%s]", ix.ID, ix.State), err)
		}
		return
	}

	// Patch archive + state
	ix.Archive.Name = uri
	ix.State = models.ArchiveUploaded.String()
	err = api.mongoDB.PatchInteractive(ctx, interactives.PatchArchive, ix)
	if err != nil {
		log.Error(ctx, fmt.Sprintf("error updating mongo for interactive [%s], State [%s]", ix.ID, ix.State), err)
	}

	// Send kafka message to importer
	// CollectionID will always be there (interactive can only be uploaded inside a collection)
	err = api.producer.InteractiveUploaded(&event.InteractiveUploaded{
		ID:           ix.ID,
		FilePath:     uri,
		Title:        ix.Metadata.Title,
		CollectionID: ix.Metadata.CollectionID,
	})
	if err != nil {
		// update state in DB
		ix.State = models.ArchiveDispatchFailed.String()
		err = api.mongoDB.PatchInteractive(ctx, interactives.PatchAttribute(mongo.State), ix)
		if err != nil {
			log.Error(ctx, fmt.Sprintf("error updating mongo for interactive [%s], State [%s]", ix.ID, ix.State), err)
		}
		return
	}
	ix.State = models.ArchiveDispatchedToImporter.String()
	err = api.mongoDB.PatchInteractive(ctx, interactives.PatchAttribute(mongo.State), ix)
	if err != nil {
		log.Error(ctx, fmt.Sprintf("error updating mongo for interactive [%s], State [%s]", ix.ID, ix.State), err)
	}
}
