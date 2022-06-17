package api

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/mold/v4/modifiers"
	"github.com/go-playground/validator/v10"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ONSdigital/dp-interactives-api/models"
)

const (
	UpdateFieldKey      = "interactive"
	FileFieldKey        = "file"
	maxUploadFileSizeMb = 2500
)

type FormDataValidator func(*http.Request) error

var (
	v                                 = validator.New()
	conform                           = modifiers.New()
	WantOnlyOneAttachmentWithMetadata = func(r *http.Request) error {
		numOfAttachments, update := len(r.MultipartForm.File), r.FormValue(UpdateFieldKey)
		if numOfAttachments == 1 && update != "" {
			return nil
		}
		return errors.New("expecting one attachment with metadata")
	}
	WantAtleastMaxOneAttachmentAndOrMetadata = func(r *http.Request) error {
		numOfAttachments, update := len(r.MultipartForm.File), r.FormValue(UpdateFieldKey)
		if numOfAttachments == 1 || update != "" {
			return nil
		}
		return errors.New("no attachment (max one) or metadata present")
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

func newFormDataRequest(req *http.Request, api *API, attachmentValidator FormDataValidator, metadataMandatory bool) (*FormDataRequest, []error) {
	f := &FormDataRequest{
		req:                 req,
		api:                 api,
		isMetadataMandatory: metadataMandatory,
	}
	return f, f.validate(attachmentValidator)
}

func (f *FormDataRequest) validate(attachmentValidator FormDataValidator) (errs []error) {
	var data []byte
	var err error
	var filename string

	// Expecting 1 file attachment - only zip
	if err = f.req.ParseMultipartForm(maxUploadFileSizeMb << 20); err != nil {
		msg := fmt.Sprintf("error parsing form data %s", err.Error())
		errs = append(errs, validatorError(FileFieldKey, msg))
	}

	if f.req.MultipartForm != nil {
		if numOfAttach := len(f.req.MultipartForm.File); numOfAttach > 0 {
			var fileHeader *multipart.FileHeader
			var fileKey string

			for k, v := range f.req.MultipartForm.File {
				fileHeader = v[0]
				fileKey = k
			}

			file, _, vErr := f.req.FormFile(fileKey)
			if vErr != nil {
				msg := fmt.Sprintf("error reading form data %s", err.Error())
				errs = append(errs, validatorError(FileFieldKey, msg))
			}
			defer file.Close()

			if ext := filepath.Ext(fileHeader.Filename); ext != ".zip" {
				msg := fmt.Sprintf("file extension (%s) should be zip", ext)
				errs = append(errs, validatorError(FileFieldKey, msg))
			}

			if mb := fileHeader.Size / (1 << 20); mb >= maxUploadFileSizeMb {
				msg := fmt.Sprintf("size of content (%d) MB exceeded allowed limit (%d MB)", maxUploadFileSizeMb, mb)
				errs = append(errs, validatorError(FileFieldKey, msg))
			}

			if data, err = ioutil.ReadAll(file); err != nil {
				msg := fmt.Sprintf("http body read error %s", err.Error())
				errs = append(errs, validatorError(FileFieldKey, msg))
			}

			filename = fileHeader.Filename
		}

		if err = attachmentValidator(f.req); err != nil {
			errs = append(errs, validatorError(FileFieldKey, err.Error()))
		}
	}

	// Unmarshal the update field from JSON
	updateModelJson := f.req.FormValue(UpdateFieldKey)
	if updateModelJson == "" && f.isMetadataMandatory {
		errs = append(errs, validatorError(UpdateFieldKey, "missing mandatory key in form data"))
	}

	var interactive *models.Interactive
	if updateModelJson != "" {
		if err = json.Unmarshal([]byte(updateModelJson), &interactive); err != nil {
			errs = append(errs, validatorError(UpdateFieldKey, "cannot unmarshal update json"))
		}

		if interactive.Metadata == nil {
			interactive.Metadata = &models.InteractiveMetadata{}
		}
		interactive.Metadata.Label = strings.TrimSpace(interactive.Metadata.Label)

		if err = conform.Struct(f.req.Context(), interactive); err != nil {
			return []error{err}
		}

		if err = v.Struct(interactive); err != nil {
			if validationErrs, ok := err.(validator.ValidationErrors); ok {
				for _, vErr := range validationErrs {
					errs = append(errs, validatorError(strings.ToLower(vErr.Namespace()), vErr.Tag()))
				}
				return
			} else {
				errs = append(errs, validatorError(UpdateFieldKey, err.Error()))
			}
		}
	}

	if len(errs) == 0 {
		hasher := sha1.New()
		hasher.Write(data)
		sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

		f.FileData = data
		f.FileName = filename
		f.Interactive = interactive
		f.Sha = sha
	}

	return
}

func (f *FormDataRequest) hasMetadata() bool {
	if f.Interactive == nil || f.Interactive.Metadata == nil {
		return false
	}
	return f.Interactive.Metadata.HasData()
}

func validatorError(ns, msg string) error {
	return fmt.Errorf("%s: %s", ns, msg)
}
