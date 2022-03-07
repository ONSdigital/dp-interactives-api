package test_support

import (
	"bytes"
	"embed"
	"encoding/json"
	"github.com/ONSdigital/dp-interactives-api/api"
	"github.com/ONSdigital/dp-interactives-api/models"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

var (
	//go:embed resources/*
	resources embed.FS
)

// Creates a new file upload http request with optional extra params
func NewFileUploadRequest(method, uri, paramName, path string, update *models.InteractiveUpdate) *http.Request {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	if path != "-" {
		file, _ := resources.Open(path)
		fileContents, _ := ioutil.ReadAll(file)
		fi, _ := file.Stat()
		file.Close()

		part, _ := writer.CreateFormFile(paramName, fi.Name())
		part.Write(fileContents)
	}

	jsonB, _ := json.Marshal(update)
	_ = writer.WriteField(api.UpdateFieldKey, string(jsonB))

	request, _ := http.NewRequest(method, uri, body)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	writer.Close()

	return request
}
