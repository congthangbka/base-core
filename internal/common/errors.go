package common

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound = errors.New("not found")
	ErrInvalid  = errors.New("invalid input")
	ErrInternal = errors.New("internal error")
)

type ServiceError struct {
	Err     error
	Message string
	Code    string
}

func (e *ServiceError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "unknown error"
}

func (e *ServiceError) Unwrap() error {
	return e.Err
}

func NewServiceError(err error, message, code string) *ServiceError {
	return &ServiceError{
		Err:     err,
		Message: message,
		Code:    code,
	}
}

// WrapError wraps an error with additional context
func WrapError(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}
