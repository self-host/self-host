/*
Copyright (C) 2021 The Self-host Authors.
This file is part of Self-host <https://github.com/self-host/self-host>.

Self-host is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Self-host is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Self-host.  If not, see <http://www.gnu.org/licenses/>.
*/
package errors

import (
	"encoding/json"
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
		if err.Code >= 500 {
			logger.Error("dberror", zap.Error(err.Cause))
		}
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
	}

	logger.Error("dberror", zap.Error(e))
	return ErrorDBUndefined
}

func SendHTTPError(w http.ResponseWriter, e ClientError) {
	if e == nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if ce, ok := e.(*HTTPError); ok {
		w.WriteHeader(ce.Code)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(e)
}
