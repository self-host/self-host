// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package middleware

import (
	"fmt"
	"net/http"

	ie "github.com/self-host/self-host/internal/errors"
)

type BasicAuthItem struct {
	User     string
	Password string
}

func BasicAuth(auths []BasicAuthItem) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		fn := func(w http.ResponseWriter, r *http.Request) {
			u, p, ok := r.BasicAuth()
			if ok {
				for _, item := range auths {
					if item.User == u && item.Password == p {
						next.ServeHTTP(w, r)
						return
					}
				}
			}
			realm := "Selfhost"
			w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
			ie.SendHTTPError(w, ie.ErrorUnauthorized)
		}
		return http.HandlerFunc(fn)
	}
}
