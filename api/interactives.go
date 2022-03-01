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

	"github.com/ONSdigital/dp-api-clients-go/v2/interactives"
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
	Metadata map[string]string
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
	mDataJson, _ := json.Marshal(retVal.Metadata)
	activeFlag := true
	err = api.mongoDB.UpsertInteractive(ctx, id, &models.Interactive{
		SHA:          retVal.Sha,
		MetadataJson: string(mDataJson),
		Active:       &activeFlag,
		State:        models.ArchiveUploaded.String(),
		Archive:      models.Archive{Name: fileWithPath},
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

	metadata := map[string]string{}
	json.Unmarshal([]byte(vis.MetadataJson), &metadata)
	WriteJSONBody(metadata, w, http.StatusOK)
}

func (api *API) UpdateInteractiveHandler(w http.ResponseWriter, req *http.Request) {

	// 1. Check body has some metadata and json decodes
	ctx := req.Context()
	if req.Body == nil {
		http.Error(w, "Empty body recieved", http.StatusBadRequest)
		log.Error(ctx, "Empty body recieved", ErrEmptyBody)
		return
	}
	defer req.Body.Close()
	update := interactives.InteractiveUpdate{}
	var bodyBytes, mDataJson []byte
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
	if update.ImportSuccessful == nil {
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

	// 3. Update model
	existingMap := map[string]string{}
	json.Unmarshal([]byte(existing.MetadataJson), &existingMap)
	mergedMap := mergeKeys(update.Interactive.Metadata, existingMap)

	if mDataJson, err = json.Marshal(mergedMap); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, err.Error(), err)
		return
	}
	state := models.ImportFailure
	if *update.ImportSuccessful {
		state = models.ImportSuccess
	}
	var archive models.Archive
	if update.Interactive.Archive != nil {
		archive = models.Archive{
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
		MetadataJson: string(mDataJson),
		State:        state.String(),
		Archive:      archive,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, "Unable to write to DB", err)
		return
	}

	WriteJSONBody(mergedMap, w, http.StatusOK)
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
	response := make([]*interactives.Interactive, 0)
	for _, interactive := range db {
		i, err := models.ToRest(interactive)
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

// Given two maps, recursively merge right into left, NEVER replacing any key that already exists in left
func mergeKeys(left, right map[string]string) map[string]string {
	if left == nil {
		left = make(map[string]string)
	}
	for key, rightVal := range right {
		if leftVal, present := left[key]; present {
			//then we don't want to replace it - recurse
			left[key] = leftVal
		} else {
			// key not in left so we can just shove it in
			left[key] = rightVal
		}
	}
	return left
}

func validateReq(req *http.Request, api *API) (*validatedReq, error) {
	var data []byte
	var vErr error
	var fileHeader *multipart.FileHeader
	var fileKey string
	metadata := map[string]string{}

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

	for k, v := range req.MultipartForm.Value {
		metadata[k] = v[0]
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

	// 3. Check if duplicate exists
	hasher := sha1.New()
	hasher.Write(data)
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	vis, _ := api.mongoDB.GetActiveInteractiveFromSHA(req.Context(), sha)
	if vis != nil {
		return nil, fmt.Errorf("archive already exists id (%s) with interactive (%s)", vis.ID, vis.Archive.Name)
	}

	return &validatedReq{
		Reader:   bytes.NewReader(data),
		Sha:      sha,
		FileName: fileHeader.Filename,
		Metadata: metadata}, nil
}
