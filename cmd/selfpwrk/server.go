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
package main

import (
	"context"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"

	"github.com/self-host/self-host/api/selfpwrk"
	"github.com/self-host/self-host/middleware"
)

var (
	randomUser string
	randomPass string
	r          *rand.Rand // Rand for this package.
)

func RandomString(strlen int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := range result {
		result[i] = chars[r.Intn(len(chars))]
	}
	return string(result)
}

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
	randomUser = RandomString(20)
	randomPass = RandomString(20)
}

func Server(address string) (<-chan error, error) {
	/*
		swagger, err := selfpwrk.GetSwagger()
		if err != nil {
			return nil, err
		}
	*/

	// swagger.Server = nil

	api := selfpwrk.New()

	r := chi.NewRouter()
	// Zap logging of HTTP requests
	r.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t1 := time.Now()
			defer func() {
				logger.Info(r.Method,
					zap.Duration("dur-ms", time.Since(t1)*1000),
					zap.String("url", r.URL.String()),
				)
			}()
			h.ServeHTTP(w, r)
		})
	})
	r.Use(chiware.CleanPath)
	r.Use(chiware.Timeout(60 * time.Second))

	// How do we handle this better.
	// We are going to need something in a config file...
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	/**/

	r.Get("/openapi3.json", func(w http.ResponseWriter, r *http.Request) {
		f, err := selfpwrk.GetOpenAPIFile()
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(f)
	})

	auth := []middleware.BasicAuthItem{
		{
			User:     randomUser,
			Password: randomPass,
		},
	}

	inlineMiddlewares := make([]selfpwrk.MiddlewareFunc, 0)
	inlineMiddlewares = append(inlineMiddlewares, middleware.SetHeader("Content-Type", "application/json"))
	inlineMiddlewares = append(inlineMiddlewares, middleware.BasicAuth(auth))
	// inlineMiddlewares = append(inlineMiddlewares, middleware.OapiRequestValidator(swagger))

	// Register
	selfpwrk.HandlerWithOptions(api, selfpwrk.ChiServerOptions{
		BaseRouter:  r,
		Middlewares: inlineMiddlewares,
	})

	srv := http.Server{
		Handler: r,
		Addr:    address,
	}

	errC := make(chan error, 1)

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		<-ctx.Done()

		logger.Info("Shutdown signal received")

		ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer func() {
			logger.Sync()
			stop()
			cancel()
			close(errC)
		}()

		srv.SetKeepAlivesEnabled(false)

		if err := srv.Shutdown(ctxTimeout); err != nil {
			errC <- err
		}

		logger.Info("Shutdown completed")
	}()

	go func() {
		logger.Info("Listening and serving", zap.String("address", address))

		// "ListenAndServe always returns a non-nil error. After Shutdown or Close, the returned error is
		// ErrServerClosed."
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errC <- err
		}
	}()

	return errC, nil

}
