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
	ErrCantUpdateMeta        = errors.New("cannot update metadata for a published interactive")
	ErrCantDeletePublishedIn = errors.New("cannot delete a published interactive")
	ErrPubErrNoCollectionID  = errors.New("cannot publish interactive, no collection ID")
)

func (api *API) UploadInteractivesHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Validate request
	ctx := r.Context()
	formDataRequest, errs := newFormDataRequest(r, api, WantOnlyOneAttachmentWithMetadata, true)
	if errs != nil {
		api.respond.Errors(ctx, w, http.StatusBadRequest, errs)
		return
	}

	if api.cfg.ValidateSHAEnabled {
		// 2. Check if duplicate exists
		existing, _ := api.mongoDB.GetActiveInteractiveGivenSha(ctx, formDataRequest.Sha)
		if existing != nil {
			api.respond.Error(ctx, w, http.StatusBadRequest, fmt.Errorf("archive already exists id (%s) with sha (%s)", existing.ID, existing.SHA))
			return
		}
	}

	// 3. Check "label + title are unique"
	update := formDataRequest.Interactive
	existing, _ := api.mongoDB.GetActiveInteractiveGivenField(ctx, "metadata.label", update.Metadata.Label)
	if existing != nil {
		api.respond.Error(ctx, w, http.StatusBadRequest, fmt.Errorf("archive with label (%s) already exists id (%s)", existing.Metadata.Label, existing.ID))
		return
	}
	existing, _ = api.mongoDB.GetActiveInteractiveGivenField(ctx, "metadata.title", update.Metadata.Title)
	if existing != nil {
		api.respond.Error(ctx, w, http.StatusBadRequest, fmt.Errorf("archive with title (%s) already exists id (%s)", existing.Metadata.Title, existing.ID))
		return
	}

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

	// fail if attempting to update metadata for a published model
	if *existing.Published && formDataRequest.hasMetadata() {
		api.respond.Error(ctx, w, http.StatusForbidden, ErrCantUpdateMeta)
		return
	}

	// -- ALL GOOD ABOVE

	// prepare updated model
	updatedModel := &models.Interactive{
		ID:        id,
		Published: existing.Published,
		State:     existing.State,
		Archive:   existing.Archive,
		Metadata:  existing.Metadata,
	}

	update := formDataRequest.Interactive
	if update.Metadata != nil {
		updatedModel.Metadata = updatedModel.Metadata.Update(update.Metadata, api.newSlug)
	}
	if update.Published != nil {
		updatedModel.Published = update.Published
	}

	// Finally check if file to be uploaded
	uri := ""
	if formDataRequest.FileData != nil {
		if api.cfg.ValidateSHAEnabled {
			// Check if duplicate SHA exists
			i, _ := api.mongoDB.GetActiveInteractiveGivenSha(ctx, formDataRequest.Sha)
			if i != nil {
				api.respond.Error(ctx, w, http.StatusBadRequest, fmt.Errorf("archive already exists id (%s) with sha (%s)", i.ID, i.SHA))
				return
			}
		}

		// Process form data (S3)
		uri, err = api.uploadFile(formDataRequest.Sha, formDataRequest.FileName, formDataRequest.FileData)
		if err != nil {
			api.respond.Error(ctx, w, http.StatusInternalServerError, fmt.Errorf("unable to upload %w", err))
			return
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
						api.respond.Error(ctx, w, http.StatusInternalServerError, fmt.Errorf("error setting collectionID %s %s %w", id, colID, err))
						return
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
				api.respond.Error(ctx, w, http.StatusInternalServerError, fmt.Errorf("error publishing collectionID %s %s %w", id, collID, err))
				return
			}
			updatedModel.Published = &enabled
		} else {
			log.Error(ctx, fmt.Sprintf("no collection id for interactive (%s)", existing.ID), ErrPubErrNoCollectionID)
		}
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

func (api *API) PatchInteractiveHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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

	var patchAttribute mongo.PatchAttribure
	switch patchReq.Attribute {
	case "Archive":
		patchAttribute = mongo.Archive
		if patchReq.Interactive.Archive != nil {
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
				})
			}
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

func (api *API) ListInteractivesHandler(req *http.Request, limit int, offset int) (interface{}, int, error) {
	ctx := req.Context()
	var filter *models.InteractiveFilter

	filterJson := req.URL.Query().Get("filter")
	if filterJson != "" {
		defer req.Body.Close()
		filter = &models.InteractiveFilter{}

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
