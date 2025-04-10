package service

type ErrorCode string

const (
	ErrorCodeRecordNotFound   ErrorCode = "NOT_FOUND"
	ErrorCodeInvalidPassword  ErrorCode = "INVALID_PASSWORD"
	ErrorCodeInvalidOperation ErrorCode = "INVALID_OPERATION"
	ErrorCodePermissionDenied ErrorCode = "PERMISSION_DENIED"
)

type ServiceError struct {
	Message string
	Err     error
	Code    ErrorCode
}

func NewServiceError(code ErrorCode, message string, err error) *ServiceError {
	return &ServiceError{Message: message, Err: err, Code: code}
}

func NewSentinelServiceError(code ErrorCode) *ServiceError {
	return NewServiceError(code, "", nil)
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
	ErrRecordNotFound   = NewSentinelServiceError(ErrorCodeRecordNotFound)
	ErrInvalidPassword  = NewSentinelServiceError(ErrorCodeInvalidPassword)
	ErrInvalidOperation = NewSentinelServiceError(ErrorCodeInvalidOperation)
	ErrPermissionDenied = NewSentinelServiceError(ErrorCodePermissionDenied)
)
