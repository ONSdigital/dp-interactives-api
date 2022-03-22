package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/ONSdigital/dp-interactives-api/event"
	"github.com/ONSdigital/dp-interactives-api/models"
	"github.com/ONSdigital/dp-interactives-api/mongo"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

var (
	enabled, disabled = true, false
	ErrInvalidBody    = errors.New("body has invalid format")
	ErrCantUpdateSlug = errors.New("cannot update readable slug for a published interactive")
)

func (api *API) UploadInteractivesHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	// 1. Validate request
	formDataRequest, err := newFormDataRequest(req, api, WantOnlyOneAttachment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Error(ctx, "http request validation failed", err)
		return
	}
	update := formDataRequest.Update.Interactive
	if len(update.Metadata.Title) == 0 {
		err = errors.New("title must be non empty")
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Error(ctx, "title must be non empty", err)
		return
	}

	// 2. Check if duplicate exists
	existing, _ := api.mongoDB.GetActiveInteractiveGivenSha(ctx, formDataRequest.Sha)
	if existing != nil {
		err = fmt.Errorf("archive already exists id (%s) with sha (%s)", existing.ID, existing.SHA)
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Error(ctx, "archive with sha already exists", err)
		return
	}

	// 3. Check "title is unique"
	existing, _ = api.mongoDB.GetActiveInteractiveGivenTitle(ctx, update.Metadata.Title)
	if existing != nil {
		err = fmt.Errorf("archive already exists id (%s) with title (%s)", existing.ID, existing.Metadata.Title)
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Error(ctx, "archive with title already exists", err)
		return
	}

	// 4. Process form data (S3)
	uri, err := api.uploadFile(formDataRequest.Sha, formDataRequest.FileName, formDataRequest.FileData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, "processing form data", err)
		return
	}

	// 5. Write to DB
	id := api.newUUID("")
	update.Metadata.ResourceID = api.newResourceID("")
	update.Metadata.HumanReadableSlug = api.newSlug(update.Metadata.Title)

	interact := &models.Interactive{
		ID:        id,
		SHA:       formDataRequest.Sha,
		Metadata:  update.Metadata,
		Active:    &enabled,
		Published: &disabled,
		State:     models.ArchiveUploaded.String(),
		Archive:   &models.Archive{Name: uri},
	}
	err = api.mongoDB.UpsertInteractive(ctx, id, interact)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, "Unable to write to DB", err)
		return
	}

	// 5. send kafka message to importer
	err = api.producer.InteractiveUploaded(&event.InteractiveUploaded{ID: id, FilePath: uri})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, "Unable to notify importer", err)
		return
	}

	WriteJSONBody(interact, w, http.StatusAccepted)
}

func (api *API) GetInteractiveMetadataHandler(w http.ResponseWriter, req *http.Request) {
	// get id
	ctx := req.Context()
	vars := mux.Vars(req)
	id := vars["id"]

	// fetch info from DB
	vis, err := api.mongoDB.GetInteractive(ctx, id)
	if (vis == nil && err == nil) || err == mongo.ErrNoRecordFound || (vis != nil && !*vis.Active) {
		http.Error(w, fmt.Sprintf("interactive-id (%s) is either deleted or does not exist", id), http.StatusNotFound)
		log.Error(ctx, fmt.Sprintf("interactive-id (%s) is either deleted or does not exist", id), err)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, fmt.Sprintf("error fetching interactive id (%s)", id), err)
		return
	}

	WriteJSONBody(*vis, w, http.StatusOK)
}

func (api *API) UpdateInteractiveHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	// 1. Validate request
	formDataRequest, err := newFormDataRequest(req, api, WantMaxOneAttachment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Error(ctx, "http request validation failed", err)
		return
	}
	update := formDataRequest.Update.Interactive

	// 2. Upload file if requested
	if formDataRequest.FileData != nil {
		// Check if duplicate SHA exists
		i, _ := api.mongoDB.GetActiveInteractiveGivenSha(ctx, formDataRequest.Sha)
		if i != nil {
			err = fmt.Errorf("archive already exists id (%s) with sha (%s)", i.ID, i.SHA)
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Error(ctx, "archive with sha already exists", err)
			return
		}

		// Process form data (S3)
		_, err := api.uploadFile(formDataRequest.Sha, formDataRequest.FileName, formDataRequest.FileData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(ctx, "processing form data", err)
			return
		}
	}

	// 3. Check that id exists and is not deleted
	vars := mux.Vars(req)
	id := vars["id"]
	existing, err := api.mongoDB.GetInteractive(ctx, id)
	if (existing == nil && err == nil) || err == mongo.ErrNoRecordFound || (existing != nil && !*existing.Active) {
		http.Error(w, fmt.Sprintf("interactive-id (%s) is either deleted or does not exist", id), http.StatusNotFound)
		log.Error(ctx, fmt.Sprintf("interactive-id (%s) is either deleted or does not exist", id), err)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, fmt.Sprintf("error fetching interactive id (%s)", id), err)
		return
	}

	// 4. fail if attempting to update the slug for a published model
	if *existing.Published && update.Metadata != nil && existing.Metadata.HumanReadableSlug != update.Metadata.HumanReadableSlug {
		http.Error(w, ErrCantUpdateSlug.Error(), http.StatusForbidden)
		logMsg := fmt.Sprintf("attempting to update slug for a published model existing (%s), update (%s)", existing.Metadata.HumanReadableSlug, update.Metadata.HumanReadableSlug)
		log.Error(ctx, logMsg, ErrCantUpdateSlug)
		return
	}

	// 5. prepare updated model
	updatedModel := &models.Interactive{
		ID:            id,
		Published:     update.Published,
		State:         models.ImportFailure.String(),
		ImportMessage: &formDataRequest.Update.ImportMessage,
	}

	if formDataRequest.Update.ImportSuccessful != nil && *formDataRequest.Update.ImportSuccessful {
		updatedModel.State = models.ImportSuccess.String()
	}

	if update.Metadata != nil {
		updatedModel.Metadata = update.Metadata
		// dont update title (is the primary key)
		updatedModel.Metadata.Title = existing.Metadata.Title
		updatedModel.Metadata.Uri = existing.Metadata.Uri
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

	// 6. write to DB
	err = api.mongoDB.UpsertInteractive(ctx, id, updatedModel)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, "Unable to write to DB", err)
		return
	}

	// 7. get updated model
	i, err := api.mongoDB.GetInteractive(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, fmt.Sprintf("error fetching interactive id (%s)", id), err)
		return
	}

	WriteJSONBody(i, w, http.StatusOK)
}

func (api *API) ListInteractivesHandler(w http.ResponseWriter, req *http.Request, limit int, offset int) (interface{}, int, error) {
	// fetches all/filtered visulatisations
	ctx := req.Context()
	var filter *models.InteractiveMetadata
	// get an optional metadata filter
	filterJson := req.URL.Query().Get("filter")
	if filterJson != "" {
		defer req.Body.Close()
		filter = &models.InteractiveMetadata{}

		if err := json.Unmarshal([]byte(filterJson), &filter); err != nil {
			http.Error(w, "Error unmarshalling body", http.StatusBadRequest)
			log.Error(ctx, "Error unmarshalling body", ErrInvalidBody)
			return nil, 0, err
		}
	}
	db, totalCount, err := api.mongoDB.ListInteractives(ctx, offset, limit, filter)
	if err != nil {
		log.Error(ctx, "api endpoint getDatasets datastore.GetDatasets returned an error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, 0, err
	}
	response := make([]*models.Interactive, 0)
	for _, interactive := range db {
		i, err := models.Map(interactive)
		if err != nil {
			log.Error(ctx, "cannot map db to http response", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, 0, err
		}
		response = append(response, i)
	}

	return response, totalCount, nil
}

func (api *API) DeleteInteractivesHandler(w http.ResponseWriter, req *http.Request) {
	// get id
	ctx := req.Context()
	vars := mux.Vars(req)
	id := vars["id"]

	// error if it doesnt exist
	vis, err := api.mongoDB.GetInteractive(ctx, id)
	if (vis == nil && err == nil) || err == mongo.ErrNoRecordFound {
		http.Error(w, fmt.Sprintf("interactive-id (%s) does not exist", id), http.StatusNotFound)
		log.Error(ctx, fmt.Sprintf("interactive-id (%s) does not exist", id), err)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, fmt.Sprintf("error fetching interactive id (%s)", id), err)
		return
	}

	// set to inactive
	err = api.mongoDB.UpsertInteractive(ctx, id, &models.Interactive{
		Active: &disabled,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, "Unable to unset active flag", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
