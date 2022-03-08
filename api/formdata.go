package api

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/dp-interactives-api/models"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

const (
	UpdateFieldKey      = "update"
	maxUploadFileSizeMb = 50
)

type AttachmentValidator func(numOfAttachments int) bool

var (
	WantOnlyOneAttachment = func(numOfAttachments int) bool {
		return numOfAttachments == 1
	}
	WantMaxOneAttachment = func(numOfAttachments int) bool {
		return numOfAttachments <= 1
	}
)

type FormDataRequest struct {
	req      *http.Request
	api      *API
	FileData []byte
	Sha      string
	FileName string
	Update   *models.InteractiveUpdate
}

func newFormDataRequest(req *http.Request, api *API, attachmentValidator AttachmentValidator) (*FormDataRequest, error) {
	f := &FormDataRequest{
		req: req,
		api: api,
	}
	return f, f.validate(attachmentValidator)
}

func (f *FormDataRequest) validate(attachmentValidator AttachmentValidator) error {
	var data []byte
	var vErr error
	var filename string

	// 1. Expecting 1 file attachment - only zip
	vErr = f.req.ParseMultipartForm(maxUploadFileSizeMb << 20)
	if vErr != nil {
		return fmt.Errorf("parsing form data (%s)", vErr.Error())
	}
	numOfAttach := len(f.req.MultipartForm.File)
	if !attachmentValidator(numOfAttach) {
		return fmt.Errorf("attachment validation, not expecting (%d) attachments", numOfAttach)
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
	updateModelJson := f.req.FormValue(UpdateFieldKey)
	update := &models.InteractiveUpdate{}
	if vErr = json.Unmarshal([]byte(updateModelJson), update); vErr != nil {
		return fmt.Errorf("cannot unmarshal update json %w", vErr)
	}
	if update.Interactive.Metadata == nil {
		update.Interactive.Metadata = &models.InteractiveMetadata{}
	}
	update.Interactive.Metadata.Title = strings.TrimSpace(update.Interactive.Metadata.Title)

	hasher := sha1.New()
	hasher.Write(f.FileData)
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	f.FileData = data
	f.FileName = filename
	f.Update = update
	f.Sha = sha

	return nil
}
