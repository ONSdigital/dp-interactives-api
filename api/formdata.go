package api

import (
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

	"github.com/ONSdigital/dp-interactives-api/models"
)

const (
	UpdateFieldKey      = "interactive"
	maxUploadFileSizeMb = 50
)

type FormDataValidator func(numOfAttachments int, update string) error

var (
	WantOnlyOneAttachmentWithMetadata = func(numOfAttachments int, update string) error {
		if numOfAttachments == 1 && update != "" {
			return nil
		} else {
			return errors.New("expecting one attachment with metadata")
		}
	}
	WantAtleastMaxOneAttachmentAndOrMetadata = func(numOfAttachments int, update string) error {
		if numOfAttachments == 1 || update != "" {
			return nil
		} else {
			return errors.New("no attachment (max one) or metadata present")
		}
	}
)

type FormDataRequest struct {
	req                 *http.Request
	api                 *API
	FileData            []byte
	Sha                 string
	FileName            string
	Interactive         *models.Interactive
	isMetadataMandatory bool
}

func newFormDataRequest(req *http.Request, api *API, attachmentValidator FormDataValidator, metadataMandatory bool) (*FormDataRequest, error) {
	f := &FormDataRequest{
		req:                 req,
		api:                 api,
		isMetadataMandatory: metadataMandatory,
	}
	return f, f.validate(attachmentValidator)
}

func (f *FormDataRequest) validate(attachmentValidator FormDataValidator) error {
	var data []byte
	var vErr error
	var filename string

	// 1. Expecting 1 file attachment - only zip
	vErr = f.req.ParseMultipartForm(maxUploadFileSizeMb << 20)
	if vErr != nil {
		return fmt.Errorf("parsing form data (%s)", vErr.Error())
	}
	numOfAttach := len(f.req.MultipartForm.File)
	updateModelJson := f.req.FormValue(UpdateFieldKey)
	if vErr := attachmentValidator(numOfAttach, updateModelJson); vErr != nil {
		return vErr
	}
	if updateModelJson == "" && f.isMetadataMandatory {
		return fmt.Errorf("missing mandatory (%s) key in form data", UpdateFieldKey)
	}
	if numOfAttach > 0 {
		var fileHeader *multipart.FileHeader
		var fileKey string

		for k, v := range f.req.MultipartForm.File {
			fileHeader = v[0]
			fileKey = k
		}

		file, _, vErr := f.req.FormFile(fileKey)
		if vErr != nil {
			return fmt.Errorf("error reading form data (%s)", vErr.Error())
		}
		defer file.Close()

		if ext := filepath.Ext(fileHeader.Filename); ext != ".zip" {
			return fmt.Errorf("file extension (%s) should be zip", ext)
		}
		mb := fileHeader.Size / (1 << 20)
		if mb >= maxUploadFileSizeMb {
			return fmt.Errorf("size of content (%d) MB exceeded allowed limit (%d MB)", maxUploadFileSizeMb, mb)
		}

		if data, vErr = ioutil.ReadAll(file); vErr != nil {
			return fmt.Errorf("http body read error (%s)", vErr.Error())
		}

		filename = fileHeader.Filename
	}

	// 2. Unmarshal the update field from JSON
	var interactive *models.Interactive
	if updateModelJson != "" {
		if err := json.Unmarshal([]byte(updateModelJson), &interactive); err != nil {
			return fmt.Errorf("cannot unmarshal update json %w", err)
		}
		if interactive.Metadata == nil {
			interactive.Metadata = &models.InteractiveMetadata{}
		}
		interactive.Metadata.Label = strings.TrimSpace(interactive.Metadata.Label)
	}

	hasher := sha1.New()
	hasher.Write(data)
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	f.FileData = data
	f.FileName = filename
	f.Interactive = interactive
	f.Sha = sha

	return nil
}

func (f *FormDataRequest) hasMetadata() bool {
	if f.Interactive == nil || f.Interactive.Metadata == nil {
		return false
	}
	return f.Interactive.Metadata.HasData()
}
