package api

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"

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
	err = api.mongoDB.UpsertInteractive(ctx, id, &models.Interactive{
		SHA:          retVal.Sha,
		FileName:     fileWithPath,
		MetadataJson: string(mDataJson),
		State:        models.ArchiveUploaded.String(),
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

	w.WriteHeader(http.StatusAccepted)
}

func (api *API) GetInteractiveMetadataHandler(w http.ResponseWriter, req *http.Request) {
	// get id
	ctx := req.Context()
	vars := mux.Vars(req)
	id := vars["id"]

	// fetch info from DB
	vis, err := api.mongoDB.GetInteractive(ctx, id)
	if (vis == nil && err == nil) || err == mongo.ErrNoRecordFound || (vis != nil && vis.State == models.IsDeleted.String()) {
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

func (api *API) UpdateInteractiveInfoHandler(w http.ResponseWriter, req *http.Request) {
	// called from the importer to update vis info/state
}

func (api *API) ListInteractivessHandler(w http.ResponseWriter, req *http.Request) {
	// fetches all/filtered visulatisations
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
	vis, _ := api.mongoDB.GetInteractiveFromSHA(req.Context(), sha)
	if vis != nil && vis.State != models.IsDeleted.String() {
		return nil, fmt.Errorf("archive already exists (%s)", vis.FileName)
	}

	return &validatedReq{
		Reader:   bytes.NewReader(data),
		Sha:      sha,
		FileName: fileHeader.Filename,
		Metadata: metadata}, nil
}
