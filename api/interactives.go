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
)

func (api *API) UploadInteractivesHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	// 1. Validate request
	formDataRequest, err := newFormDataRequest(req, api, WantOnlyOneAttachmentWithMetadata, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Error(ctx, "http request validation failed", err)
		return
	}
	update := formDataRequest.Update.Interactive
	if len(update.Metadata.Label) == 0 || len(update.Metadata.InternalID) == 0 || len(update.Metadata.Title) == 0 {
		err = fmt.Errorf("label (%s) title (%s) internal_id (%s) are mandatory", update.Metadata.Label, update.Metadata.Title, update.Metadata.InternalID)
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Error(ctx, err.Error(), err)
		return
	}

	if api.validateSha {
		// 2. Check if duplicate exists
		existing, _ := api.mongoDB.GetActiveInteractiveGivenSha(ctx, formDataRequest.Sha)
		if existing != nil {
			err = fmt.Errorf("archive already exists id (%s) with sha (%s)", existing.ID, existing.SHA)
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Error(ctx, "archive with sha already exists", err)
			return
		}
	}

	// 3. Check "label + title are unique"
	existing, _ := api.mongoDB.GetActiveInteractiveGivenField(ctx, "metadata.label", update.Metadata.Label)
	if existing != nil {
		err = fmt.Errorf("archive with label (%s) already exists id (%s)", existing.Metadata.Label, existing.ID)
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Error(ctx, "archive with label already exists", err)
		return
	}
	existing, _ = api.mongoDB.GetActiveInteractiveGivenField(ctx, "metadata.title", update.Metadata.Title)
	if existing != nil {
		err = fmt.Errorf("archive with title (%s) already exists id (%s)", existing.Metadata.Title, existing.ID)
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Error(ctx, "archive with title already exists", err)
		return
	}

	// 4. upload to S3
	uri, err := api.uploadFile(formDataRequest.Sha, formDataRequest.FileName, formDataRequest.FileData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, "processing form data", err)
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(ctx, "Unable to write to DB", err)
			return
		}

		if collisions == MaxCollisions {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(ctx, "Unable to write to DB - max collisions", err)
			return
		}
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

// update rules
// if published - allow only file updates
// if unpublished - allow both file + metadata
func (api *API) UpdateInteractiveHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	// 1. Validate request
	formDataRequest, err := newFormDataRequest(req, api, WantAtleastMaxOneAttachmentAndOrMetadata, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Error(ctx, "http request validation failed", err)
		return
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

	// fail if attempting to update metadata for a published model
	if *existing.Published && formDataRequest.hasMetadata() {
		http.Error(w, ErrCantUpdateMeta.Error(), http.StatusForbidden)
		log.Error(ctx, ErrCantUpdateMeta.Error(), ErrCantUpdateMeta)
		return
	}

	// -- ALL GOOD ABOVE

	// 5. prepare updated model
	updatedModel := &models.Interactive{
		ID:        id,
		Published: existing.Published,
		State:     existing.State,
		Metadata:  existing.Metadata,
	}

	var update models.Interactive
	if formDataRequest.Update == nil { // no metada update
		update.Metadata = existing.Metadata
		update.Published = existing.Published
	} else {
		update = formDataRequest.Update.Interactive

		if update.Metadata != nil {
			updatedModel.Metadata.Update(update.Metadata, api.newSlug)
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
		if api.validateSha {
			// Check if duplicate SHA exists
			i, _ := api.mongoDB.GetActiveInteractiveGivenSha(ctx, formDataRequest.Sha)
			if i != nil {
				err = fmt.Errorf("archive already exists id (%s) with sha (%s)", i.ID, i.SHA)
				http.Error(w, err.Error(), http.StatusBadRequest)
				log.Error(ctx, "archive with sha already exists", err)
				return
			}
		}

		// Process form data (S3)
		uri, err = api.uploadFile(formDataRequest.Sha, formDataRequest.FileName, formDataRequest.FileData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(ctx, "processing form data", err)
			return
		}

		updatedModel.State = models.ArchiveUploaded.String()
		updatedModel.SHA = formDataRequest.Sha
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

	// send kafka message to importer (if file uploaded)
	if uri != "" {
		err = api.producer.InteractiveUploaded(&event.InteractiveUploaded{ID: id, FilePath: uri})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(ctx, "Unable to notify importer", err)
			return
		}
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

	// must not delete published interactives
	if *vis.Published {
		http.Error(w, ErrCantDeletePublishedIn.Error(), http.StatusForbidden)
		log.Error(ctx, ErrCantDeletePublishedIn.Error(), ErrCantDeletePublishedIn)
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
