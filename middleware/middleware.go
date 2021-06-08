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
package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3filter"
)

type Error struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

func sendError(w http.ResponseWriter, code int, message string) {
	reqErr := Error{
		Code:    int32(code),
		Message: message,
	}
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(reqErr)
}

type MiddlewareFunc func(http.HandlerFunc) http.HandlerFunc

type Options struct {
	Options      openapi3filter.Options
	ParamDecoder openapi3filter.ContentParameterDecoder
	UserData     interface{}
}
