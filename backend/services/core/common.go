package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

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

const (
	DefaultDescriptionMaxLength = 1000
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

var (
	ErrRecordNotFound     = NewSentinelServiceError(ErrorCodeRecordNotFound)
	ErrInvalidPassword    = NewSentinelServiceError(ErrorCodeInvalidPassword)
	ErrInvalidOperation   = NewSentinelServiceError(ErrorCodeInvalidOperation)
	ErrInvalidInput       = NewSentinelServiceError(ErrorCodeInvalidInput)
	ErrPermissionDenied   = NewSentinelServiceError(ErrorCodePermissionDenied)
	ErrDuplicateVariation = NewSentinelServiceError(ErrorCodeDuplicateVariation)
	ErrUnknownError       = NewSentinelServiceError(ErrorCodeUnknownError)
)

type PaginatedResult[T any] struct {
	Items      []T `json:"items" validate:"required"`
	TotalCount int `json:"totalCount" validate:"required"`
}

type ServiceVersionSpecifier struct {
	Name    string
	Version int
}

func ParseServiceVersionSpecifier(s string) (ServiceVersionSpecifier, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return ServiceVersionSpecifier{}, fmt.Errorf("invalid service version specifier: %s", s)
	}

	version, err := strconv.Atoi(parts[1])
	if err != nil {
		return ServiceVersionSpecifier{}, fmt.Errorf("invalid service version specifier: %s", s)
	}

	return ServiceVersionSpecifier{
		Name:    parts[0],
		Version: version,
	}, nil
}

func (s *ServiceVersionSpecifier) String() string {
	return fmt.Sprintf("%s:%d", s.Name, s.Version)
}
