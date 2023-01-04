package errors

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
)

// The maximum number of stackframes on any error.
var MaxStackDepth = 50

// causer allows for interoperability with pkg/errors
type causer interface {
	Cause() error
}

// Error is an error with an attached stacktrace. It can be used
// wherever the builtin error interface is expected.
type Error struct {
	Err    error
	stacks []uintptr
	frames []StackFrame
	prefix string
}

// new makes an Error from the given value. If that value is already an
// error then it will be used directly, if not, it will be passed to
// fmt.Errorf("%v"). The stacktrace will point to the line of code that
// called new.
func new(e interface{}) *Error {
	var err error

	switch e := e.(type) {
	case error:
		err = e
	default:
		err = fmt.Errorf("%v", e)
	}

	stack := make([]uintptr, MaxStackDepth)
	length := runtime.Callers(2, stack[:])
	return &Error{
		Err:    err,
		stacks: stack[:length],
	}
}

// wrap makes an Error from the given value. If that value is already an
// error then it will be used directly, if not, it will be passed to
// fmt.Errorf("%v"). The skip parameter indicates how far up the stacks
// to start the stacktrace. 0 is from the current call, 1 from its caller, etc.
func wrap(e interface{}, skip int) *Error {
	if e == nil {
		return nil
	}

	var err error

	switch e := e.(type) {
	case *Error:
		return e
	case error:
		err = e
	default:
		err = fmt.Errorf("%v", e)
	}

	stack := make([]uintptr, MaxStackDepth)
	length := runtime.Callers(2+skip, stack[:])
	return &Error{
		Err:    err,
		stacks: stack[:length],
	}
}

// wrapSkip calls errors.wrap, in case you want to skip a different amount
func wrapSkip(err interface{}, skip int) *Error {
	if err == nil {
		return nil
	}

	if _, ok := err.(causer); ok {
		err = fmt.Errorf("%+v", err)
	}

	return wrap(err, skip+1)
}

// wrapPrefix makes an Error from the given value. If that value is already an
// error then it will be used directly, if not, it will be passed to
// fmt.Errorf("%v"). The prefix parameter is used to add a prefix to the
// error message when calling Error(). The skip parameter indicates how far
// up the stacks to start the stacktrace. 0 is from the current call,
// 1 from its caller, etc.
func wrapPrefix(e interface{}, prefix string, skip int) *Error {
	if e == nil {
		return nil
	}

	err := wrap(e, 1+skip)

	if err.prefix != "" {
		prefix = fmt.Sprintf("%s: %s", prefix, err.prefix)
	}

	return &Error{
		Err:    err.Err,
		stacks: err.stacks,
		prefix: prefix,
	}

}

// errorf creates a new error with the given message. You can use it
// as a drop-in replacement for fmt.Errorf() to provide descriptive
// errors in return values.
func errorf(format string, a ...interface{}) *Error {
	return wrap(fmt.Errorf(format, a...), 1)
}

// Error returns the underlying error's message.
func (err *Error) Error() string {

	msg := err.Err.Error()
	if err.prefix != "" {
		msg = fmt.Sprintf("%s: %s", err.prefix, msg)
	}

	return msg
}

// Stack returns the callstack formatted the same way that go does
// in runtime/debug.Stack()
func (err *Error) Stack() []byte {
	buf := bytes.Buffer{}

	for _, frame := range err.StackFrames() {
		buf.WriteString(frame.String())
	}

	return buf.Bytes()
}

// Callers satisfies the bugsnag ErrorWithCallerS() interface
// so that the stacks can be read out.
func (err *Error) Callers() []uintptr {
	return err.stacks
}

// ErrorStack returns a string that contains both the
// error message and the callstack.
func (err *Error) ErrorStack() string {
	return err.TypeName() + " " + err.Error() + "\n" + string(err.Stack())
}

// StackFrames returns an array of frames containing information about the
// stacks.
func (err *Error) StackFrames() []StackFrame {
	if err.frames == nil {
		err.frames = make([]StackFrame, len(err.stacks))

		for i, pc := range err.stacks {
			err.frames[i] = NewStackFrame(pc)
		}
	}

	return err.frames
}

// TypeName returns the type this error. e.g. *errors.stringError.
func (err *Error) TypeName() string {
	if _, ok := err.Err.(uncaughtPanic); ok {
		return "panic"
	}
	return reflect.TypeOf(err.Err).String()
}

// Return the wrapped error (implements api for As function).
func (err *Error) unwrap() error {
	return err.Err
}
