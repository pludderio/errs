package errors

import (
	"errors"
	"fmt"
)

// Err intelligently creates/handles errors, while preserving the stacks trace.
// It works with errors from github.com/pkg/errors too.
func Err(err interface{}, fmtParams ...interface{}) error {
	if err == nil {
		return nil
	}

	if _, ok := err.(causer); ok {
		err = fmt.Errorf("%+v", err)
	} else if errString, ok := err.(string); ok && len(fmtParams) > 0 {
		err = fmt.Errorf(errString, fmtParams...)
	}

	return wrap(err, 1)
}

// ErrSkip intelligently creates/handles errors, while preserving the stacks trace.
// It works with errors from github.com/pkg/errors too.
func ErrSkip(err interface{}, skip int, fmtParams ...interface{}) error {
	if err == nil {
		return nil
	}

	if _, ok := err.(causer); ok {
		err = fmt.Errorf("%+v", err)
	} else if errString, ok := err.(string); ok && len(fmtParams) > 0 {
		err = fmt.Errorf(errString, fmtParams...)
	}

	return wrapSkip(err, skip)
}

// Unwrap returns the original error that was wrapped
func Unwrap(err error) error {
	if err == nil {
		return nil
	}

	deeper := true
	for deeper {
		deeper = false
		if e, ok := err.(*Error); ok {
			err = e.Err
			deeper = true
		}
		if c, ok := err.(causer); ok {
			err = c.Cause()
			deeper = true
		}
	}

	return err
}

// Is compares two wrapped errors to determine if the underlying errors are the same
// It also interops with errors from pkg/errors
func Is(e error, original error) bool {
	if c, ok := e.(causer); ok {
		e = c.Cause()
	}
	if c, ok := original.(causer); ok {
		original = c.Cause()
	}
	return errors.Is(e, original)
}

// Prefix prefixes the message of the error with the given string
func Prefix(prefix string, err interface{}) error {
	if err == nil {
		return nil
	}
	return wrapPrefix(Err(err), prefix, 0)
}

// Trace returns the stacks trace
func Trace(err error) string {
	if err == nil {
		return ""
	}
	return string(Err(err).(*Error).Stack())
}

// FullTrace returns the error type, message, and stacks trace
func FullTrace(err error) string {
	if err == nil {
		return ""
	}
	return Err(err).(*Error).ErrorStack()
}

// Base returns a simple error with no stacks trace attached
func Base(format string, a ...interface{}) error {
	return fmt.Errorf(format, a...)
}

// HasTrace checks if error has a trace attached
func HasTrace(err error) bool {
	_, ok := err.(*Error)
	return ok
}
