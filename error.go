// Copyright (c) 2014 Datacratic. All rights reserved.

package rest

import (
	"fmt"
)

// ErrorType is used to categories errors reported into types.
type ErrorType int

const (
	// NoError indicates that no errors occured.
	NoError ErrorType = iota

	// EndpointError indicates that the remote endpoint returned an error.
	EndpointError

	// HandlerError indicates that the route handler returned an error.
	HandlerError

	// UnknownRoute indicates that no matching routes were found for the path.
	UnknownRoute

	// UnexpectedStatusCode indicates that the returned status code of an HTTP
	// request was not expected.
	UnexpectedStatusCode

	// UnsupportedContentType indicates that the content-type header of an HTTP
	// request contained an unsupported value.
	UnsupportedContentType

	// ReadBodyError indicates that an error occured while reading the body of
	// an HTTP request or response.
	ReadBodyError

	// NewRequestError indicates that an error occured while creating an HTTP
	// request.
	NewRequestError

	// SendRequestError indicates that an error occured while sending an HTTP
	// request.
	SendRequestError

	// UnmarshalError indicates that an error occured while deserializing the
	// body of an HTTP response.
	UnmarshalError

	// MarshalError indicates that an error occured while serializing the body
	// of an HTTP request.
	MarshalError
)

// CodedError can be used to control the HTTP status code returned by the
// Endpoint struct when an error occured.
type CodedError struct {
	// Code is the HTTP status code to return with the error.
	Code int

	// Err is the error being wrapped by the error struct.
	Err error
}

// Error returns the string representation of the error.
func (err CodedError) Error() string {
	return fmt.Sprintf("HTTP Code: %d, Error: %s", err.Code, err.Err.Error())
}

// Error constructs a CodedError used to control the HTTP status code.
func Error(code int, err error) error {
	return &CodedError{code, err}
}
