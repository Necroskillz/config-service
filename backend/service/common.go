package service

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type ErrorCode string

const (
	ErrorCodeRecordNotFound     ErrorCode = "NOT_FOUND"
	ErrorCodeInvalidPassword    ErrorCode = "INVALID_PASSWORD"
	ErrorCodeInvalidOperation   ErrorCode = "INVALID_OPERATION"
	ErrorCodeInvalidInput       ErrorCode = "INVALID_INPUT"
	ErrorCodePermissionDenied   ErrorCode = "PERMISSION_DENIED"
	ErrorCodeDuplicateVariation ErrorCode = "DUPLICATE_VARIATION"
	ErrorCodeUnknownError       ErrorCode = "UNKNOWN_ERROR"
)

type ServiceError struct {
	Message string
	Err     error
	Code    ErrorCode
}

func NewDbError(err error, entityType string) *ServiceError {
	if errors.Is(err, pgx.ErrNoRows) {
		return NewServiceError(ErrorCodeRecordNotFound, fmt.Sprintf("%s not found", entityType))
	}

	return NewServiceError(ErrorCodeUnknownError, err.Error())
}

func NewServiceError(code ErrorCode, message string) *ServiceError {
	return &ServiceError{Message: message, Code: code}
}

func NewSentinelServiceError(code ErrorCode) *ServiceError {
	return NewServiceError(code, "")
}

func (e *ServiceError) WithErr(err error) *ServiceError {
	e.Err = err
	return e
}

func (e *ServiceError) Error() string {
	return e.Message
}

func (e *ServiceError) Is(target error) bool {
	t, ok := target.(*ServiceError)
	if !ok {
		return false
	}

	return e.Code == t.Code
}

type VersionLinkDto struct {
	ID      uint `json:"id" validate:"required"`
	Version int  `json:"version" validate:"required"`
}

var (
	ErrRecordNotFound     = NewSentinelServiceError(ErrorCodeRecordNotFound)
	ErrInvalidPassword    = NewSentinelServiceError(ErrorCodeInvalidPassword)
	ErrInvalidOperation   = NewSentinelServiceError(ErrorCodeInvalidOperation)
	ErrInvalidInput       = NewSentinelServiceError(ErrorCodeInvalidInput)
	ErrPermissionDenied   = NewSentinelServiceError(ErrorCodePermissionDenied)
	ErrDuplicateVariation = NewSentinelServiceError(ErrorCodeDuplicateVariation)
	ErrUnknownError       = NewSentinelServiceError(ErrorCodeUnknownError)
)
