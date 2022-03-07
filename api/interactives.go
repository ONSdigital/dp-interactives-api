package api

import (
	"errors"
	"fmt"
	"github.com/ONSdigital/dp-interactives-api/event"
	"github.com/ONSdigital/dp-interactives-api/models"
	"github.com/ONSdigital/dp-interactives-api/mongo"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"net/http"
)

var (
	ErrEmptyBody      = errors.New("empty request body")
	ErrInvalidBody    = errors.New("body has invalid format")
	ErrNoMetadata     = errors.New("no metadata specified")
	ErrCantUpdateSlug = errors.New("cannot update readable slug for a published interactive")

	NewID = func() string {
		return uuid.NewV4().String()
	}
)

func (api *API) UploadInteractivesHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	// 1. Validate request
	formDataRequest, err := NewFormDataRequest(req, api, WantOnlyOneAttachment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Error(ctx, "http request validation failed", err)
		return
	}
	if len(formDataRequest.Update.Interactive.Metadata.Title) == 0 {
		err = errors.New("title must be non empty")
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Error(ctx, "title must be non empty", err)
		return
	}

	// 2. Check if duplicate exists
	vis, _ := api.mongoDB.GetActiveInteractiveGivenSha(ctx, formDataRequest.Sha)
	if vis != nil {
		err = fmt.Errorf("archive already exists id (%s) with sha (%s)", vis.ID, vis.SHA)
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Error(ctx, "archive with sha already exists", err)
		return
	}

	// 3. Check "title is unique"
	vis, _ = api.mongoDB.GetActiveInteractiveGivenTitle(ctx, formDataRequest.Update.Interactive.Metadata.Title)
	if vis != nil {
		err = fmt.Errorf("archive already exists id (%s) with title (%s)", vis.ID, vis.Metadata.Title)
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Error(ctx, "archive with title already exists", err)
		return
	}

	// 4. Process form data (S3)
	err = api.uploadFile(formDataRequest.Sha, formDataRequest.FileName, formDataRequest.FileData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, "processing form data", err)
		return
	}

	// 5. Write to DB
	id := NewID()
	var activeFlag, pubFlag = true, false
	interact := &models.Interactive{
		ID:        id,
		SHA:       formDataRequest.Sha,
		Metadata:  formDataRequest.Update.Interactive.Metadata,
		Active:    &activeFlag,
		Published: &pubFlag,
		State:     models.ArchiveUploaded.String(),
		Archive:   &models.Archive{Name: formDataRequest.FileName},
	}
	err = api.mongoDB.UpsertInteractive(ctx, id, interact)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, "Unable to write to DB", err)
		return
	}

	// 5. send kafka message to importer
	err = api.producer.InteractiveUploaded(&event.InteractiveUploaded{ID: id, FilePath: formDataRequest.FileName})
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
	formDataRequest, err := NewFormDataRequest(req, api, WantMaxOneAttachment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Error(ctx, "http request validation failed", err)
		return
	}

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
		err = api.uploadFile(formDataRequest.Sha, formDataRequest.FileName, formDataRequest.FileData)
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
	if *existing.Published && formDataRequest.Update.Interactive.Metadata != nil && existing.Metadata.HumanReadableSlug != formDataRequest.Update.Interactive.Metadata.HumanReadableSlug {
		http.Error(w, ErrCantUpdateSlug.Error(), http.StatusForbidden)
		logMsg := fmt.Sprintf("attempting to update slug for a published model existing (%s), update (%s)", existing.Metadata.HumanReadableSlug, formDataRequest.Update.Interactive.Metadata.HumanReadableSlug)
		log.Error(ctx, logMsg, ErrCantUpdateSlug)
		return
	}

	// 5. prepare updated model
	updatedModel := &models.Interactive{
		ID:            id,
		Published:     formDataRequest.Update.Interactive.Published,
		State:         models.ImportFailure.String(),
		ImportMessage: &formDataRequest.Update.ImportMessage,
	}

	if formDataRequest.Update.ImportSuccessful != nil && *formDataRequest.Update.ImportSuccessful {
		updatedModel.State = models.ImportSuccess.String()
	}

	if formDataRequest.Update.Interactive.Metadata != nil {
		updatedModel.Metadata = formDataRequest.Update.Interactive.Metadata
		// dont update title (is the primary key)
		updatedModel.Metadata.Title = existing.Metadata.Title
		updatedModel.Metadata.Uri = existing.Metadata.Uri
	}

	if formDataRequest.Update.Interactive.Archive != nil {
		updatedModel.Archive = &models.Archive{
			Name: formDataRequest.Update.Interactive.Archive.Name,
			Size: formDataRequest.Update.Interactive.Archive.Size,
		}
		for _, f := range formDataRequest.Update.Interactive.Archive.Files {
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

	WriteJSONBody(updatedModel, w, http.StatusOK)
}

func (api *API) ListInteractivesHandler(w http.ResponseWriter, req *http.Request, limit int, offset int) (interface{}, int, error) {
	// fetches all/filtered visulatisations
	ctx := req.Context()
	db, totalCount, err := api.mongoDB.ListInteractives(ctx, offset, limit)
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
	activeFlag := false
	err = api.mongoDB.UpsertInteractive(ctx, id, &models.Interactive{
		Active: &activeFlag,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, "Unable to unset active flag", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
