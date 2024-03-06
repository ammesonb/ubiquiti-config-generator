package utils

import (
	"errors"
	"fmt"
	"strings"
)

// UbiConfGenError wraps custom errors thrown by this application
// It contains a message and a single object for context, plus a parent error if relevant
// It can unwrap the parent error and will include it in Error() string if present
type UbiConfGenError struct {
	Message   string
	Context   any
	ParentErr error
}

func (e UbiConfGenError) Error() string {
	parentStr := ""
	if e.ParentErr != nil {
		parentStr = ": " + e.ParentErr.Error()
	}
	msg := e.Message
	// Format context if present, otherwise skip it
	if strings.ContainsRune(msg, '%') {
		msg = fmt.Sprintf(msg, e.Context)
	}
	return msg + parentStr
}

func (e UbiConfGenError) Unwrap() error {
	return e.ParentErr
}

func (e UbiConfGenError) Is(target error) bool {
	return e.Matches(target)
}
func (e UbiConfGenError) As(target any) bool {
	return errors.As(e, &target)
}

func (e UbiConfGenError) Matches(err error) bool {
	// Must be the correct type
	ubiErr := UbiConfGenError{}
	if !errors.As(err, &ubiErr) {
		return false
	}

	// Errors are identical if the message and object are the same, no deep recursive check yet
	// since they frequently are different error classes, and harder to test for, so sticking to surface checks for now
	return ubiErr.Message == e.Message && ubiErr.Context == e.Context
}

// Err returns a new error with a simple string message, like errors.New()
func Err(message string) UbiConfGenError {
	return UbiConfGenError{Message: message}
}

// ErrWithCtx returns a new error with the given message and context
func ErrWithCtx(message string, ctx any) UbiConfGenError {
	return UbiConfGenError{Message: message, Context: ctx}
}

// ErrWithParent returns a new error with the given message and parent
func ErrWithParent(message string, err error) UbiConfGenError {
	return UbiConfGenError{Message: message, ParentErr: err}
}

// ErrWithCtxParent returns a new error with the given message, context, and parent
func ErrWithCtxParent(message string, context any, err error) UbiConfGenError {
	return UbiConfGenError{Message: message, Context: context, ParentErr: err}
}

// ErrWithVarCtx returns an error accepting variadic context
func ErrWithVarCtx(message string, context ...any) UbiConfGenError {
	return UbiConfGenError{Message: fmt.Sprintf(message, context...)}
}

// ErrWithVarCtxParent returns an error accepting variadic context and a parent
func ErrWithVarCtxParent(message string, err error, context ...any) UbiConfGenError {
	return UbiConfGenError{Message: fmt.Sprintf(message, context...), ParentErr: err}
}
