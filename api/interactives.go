package api

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/log.go/v2/log"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func (api *API) UploadVisualisationHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	fileName := "NS_UPLOAD_TEST.zip"
	var data []byte
	var err error

	if data, err = ioutil.ReadAll(req.Body); err != nil {
		http.Error(w, "No attachment found or error reading", http.StatusBadRequest)
		log.Error(ctx, "Invalid payload", errors.New("no attachment found or error reading"))
		return
	}

	err = api.s3.ValidateBucket()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, "Invalid s3 bucket", err)
		return
	}

	// upload file to s3 bucket
	_, err = api.s3.Upload(&s3manager.UploadInput{Body: bytes.NewReader(data), Key: &fileName})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(ctx, "S3 upload error", err)
		return
	}

	// sent kafka message to importer
}

func (api *API) DeleteVisualisationHandler(w http.ResponseWriter, req *http.Request) {
	// get collection ID
	// delete it from s3 and db
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
