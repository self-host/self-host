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
along with dogtag.  If not, see <http://www.gnu.org/licenses/>.
*/
//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=types.cfg.yaml openapiv3.yaml
//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=server.cfg.yaml openapiv3.yaml

package selfpwrk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/self-host/self-host/api/selfpwrk/library"
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

	cache_item := ProgramCacheGet(t.GetId())
	if cache_item == nil {
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

		cache_item = ProgramCacheAdd(prog)
	}

	if cache_item == nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	// This header can be used by an external caller to re-use this particular worker as to avoid re-compilation
	// on every call. That is, given that the cache item does not expire before the next call.
	w.Header().Set("X-Expires", cache_item.expires.In(time.FixedZone("GMT", 0)).Format(time.RFC1123))

	if t.Http != nil {
		h := (NewTaskHttp)(*t.Http)
		ctx = context.WithValue(ctx, "http", &h)

		// Set it to something...
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		// The "CGI" program manages headers and output via w, r
		err := cache_item.program.RunWithHTTP(ctx, w, r)
		if err != nil {
			ie.SendHTTPError(w, ie.NewInternalServerError(err))
			return
		}
	} else {
		// Run program
		err := cache_item.program.Run(ctx)
		if err != nil {
			ie.SendHTTPError(w, ie.NewInternalServerError(err))
			return
		}
	}
}

func (ra *RestApi) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := Status{
		Load: ProgramCacheGetLoad(),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}
