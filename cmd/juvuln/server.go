// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/self-host/self-host/api/juvuln"
	"go.uber.org/zap"
	// "github.com/self-host/self-host/middleware"
)

func Server(quit <-chan struct{}, address string) (<-chan error, error) {
	api := juvuln.New()

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
		AllowedOrigins: []string{"https://*", "http://*"}, /**/
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Get("/openapi3.json", func(w http.ResponseWriter, r *http.Request) {
		f, err := juvuln.GetOpenAPIFile()
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(f)
	})

	// inlineMiddlewares := make([]juvuln.MiddlewareFunc, 0)

	// Register
	juvuln.HandlerWithOptions(api, juvuln.ChiServerOptions{
		BaseRouter: r,
		// Middlewares: inlineMiddlewares,
	})

	srv := http.Server{
		Handler: r,
		Addr:    address,
	}

	errC := make(chan error, 1)

	go func() {
		<-quit

		ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer func() {
			cancel()
			close(errC)
		}()

		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(ctxTimeout); err != nil {
			errC <- err
		}
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
