// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package middleware

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"

	ie "github.com/self-host/self-host/internal/errors"
	"github.com/self-host/self-host/internal/services"
	"github.com/self-host/self-host/postgres"
)

var URLParamRegex = regexp.MustCompile(`(?m)\{([^\}]+)\}`)

func PolicyValidator() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			domain, apiKey, ok := r.BasicAuth()
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

			db, err := postgres.GetDB(domain)
			if err != nil {
				ie.SendHTTPError(w, ie.NewInvalidRequestError(err))
				return
			}

			// Wire up a policy check service and check if the token has access
			check := services.NewPolicyCheckService(db)

			for _, scope := range scopes {
				// Check permission
				splitScope := strings.Split(scope, ":")
				if len(splitScope) != 2 {
					ie.SendHTTPError(w, ie.ErrorUnprocessable)
				}

				action := splitScope[0]
				resource := splitScope[1]

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

				access, err := check.UserHasAccessViaToken(ctx, []byte(apiKey), action, resource)
				if err != nil {
					ie.SendHTTPError(w, ie.ParseDBError(err))
					return
				}
				if access == false {
					ie.SendHTTPError(w, ie.ErrorForbidden)
					return
				}
			}

			var newctx context.Context
			if domain != "" && apiKey != "" {
				newctx = context.WithValue(ctx, "domaintoken", &services.DomainToken{
					Domain: domain,
					Token:  apiKey,
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
