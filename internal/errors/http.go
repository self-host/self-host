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
	"fmt"
)

// HTTPError implements ClientError interface.
type HTTPError struct {
	Cause   error  `json:"-"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e *HTTPError) Error() string {
	if e.Cause == nil {
		return e.Message
	}
	return e.Message + " : " + e.Cause.Error()
}

// ResponseBody returns JSON response body.
func (e *HTTPError) ResponseBody() ([]byte, error) {
	body, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing response body: %v", err)
	}
	return body, nil
}

// ResponseHeaders returns http status code and headers.
func (e *HTTPError) ResponseHeaders() (int, map[string]string) {
	return e.Code, map[string]string{
		"Content-Type": "application/json",
	}
}

func NewHTTPError(err error, code int, message string) error {
	return &HTTPError{
		Cause:   err,
		Message: message,
		Code:    code,
	}
}
