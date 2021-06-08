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
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	ie "github.com/self-host/self-host/internal/errors"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type visitorController struct {
	sync.RWMutex

	visitors map[string]*visitor

	rateLimit       int
	maxBurst        int
	cleanUpInterval time.Duration
}

// Retrieve and return the rate limiter for the current visitor if it
// already exists. Otherwise create a new rate limiter and add it to
// the visitors map, using the API token as the key.
func (c *visitorController) GetVisitor(token string) (*rate.Limiter, time.Time) {
	c.RLock()
	v, exists := c.visitors[token]
	c.RUnlock()

	if !exists {
		// Should this be a viper config?
		rt := rate.Every(time.Hour / time.Duration(c.rateLimit))
		limiter := rate.NewLimiter(rt, c.maxBurst)
		lastSeen := time.Now()
		c.Lock()
		c.visitors[token] = &visitor{
			limiter:  limiter,
			lastSeen: lastSeen,
		}
		c.Unlock()
		return limiter, lastSeen
	}

	v.lastSeen = time.Now()
	return v.limiter, v.lastSeen
}

func (c *visitorController) Start() {
	go func() {
		for {
			select {
			case <-time.After(time.Minute / 10.0):
				c.Lock()
				for token, v := range c.visitors {
					if v == nil || time.Since(v.lastSeen) > c.cleanUpInterval {
						delete(c.visitors, token)
					}
				}
				c.Unlock()
			}
		}
	}()
}

func newVisitorController(r, b int, cleanUp time.Duration) *visitorController {
	return &visitorController{
		visitors:        make(map[string]*visitor),
		rateLimit:       r,
		maxBurst:        b,
		cleanUpInterval: cleanUp,
	}
}

func RateControl(reqPerHour int, maxburst int, cleanup time.Duration) func(http.HandlerFunc) http.HandlerFunc {
	// FIXME: From config, somehow.
	vc := newVisitorController(reqPerHour, maxburst, cleanup)
	vc.Start()

	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			domain, api_key, ok := r.BasicAuth()
			if ok == false {
				ie.SendHTTPError(w, ie.ErrorUnauthorized)
				return
			}

			if domain == "" || api_key == "" {
				ie.SendHTTPError(w, ie.ErrorUnauthorized)
				return
			}

			lutkey := domain + "." + api_key
			limiter, _ := vc.GetVisitor(lutkey)

			if limiter.Allow() == false {
				hourRate := limiter.Limit() * 3600

				// Number of requests per hour
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%v", hourRate))

				// FIXME: How do we represent these when we have a leaky bucket?
				// w.Header().Set("X-RateLimit-Reset", ... ))
				// w.Header().Set("X-RateLimit-Remaining", ...)

				ie.SendHTTPError(w, ie.ErrorTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
