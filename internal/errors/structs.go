// Contains structures for project-specific error handling

package errors

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

func (e UbiConfGenError) Matches(err error) bool {
	// Must be the correct type
	var ubiErr UbiConfGenError
	if !errors.As(err, &ubiErr) {
		return false
	}

	// Errors are identical if the message and object are the same, no deep recursive check yet
	// since they frequently are different error classes and harder to test for, so sticking to surface checks for now
	return ubiErr.Message == e.Message && ubiErr.Context == e.Context
}
