package models

import (
	"context"
	"errors"
	"github.com/ONSdigital/log.go/v2/log"
)

type Error struct {
	Cause       error  `json:"-"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return e.Code + ": " + e.Description
}

func NewError(ctx context.Context, cause error, code string, description string) *Error {
	err := &Error{
		Cause:       cause,
		Code:        code,
		Description: description,
	}
	log.Error(ctx, description, err)
	return err
}

func NewValidationError(ctx context.Context, code string, description string) *Error {
	err := &Error{
		Cause:       errors.New(code),
		Code:        code,
		Description: description,
	}

	log.Error(ctx, description, err, log.Data{"code": code})
	return err
}
