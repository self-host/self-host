// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=types.cfg.yaml openapiv3.yaml
//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=server.cfg.yaml openapiv3.yaml

package malgomaj

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/self-host/self-host/api/malgomaj/library"
	ie "github.com/self-host/self-host/internal/errors"
)

type Error struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

type RestApi struct{}

func New() *RestApi {
	return &RestApi{}
}

func (ra *RestApi) CreateTask(w http.ResponseWriter, r *http.Request) {
	// We expect a NewTask object in the request body.
	var t NewTask
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	ctx := context.Background()

	cacheItem := ProgramCacheGet(t.GetId())
	if cacheItem == nil {
		// Not found; create program
		var err error
		prog, err := NewProgram(
			t.Domain,
			t.ProgramUuid,
			string(t.Language),
			time.Duration(t.Deadline)*time.Millisecond,
			t.SourceCode)
		if err != nil {
			ie.SendHTTPError(w, ie.NewInternalServerError(err))
			return
		}

		// Compile
		err = prog.Compile(ctx)
		if err != nil {
			e, ok := err.(*library.LibraryError)
			if ok {
				ie.SendHTTPError(w, &ie.HTTPError{
					Code:    500,
					Message: fmt.Sprintf("library error (%d) %v", e.Code, e.Message),
				})
				return
			}

			ie.SendHTTPError(w, ie.NewInternalServerError(err))
			return
		}

		cacheItem = ProgramCacheAdd(prog)
	}

	if cacheItem == nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	// This header can be used by an external caller to re-use this particular worker as to avoid re-compilation
	// on every call. That is, given that the cache item does not expire before the next call.
	w.Header().Set("X-Expires", cacheItem.expires.In(time.FixedZone("GMT", 0)).Format(time.RFC1123))

	if t.Http != nil {
		h := (NewTaskHttp)(*t.Http)
		ctx = context.WithValue(ctx, "http", &h)

		// Set it to something...
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		// The "CGI" program manages headers and output via w, r
		err := cacheItem.program.RunWithHTTP(ctx, w, r)
		if err != nil {
			ie.SendHTTPError(w, ie.NewInternalServerError(err))
			return
		}
	} else {
		// Run program
		err := cacheItem.program.Run(ctx)
		if err != nil {
			ie.SendHTTPError(w, ie.NewInternalServerError(err))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (ra *RestApi) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := Status{
		Load: ProgramCacheGetLoad(),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}
