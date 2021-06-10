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
	"context"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"

	ie "github.com/self-host/self-host/internal/errors"
)

func OapiRequestValidator(swagger *openapi3.Swagger) func(http.HandlerFunc) http.HandlerFunc {
	return OapiRequestValidatorWithOptions(swagger, nil)
}

func OapiRequestValidatorWithOptions(swagger *openapi3.Swagger, options *Options) func(http.HandlerFunc) http.HandlerFunc {
	router, err := legacyrouter.NewRouter(swagger)
	if err != nil {
		// Fatal?
		// FIXME: log error
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route, pathParams, err := router.FindRoute(r)
			if err != nil {
				ie.SendHTTPError(w, ie.NewInvalidRequestError(err))
				return
			}

			requestValidationInput := &openapi3filter.RequestValidationInput{
				Request:    r,
				PathParams: pathParams,
				Route:      route,
				Options: &openapi3filter.Options{
					AuthenticationFunc: func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
						return nil
					},
				},
			}
			if err := openapi3filter.ValidateRequest(r.Context(), requestValidationInput); err != nil {
				e, ok := err.(*openapi3filter.RequestError)
				if ok {
					nerr := &ie.HTTPError{
						Code:    400,
						Message: e.Error(),
					}
					ie.SendHTTPError(w, nerr)
				} else {
					ie.SendHTTPError(w, ie.NewInvalidRequestError(err))
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
