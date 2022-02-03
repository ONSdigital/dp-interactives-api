package api

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/ONSdigital/dp-interactives-api/event"
	"github.com/ONSdigital/dp-interactives-api/models"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	uuid "github.com/satori/go.uuid"
)

const (
	maxUploadFileSizeMb = 50
)

type validatedReq struct {
	Reader   *bytes.Reader
	Sha      string
	FileName string
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
	lala := models.ArchiveUploaded
	err = api.mongoDB.UpsertInteractive(ctx, id, &models.Interactive{
		SHA:      retVal.Sha,
		FileName: fileWithPath,
		State:    &lala,
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

func (api *API) GetInteractiveInfoHandler(w http.ResponseWriter, req *http.Request) {
	// get id
	// fetch info from DB
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

	file, fileHeader, vErr := req.FormFile("file")
	if vErr != nil {
		return nil, fmt.Errorf("error reading form data (%s)", vErr.Error())
	}
	defer file.Close()

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

	// Check if duplicate exists
	hasher := sha1.New()
	hasher.Write(data)
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	vis, _ := api.mongoDB.GetInteractiveFromSHA(req.Context(), sha)
	if vis != nil {
		return nil, fmt.Errorf("archive already exists (%s)", vis.FileName)
	}

	return &validatedReq{Reader: bytes.NewReader(data), Sha: sha, FileName: fileHeader.Filename}, nil
}
