// Copyright (c) 2014 Datacratic. All rights reserved.

package rest

import (
	"fmt"
)

// ErrorType is used to categories errors reported into types.
type ErrorType string

const (
	// EndpointError indicates that the remote endpoint returned an error.
	EndpointError = "endpoint-error"

	// HandlerError indicates that the route handler returned an error.
	HandlerError = "handler-error"

	// UnknownRoute indicates that no matching routes were found for the path.
	UnknownRoute = "unknown-route"

	// UnexpectedStatusCode indicates that the returned status code of an HTTP
	// request was not expected.
	UnexpectedStatusCode = "unexpected-status-code"

	// UnsupportedContentType indicates that the content-type header of an HTTP
	// request contained an unsupported value.
	UnsupportedContentType = "unsupported-content-type"

	// ReadBodyError indicates that an error occured while reading the body of
	// an HTTP request or response.
	ReadBodyError = "ready-body-error"

	// NewRequestError indicates that an error occured while creating an HTTP
	// request.
	NewRequestError = "new-request-error"

	// SendRequestError indicates that an error occured while sending an HTTP
	// request.
	SendRequestError = "send-request-error"

	// TimeoutError indicates that the request timed out while sending an HTTP
	// request.
	TimeoutError = "timeout-error"

	// UnmarshalError indicates that an error occured while deserializing the
	// body of an HTTP response.
	UnmarshalError = "unmarshal-error"

	// MarshalError indicates that an error occured while serializing the body
	// of an HTTP request.
	MarshalError = "marshal-error"
)

// Error is a typed wrapper for an error that occured while processing a REST
// request.
type Error struct {
	// Type is the type of the error.
	Type ErrorType

	// Sub is the wrapped error.
	Sub error
}

// ErrorFmt creates a new Error object with the given type.
//
// We can't call this function Errorf because of go vet who chokes if the first
// argument isn't the format string.
func ErrorFmt(errT ErrorType, format string, args ...interface{}) *Error {
	return &Error{errT, fmt.Errorf(format, args...)}
}

// Error returns the string representation of the error.
func (err *Error) Error() string {
	return fmt.Sprintf("REST error(%s): %s", err.Type, err.Sub.Error())
}

// CodedError is used to control the HTTP return code of a REST request when an
// error occurs.
//
// \todo Need to replace this by a solution that allows the modification of ANY
// part of the HTTP response.
type CodedError struct {
	Code int
	Sub  error
}

// Error returns the string representation of the error.
func (err *CodedError) Error() string {
	return fmt.Sprintf("Coded error(%d): %s", err.Code, err.Sub.Error())
}
