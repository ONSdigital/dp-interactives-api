package api

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/monoculum/formam/v3"

	"github.com/ONSdigital/dp-interactives-api/event"
	"github.com/ONSdigital/dp-interactives-api/models"
	"github.com/ONSdigital/dp-interactives-api/mongo"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
)

const (
	maxUploadFileSizeMb = 50
)

var (
	ErrEmptyBody   = errors.New("empty request body")
	ErrInvalidBody = errors.New("body has invalid format")
	ErrNoMetadata  = errors.New("no metadata specified")
)

type validatedReq struct {
	Reader   *bytes.Reader
	Sha      string
	FileName string
	Metadata *interactives.InteractiveMetadata
}

var NewID = func() string {
	return uuid.NewV4().String()
}

func (api *API) UploadInteractivesHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	var err error
	var retVal *validatedReq

	// 1. Validate request
	if retVal, err = validateReq(req, api); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Error(ctx, "http response validation failed", err)
		return
	}

	err = api.s3.ValidateBucket()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, "Invalid s3 bucket", err)
		return
	}

	// 2. upload file to s3 bucket
	fileWithPath := retVal.Sha + "/" + retVal.FileName
	_, err = api.s3.Upload(&s3manager.UploadInput{Body: retVal.Reader, Key: &fileWithPath})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, "S3 upload error", err)
		return
	}

	// 3. Write to DB
	id := NewID()
	activeFlag := true
	err = api.mongoDB.UpsertInteractive(ctx, id, &models.Interactive{
		SHA:          retVal.Sha,
		Metadata: 	  retVal.Metadata,
		Active:       &activeFlag,
		State:        models.ArchiveUploaded.String(),
		Archive:      &models.Archive{Name: fileWithPath},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, "Unable to write to DB", err)
		return
	}

	// 4. send kafka message to importer
	err = api.producer.InteractiveUploaded(&event.InteractiveUploaded{ID: id, FilePath: fileWithPath})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, "Unable to notify importer", err)
		return
	}

	WriteJSONBody(id, w, http.StatusAccepted)
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

	WriteJSONBody(*vis.Metadata, w, http.StatusOK)
}

func (api *API) UpdateInteractiveHandler(w http.ResponseWriter, req *http.Request) {

	// 1. Check body json decodes
	ctx := req.Context()
	if req.Body == nil {
		http.Error(w, "Empty body recieved", http.StatusBadRequest)
		log.Error(ctx, "Empty body recieved", ErrEmptyBody)
		return
	}
	defer req.Body.Close()
	update := models.InteractiveUpdate{}
	var bodyBytes []byte
	var err error
	if bodyBytes, err = ioutil.ReadAll(req.Body); err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		log.Error(ctx, "Error reading body", ErrInvalidBody)
		return
	}

	if err := json.Unmarshal(bodyBytes, &update); err != nil {
		http.Error(w, "Error reading body (unmarshal)", http.StatusBadRequest)
		log.Error(ctx, "Error reading body (unmarshal)", ErrInvalidBody)
		return
	}
	if update.ImportSuccessful == nil || update.Interactive.Metadata == nil {
		http.Error(w, "Nothing to update", http.StatusBadRequest)
		log.Error(ctx, "Nothng to update", ErrNoMetadata)
		return
	}

	// 2. Check that id exists and is not deleted
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

	state := models.ImportFailure
	if *(update.ImportSuccessful) {
		state = models.ImportSuccess
	}
	// dont update title (is the primary key)
	update.Interactive.Metadata.Title = existing.Metadata.Title

	var archive *models.Archive
	if update.Interactive.Archive != nil {
		archive = &models.Archive{
			Name: update.Interactive.Archive.Name,
			Size: update.Interactive.Archive.Size,
		}
		for _, f := range update.Interactive.Archive.Files {
			archive.Files = append(archive.Files, &models.File{
				Name:     f.Name,
				Mimetype: f.Mimetype,
				Size:     f.Size,
			})
		}
	}

	// 4. write to DB
	err = api.mongoDB.UpsertInteractive(ctx, id, &models.Interactive{
		Metadata: update.Interactive.Metadata,
		State:    state.String(),
		Archive:  archive,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, "Unable to write to DB", err)
		return
	}

	WriteJSONBody(update.Interactive.Metadata, w, http.StatusOK)

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

func validateReq(req *http.Request, api *API) (*validatedReq, error) {
	var data []byte
	var vErr error
	var fileHeader *multipart.FileHeader
	var fileKey string
	metadata := &interactives.InteractiveMetadata{}

	// 1. Expecting 1 file attachment and some metadata
	vErr = req.ParseMultipartForm(50 << 20)
	if vErr != nil {
		return nil, fmt.Errorf("parsing form data (%s)", vErr.Error())
	}
	if numOfAttach := len(req.MultipartForm.File); numOfAttach != 1 {
		return nil, fmt.Errorf("expecting only 1 attachment, not (%d)", numOfAttach)
	}
	if numOfMetadata := len(req.MultipartForm.Value); numOfMetadata == 0 {
		return nil, fmt.Errorf("expecting some metadata")
	}

	dec := formam.NewDecoder(&formam.DecoderOptions{TagName: "bson"})
	// special handler for date-time
	dec.RegisterCustomType(func(vals []string) (interface{}, error) {
		return time.Parse("2006-01-02T15:04:05Z07:00", vals[0])
	}, []interface{}{time.Time{}}, []interface{}{&time.Time{}})
	vErr = dec.Decode(req.Form, metadata)
	if vErr != nil {
		return nil, fmt.Errorf("parsing form data (%s)", vErr.Error())
	}

	for k, v := range req.MultipartForm.File {
		fileHeader = v[0]
		fileKey = k
	}

	file, _, vErr := req.FormFile(fileKey)
	if vErr != nil {
		return nil, fmt.Errorf("error reading form data (%s)", vErr.Error())
	}
	defer file.Close()

	// 2. Expecting a zip file
	if ext := filepath.Ext(fileHeader.Filename); ext != ".zip" {
		return nil, fmt.Errorf("file extention (%s) should be zip", ext)
	}
	mb := fileHeader.Size / (1 << 20)
	if mb >= maxUploadFileSizeMb {
		return nil, fmt.Errorf("size of content (%d) MB exceeded allowed limit (%d MB)", maxUploadFileSizeMb, mb)
	}

	if data, vErr = ioutil.ReadAll(file); vErr != nil {
		return nil, fmt.Errorf("http body read error (%s)", vErr.Error())
	}

	// title must be non empty
	metadata.Title = strings.TrimSpace(metadata.Title)
	if len(metadata.Title) == 0 {
		return nil, fmt.Errorf("title must be non empty")
	}

	// 3. Check if duplicate exists
	hasher := sha1.New()
	hasher.Write(data)
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	vis, _ := api.mongoDB.GetActiveInteractiveGivenSha(req.Context(), sha)
	if vis != nil {
		return nil, fmt.Errorf("archive already exists id (%s) with sha (%s)", vis.ID, vis.SHA)
	}

	// 4. Check "title is unique"
	vis, _ = api.mongoDB.GetActiveInteractiveGivenTitle(req.Context(), metadata.Title)
	if vis != nil {
		return nil, fmt.Errorf("archive with title (%s) already exists (%s)", vis.Metadata.Title, vis.ID)
	}

	return &validatedReq{
		Reader:   bytes.NewReader(data),
		Sha:      sha,
		FileName: fileHeader.Filename,
		Metadata: metadata}, nil
}
