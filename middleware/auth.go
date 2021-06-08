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
