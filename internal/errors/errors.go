// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package errors

import (
	"net/http"
	"strings"

	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic("zap.NewProduction " + err.Error())
	}
}

// ClientError is an error whose details to be shared with client.
type ClientError interface {
	Error() string
	// ResponseBody returns response body.
	ResponseBody() ([]byte, error)
	// ResponseHeaders returns http status code and headers.
	ResponseHeaders() (int, map[string]string)
}

var (
	ErrorDuplicateKey = &HTTPError{
		Code:    400,
		Cause:   nil,
		Message: "Request caused an error due to duplicate key violation",
	}
	ErrorMalformedRequest = &HTTPError{
		Code:    400,
		Cause:   nil,
		Message: "An error occurred due to a malformed request",
	}
	ErrorUnauthorized = &HTTPError{
		Code:    401,
		Cause:   nil,
		Message: "Unauthorized",
	}
	ErrorInvalidAPIKey = &HTTPError{
		Code:    401,
		Cause:   nil,
		Message: "The API key is invalid",
	}
	ErrorForbidden = &HTTPError{
		Code:    403,
		Cause:   nil,
		Message: "Forbidden",
	}
	ErrorNotFound = &HTTPError{
		Code:    404,
		Cause:   nil,
		Message: "The requestes resource does not exist",
	}
	ErrorUserNotFound = &HTTPError{
		Code:    404,
		Cause:   nil,
		Message: "The requestes user does not exist",
	}
	ErrorInvalidUUID = &HTTPError{
		Code:    400,
		Cause:   nil,
		Message: "Invalid UUID",
	}
	ErrorTooManyRequests = &HTTPError{
		Code:    429,
		Cause:   nil,
		Message: "Too many requests",
	}
	ErrorUnprocessable = &HTTPError{
		Code:    http.StatusUnprocessableEntity,
		Cause:   nil,
		Message: http.StatusText(http.StatusUnprocessableEntity),
	}
	ErrorInvalidUnit = &HTTPError{
		Code:    400,
		Cause:   nil,
		Message: "The provided unit can not be handled",
	}
	ErrorInvalidUnitConversion = &HTTPError{
		Code:    400,
		Cause:   nil,
		Message: "Unable to convert to the requested unit",
	}
	ErrorUndefined = &HTTPError{
		Code:    500,
		Cause:   nil,
		Message: "Undefined error",
	}
	ErrorLengthRequired = &HTTPError{
		Code:    http.StatusLengthRequired,
		Cause:   nil,
		Message: http.StatusText(http.StatusLengthRequired),
	}
	ErrorRequestEntityTooLarge = &HTTPError{
		Code:    http.StatusRequestEntityTooLarge,
		Cause:   nil,
		Message: http.StatusText(http.StatusRequestEntityTooLarge),
	}
	ErrorDBNoRows = &HTTPError{
		Code:    404,
		Cause:   nil,
		Message: "The requestes resource does not exist",
	}
	ErrorDBUndefined = &HTTPError{
		Code:    500,
		Cause:   nil,
		Message: "Undefined DB error",
	}
	ErrorDBDown = &HTTPError{
		Code:    500,
		Cause:   nil,
		Message: "The DB is currently inaccessible",
	}
)

func NewInternalServerError(err error) ClientError {
	return &HTTPError{
		Code:    500,
		Message: err.Error(),
	}
}
func NewInvalidRequestError(err error) ClientError {
	return &HTTPError{
		Code:    400,
		Message: err.Error(),
	}
}

func ParseDBError(e error) ClientError {
	if err, ok := e.(*HTTPError); ok == true {
		return err
	}

	serr := e.Error()
	if strings.Contains(serr, "SQLSTATE 23505") {
		return ErrorDuplicateKey
	} else if strings.Contains(serr, "SQLSTATE 23503") {
		return ErrorMalformedRequest
	} else if strings.Contains(serr, "no rows") {
		// Expected one row, but got no rows.
		return ErrorDBNoRows
	} else if strings.Contains(serr, "failed to connect to") || strings.Contains(serr, "unexpected EOF") {
		nerr := *ErrorDBDown
		nerr.Cause = e
		return &nerr
	}

	return ErrorDBUndefined
}

func SendHTTPError(w http.ResponseWriter, e ClientError) {
	if e == nil {
		return
	}

	if err, ok := e.(*HTTPError); ok == true {
		if err.Code >= 500 {
			m := err.Message
			if err.Cause != nil {
				m = err.Cause.Error()
			}

			logger.Error("internal", zap.Int("status", err.Code), zap.String("error", m))
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	if ce, ok := e.(*HTTPError); ok {
		w.WriteHeader(ce.Code)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Write([]byte(e.Error()))
}
