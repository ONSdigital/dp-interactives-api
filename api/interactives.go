package api

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/dp-interactives-api/models"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	uuid "github.com/satori/go.uuid"
)

const (
	maxUploadFileSizeMb = 50
)

var NewID = func() string {
	return uuid.NewV4().String()
}

func (api *API) UploadVisualisationHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	var err error
	var sha string
	var reader *bytes.Reader
	defer req.Body.Close()

	if reader, sha, err = validateReqBody(req, api); err != nil {
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

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, "Invalid s3 bucket", err)
		return
	}

	// upload file to s3 bucket
	fileName := sha + "/" + "NAME_FROM_METADATA.zip"
	_, err = api.s3.Upload(&s3manager.UploadInput{Body: reader, Key: &fileName})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, "S3 upload error", err)
		return
	}

	id := NewID()
	err = api.mongoDB.UpsertVisualisation(ctx, id, &models.Visualisation{
		SHA:      sha,
		FileName: fileName,
		State:    models.ArchiveUploaded,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, "Unable to write to DB", err)
		return
	}

	// send kafka message to importer
}

func (api *API) GetVisualisationInfoHandler(w http.ResponseWriter, req *http.Request) {
	// get id
	// fetch info from DB
}

func (api *API) UpdateVisualisationInfoHandler(w http.ResponseWriter, req *http.Request) {
	// called from the importer to update vis info/state
}

func (api *API) ListVisualisationsHandler(w http.ResponseWriter, req *http.Request) {
	// fetches all/filtered visulatisations
}

func validateReqBody(req *http.Request, api *API) (*bytes.Reader, string, error) {
	var data []byte
	var vErr error
	var sha = ""

	if req.Body == nil {
		return nil, sha, ErrNoBody
	}

	if dType := req.Header.Get("Content-Type"); dType != "application/zip" {
		return nil, sha, fmt.Errorf("invalid content type %s", dType)
	}

	if data, vErr = ioutil.ReadAll(req.Body); vErr != nil {
		return nil, sha, fmt.Errorf("http body read error (%s)", vErr.Error())
	}

	mb := len(data) / (1 << 20)
	if mb >= maxUploadFileSizeMb {
		return nil, sha, fmt.Errorf("size of content (%d) MB exceeded allowed limit (%d MB)", maxUploadFileSizeMb, mb)
	}

	// Check if duplicate exists
	hasher := sha1.New()
	hasher.Write(data)
	sha = base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	vis, _ := api.mongoDB.GetVisualisationFromSHA(req.Context(), sha)
	if vis != nil {
		return nil, sha, fmt.Errorf("archive file already exists (%s)", vis.SHA+"/"+vis.FileName)
	}

	return bytes.NewReader(data), sha, nil
}
