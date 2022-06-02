package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/dp-interactives-api/event"
	"github.com/ONSdigital/dp-interactives-api/models"
	"github.com/ONSdigital/dp-interactives-api/mongo"
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
	// 1. Validate request
	ctx := r.Context()
	log.Info(ctx, "upload interactives")
	formDataRequest, errs := newFormDataRequest(r, api, WantOnlyOneAttachmentWithMetadata, true)
	if errs != nil {
		api.respond.Errors(ctx, w, http.StatusBadRequest, errs)
		return
	}

	update := formDataRequest.Interactive

	// 4. upload to S3
	uri, err := api.uploadFile(formDataRequest.Sha, formDataRequest.FileName, formDataRequest.FileData)
	if err != nil {
		api.respond.Error(ctx, w, http.StatusInternalServerError, fmt.Errorf("unable to upload %w", err))
		return
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

	// 5. send kafka message to importer
	err = api.producer.InteractiveUploaded(&event.InteractiveUploaded{
		ID:           id,
		FilePath:     uri,
		Title:        interactive.Metadata.Title,
		CurrentFiles: []string{""}, //need to send an empty val :(
	})
	if err != nil {
		api.respond.Error(ctx, w, http.StatusInternalServerError, fmt.Errorf("unable to notify importer %w", err))
		return
	}

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
							api.respond.Error(ctx, w, http.StatusInternalServerError, fmt.Errorf("error setting collectionID %s %s %w", id, colID, err))
							return
						}
					}
					updatedModel.Metadata.CollectionID = colID
				}
			}
		}
	}

	// Finally check if file to be uploaded
	uri := ""
	if formDataRequest.FileData != nil {
		// Process form data (S3)
		uri, err = api.uploadFile(formDataRequest.Sha, formDataRequest.FileName, formDataRequest.FileData)
		if err != nil {
			api.respond.Error(ctx, w, http.StatusInternalServerError, fmt.Errorf("unable to upload %w", err))
			return
		}

		updatedModel.State = models.ArchiveUploaded.String()
		updatedModel.SHA = formDataRequest.Sha
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
			api.respond.Error(ctx, w, http.StatusInternalServerError, fmt.Errorf("unable to notify importer %w", err))
			return
		}
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

	ix, err := api.mongoDB.ListInteractives(ctx, &models.InteractiveFilter{Metadata: &models.InteractiveMetadata{CollectionID: collectionId}})
	if err != nil {
		api.respond.Error(ctx, w, http.StatusInternalServerError, err)
		return
	}
	if len(ix) <= 0 {
		api.respond.Error(ctx, w, http.StatusNotFound, fmt.Errorf("no interactives linked to %s", collectionId))
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

		if err := api.mongoDB.PatchInteractive(ctx, mongo.Publish, inter); err != nil {
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

	var patchReq models.PatchRequest
	if err := json.Unmarshal(bytes, &patchReq); err != nil {
		api.respond.Error(ctx, w, http.StatusBadRequest, fmt.Errorf("cannot unmarshal request body %w", err))
		return
	}

	var patchAttribute mongo.PatchAttribute
	switch patchReq.Attribute {
	case "Archive":
		patchAttribute = mongo.Archive
		if patchReq.Interactive.Archive == nil {
			api.respond.Error(ctx, w, http.StatusBadRequest, fmt.Errorf("no archive to patch"))
		}

		i.State = models.ImportFailure.String()
		if patchReq.Interactive.Archive.ImportSuccessful {
			i.State = models.ImportSuccess.String()
		}

		i.Archive = &models.Archive{
			Name:          patchReq.Interactive.Archive.Name,
			Size:          patchReq.Interactive.Archive.Size,
			ImportMessage: patchReq.Interactive.Archive.ImportMessage,
		}
		for _, f := range patchReq.Interactive.Archive.Files {
			i.Archive.Files = append(i.Archive.Files, &models.File{
				Name:     f.Name,
				Mimetype: f.Mimetype,
				Size:     f.Size,
				URI:      f.URI,
			})
		}
	case "LinkToCollection":
		patchAttribute = mongo.LinkToCollection
		if patchReq.Interactive.Metadata != nil && patchReq.Interactive.Metadata.CollectionID != "" {
			i.Metadata.CollectionID = patchReq.Interactive.Metadata.CollectionID
		}
	default:
		api.respond.Error(ctx, w, http.StatusBadRequest, fmt.Errorf("unsuppported attribute %s", patchReq.Attribute))
		return
	}

	err = api.mongoDB.PatchInteractive(ctx, patchAttribute, i)
	if err != nil {
		api.respond.Error(ctx, w, http.StatusInternalServerError, fmt.Errorf("error patching interactive %s %w", i.ID, err))
		return
	}

	api.GetInteractiveHandler(w, r)
}

func (api *API) ListInteractivesHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	var filter *models.InteractiveFilter

	filterJson := req.URL.Query().Get("filter")
	log.Info(ctx, "list interactives", log.Data{"filter": filterJson})
	if filterJson != "" {
		defer req.Body.Close()
		filter = &models.InteractiveFilter{}

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

	api.respond.JSON(ctx, w, http.StatusOK, nil)
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
