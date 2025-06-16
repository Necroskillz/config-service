package grpc

import (
	"errors"

	"github.com/necroskillz/config-service/services/core"
	"github.com/necroskillz/config-service/util/validator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ToGRPCError(err error) error {
	if errors.Is(err, core.ErrPermissionDenied) {
		return status.Errorf(codes.PermissionDenied, "permission denied: %s", err.Error())
	}

	if errors.Is(err, core.ErrRecordNotFound) {
		return status.Errorf(codes.NotFound, "record not found: %s", err.Error())
	}

	if errors.Is(err, core.ErrInvalidOperation) {
		return status.Errorf(codes.InvalidArgument, "invalid operation: %s", err.Error())
	}

	var validationError *validator.ValidationError
	if errors.As(err, &validationError) {
		return status.Errorf(codes.InvalidArgument, "validation error: %s", validationError.Error())
	}

	if errors.Is(err, core.ErrInvalidInput) {
		return status.Errorf(codes.InvalidArgument, "invalid input: %s", err.Error())
	}

	if errors.Is(err, core.ErrUnexpectedError) {
		return status.Errorf(codes.Internal, "unexpected error: %s", err.Error())
	}

	return status.Errorf(codes.Internal, "internal server error: %s", err.Error())
}
