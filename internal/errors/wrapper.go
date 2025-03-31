package errors

import "fmt"

// Err returns a new error with a simple string message, like errors.New()
func Err(message string) UbiConfGenError {
	return UbiConfGenError{Message: message}
}

// ErrWithCtx returns a new error with the given message and context
func ErrWithCtx(message string, ctx any) UbiConfGenError {
	return UbiConfGenError{Message: message, Context: ctx}
}

// ErrWithParent returns a new error from the parent error, with the given message
func ErrWithParent(message string, err error) UbiConfGenError {
	return UbiConfGenError{Message: message, ParentErr: err}
}

// ErrWithCtxParent returns a new error with the given message, context, and parent
func ErrWithCtxParent(message string, context any, err error) UbiConfGenError {
	return UbiConfGenError{Message: message, Context: context, ParentErr: err}
}

// ErrWithVarCtx returns an error accepting variadic context, for formatted messages
func ErrWithVarCtx(message string, context ...any) UbiConfGenError {
	return UbiConfGenError{Message: fmt.Sprintf(message, context...)}
}

// ErrWithVarCtxParent returns a formatted error from the provided parent error
func ErrWithVarCtxParent(message string, err error, context ...any) UbiConfGenError {
	return UbiConfGenError{Message: fmt.Sprintf(message, context...), ParentErr: err}
}
