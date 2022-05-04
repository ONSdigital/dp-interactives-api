package models

type ErrorResponse struct {
	Errors []error `json:"errors"`
	Status int     `json:"-"`
}

func NewErrorResponse(statusCode int, errors ...error) *ErrorResponse {
	return &ErrorResponse{
		Errors: errors,
		Status: statusCode,
	}
}

type SuccessResponse struct {
	Body   []byte `json:"-"`
	Status int    `json:"-"`
}

func NewSuccessResponse(jsonBody []byte, statusCode int) *SuccessResponse {
	return &SuccessResponse{
		Body:   jsonBody,
		Status: statusCode,
	}
}
