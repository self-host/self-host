// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

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
