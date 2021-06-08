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
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"

	ie "github.com/noda/selfhost/internal/errors"
	"github.com/noda/selfhost/internal/services"
	pg "github.com/noda/selfhost/postgres"
)

var URLParamRegex = regexp.MustCompile(`(?m)\{([^\}]+)\}`)

func PolicyValidator() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			domain, api_key, ok := r.BasicAuth()
			if ok == false {
				realm := "Selfhost API"
				w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
				ie.SendHTTPError(w, ie.ErrorUnauthorized)
				return

			}

			ctx := r.Context()

			scopes, ok := ctx.Value("BasicAuth.Scopes").([]string)
			if ok == false {
				ie.SendHTTPError(w, ie.ErrorForbidden)
				return
			}

			db, err := pg.GetDB(domain)
			if err != nil {
				ie.SendHTTPError(w, ie.NewInvalidRequestError(err))
				return
			}

			// Wire up a policy check service and check if the token has access
			check := services.NewPolicyCheckService(db)

			for _, scope := range scopes {
				// Check permission
				split_scope := strings.Split(scope, ":")
				if len(split_scope) != 2 {
					ie.SendHTTPError(w, ie.ErrorUnprocessable)
				}

				action := split_scope[0]
				resource := split_scope[1]

				// Extract {<name>} parameters from the scope rule and
				// build a new resource string using the scope "template" and URL parameters
				matches := URLParamRegex.FindAllStringSubmatch(resource, -1)
				if matches != nil {
					// Rewrite resource
					for _, match := range matches {
						if len(match) == 2 {
							resource = strings.Replace(resource, match[0], chi.URLParam(r, match[1]), 1)
						}
					}
				}

				access, err := check.UserHasAccessViaToken(ctx, []byte(api_key), action, resource)
				if err != nil {
					ie.SendHTTPError(w, ie.ErrorUnprocessable)
					return
				}
				if access == false {
					ie.SendHTTPError(w, ie.ErrorForbidden)
					return
				}
			}

			var newctx context.Context
			if domain != "" && api_key != "" {
				newctx = context.WithValue(ctx, "domaintoken", &services.DomainToken{
					Domain: domain,
					Token:  api_key,
				})
			} else {
				newctx = ctx
			}

			// Store the DB handle in the Context
			newctx = context.WithValue(newctx, "db", db)

			next.ServeHTTP(w, r.WithContext(newctx))
		})
	}
}
