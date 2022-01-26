package api

import "net/http"

func (api *API) UploadVisualisationHandler(w http.ResponseWriter, req *http.Request) {
	// get file from request
	// upload file to s3 bucket
	// sent kafka message to importer
	// expect message from importer with details (visualisation info) - commit to DB
}

func (api *API) DeleteVisualisationHandler(w http.ResponseWriter, req *http.Request) {
	// get collection ID
	// delete it from s3 and db
}

func (api *API) GetVisualisationInfoHandler(w http.ResponseWriter, req *http.Request) {
	// get id
	// fetch info from DB
}
