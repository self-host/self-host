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
	"embed"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-chi/chi/v5"
	chiware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/self-host/self-host/api/aapije"
	"github.com/self-host/self-host/api/aapije/rest"
	"github.com/self-host/self-host/middleware"
	"github.com/self-host/self-host/pkg/util"
	"github.com/self-host/self-host/postgres"
)

var URLParamRegex = regexp.MustCompile(`(?m)\{([^\}]+)\}`)

//go:embed static
var content embed.FS

func Server(address string) (<-chan error, error) {
	swagger, err := rest.GetSwagger()
	if err != nil {
		return nil, err
	}

	swagger.Servers = nil

	domainfile := viper.GetString("domainfile")
	if domainfile != "" {
		v := viper.New()
		v.SetConfigName(domainfile)
		v.SetConfigType("yaml")
		v.AddConfigPath("/etc/selfhost/")
		v.AddConfigPath("$HOME/.config/selfhost")
		v.AddConfigPath(".")

		err := v.ReadInConfig()
		if err != nil {
			logger.Error("Error while loading domainfile", zap.Error(err))
		}

		if v.IsSet("domains") {
			for domain, pguri := range v.GetStringMapString("domains") {
				err := postgres.AddDB(domain, pguri)
				if err != nil {
					logger.Error("Error while adding domain", zap.Error(err))
				}
			}
		}

		v.WatchConfig()
		v.OnConfigChange(func(e fsnotify.Event) {
			err := v.ReadInConfig()
			if err != nil {
				logger.Error("Error while loading domainfile", zap.Error(err))
			}

			// Find inactive databases
			domains := postgres.GetDomains()
			for domain := range v.GetStringMapString("domains") {
				index := util.StringSliceIndex(domains, domain)
				if index == -1 || len(domains) == 0 {
					continue
				} else if len(domains) == 1 {
					// Absolute last element in the slice
					domains = make([]string, 0)
				} else {
					// Place last element at position
					domains[index] = domains[len(domains)-1]
					// "delete" last element
					domains[len(domains)-1] = ""
					// Truncate slice
					domains = domains[:len(domains)-1]
				}
			}

			// What remains in "domains" is all domains no longer active in config file
			for _, domain := range domains {
				postgres.RemoveDB(domain)
			}

			// Add new/existing domain DBs
			for domain, pguri := range v.GetStringMapString("domains") {
				err := postgres.AddDB(domain, pguri)
				if err != nil {
					logger.Error("Error while adding domain", zap.Error(err))
				}
			}
		})
	}

	restApi := aapije.NewRestApi()

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
	r.Use(chiware.Heartbeat("/status"))
	r.Use(chiware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   viper.GetStringSlice("cors.allowed_origins"),
		AllowedMethods:   viper.GetStringSlice("cors.allowed_methods"),
		AllowedHeaders:   viper.GetStringSlice("cors.allowed_headers"),
		ExposedHeaders:   viper.GetStringSlice("cors.exposed_headers"),
		AllowCredentials: viper.GetBool("cors.allow_credentials"),
		MaxAge:           viper.GetInt("cors.max_age"),
	}))

	fsys, err := fs.Sub(content, "static")
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(fsys)))) /**/

	r.Get("/openapi3.json", func(w http.ResponseWriter, r *http.Request) {
		f, err := rest.GetOpenAPIFile()
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(f)
	})

	// These are executed after all Chi Middleware, right before the RestAPI function
	inlineMiddlewares := make([]rest.MiddlewareFunc, 0)
	inlineMiddlewares = append(inlineMiddlewares, middleware.SetHeader("Content-Type", "application/json"))
	inlineMiddlewares = append(inlineMiddlewares, middleware.RateControl(
		viper.GetInt("rate_control.req_per_hour"),
		viper.GetInt("rate_control.maxburst"),
		viper.GetDuration("rate_control.cleanup"),
	))
	inlineMiddlewares = append(inlineMiddlewares, middleware.OapiRequestValidator(swagger))
	inlineMiddlewares = append(inlineMiddlewares, middleware.PolicyValidator())
	rest.HandlerWithOptions(restApi, rest.ChiServerOptions{
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
		err := srv.Shutdown(ctxTimeout)

		if err != nil {
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
