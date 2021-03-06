// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=types.cfg.yaml openapiv3.yaml
//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=server.cfg.yaml openapiv3.yaml

package juvuln

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	ie "github.com/self-host/self-host/internal/errors"
	"github.com/self-host/self-host/pkg/workforce"
	"github.com/self-host/self-host/postgres"
)

// Main struct for the RestApi implementation
type RestApi struct{}

// Create a new RestApi instance
func New() *RestApi {
	return &RestApi{}
}

var (
	pcache *ProgramCache
	logger *zap.Logger
)

func init() {
	var err error

	pcache = &ProgramCache{
		m:       make(map[string]*ProgramRevision),
		deletes: make(map[string]struct{}),
	}

	logger, err = zap.NewProduction()
	if err != nil {
		panic("zap.NewProduction " + err.Error())
	}
}

// Self registration subscription endpoint
func (ra *RestApi) WorkerSubscribe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// We expect a NewSubscriber object in the request body.
	var sub NewSubscriber
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	workforce.Add(sub.Uuid, NewWorker(
		sub.Uuid,
		fmt.Sprintf("%v://%v", sub.Scheme, sub.Authority),
		sub.Languages,
		viper.GetDuration("worker.timeout")))

	w.WriteHeader(http.StatusCreated)
}

// Self un-registration subscription endpoint
func (ra *RestApi) WorkerUnsubscribe(w http.ResponseWriter, r *http.Request, id UuidParam) {
	workforce.Delete(string(id))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// Forward a call to a webhook program
func (ra *RestApi) ForwardWebhook(w http.ResponseWriter, r *http.Request, dom DomainPathParam, id UuidParam) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Check registration
func (ra *RestApi) CheckWorker(w http.ResponseWriter, r *http.Request, id UuidParam) {
	if workforce.Exists(string(id)) == false {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Report the current load
func (ra *RestApi) WorkerLoadUpdate(w http.ResponseWriter, r *http.Request, id UuidParam) {
	if workforce.Exists(string(id)) == false {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	// We expect a UpdateLoad object in the request body.
	var obj UpdateLoad
	if err := json.NewDecoder(r.Body).Decode(&obj); err != nil {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	workforce.SetLoad(string(id), uint64(obj.Load))

	w.WriteHeader(http.StatusNoContent)
}

// Get the code for a program module
func (ra *RestApi) GetModuleAtRevision(w http.ResponseWriter, r *http.Request, p GetModuleAtRevisionParams) {
	db, err := postgres.GetDB(string(p.Domain))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	q := postgres.New(db)
	var code []byte

	if string(p.Revision) == "latest" {
		code, err = q.GetNamedModuleCodeAtHead(r.Context(), postgres.GetNamedModuleCodeAtHeadParams{
			Name:     string(p.Module),
			Language: string(p.Language),
		})
		if err != nil {
			ie.SendHTTPError(w, ie.ParseDBError(err))
			return
		}
	} else {
		irev, err := strconv.Atoi(string(p.Revision))
		if err != nil {
			ie.SendHTTPError(w, ie.ErrorMalformedRequest)
			return
		}

		code, err = q.GetNamedModuleCodeAtRevision(r.Context(), postgres.GetNamedModuleCodeAtRevisionParams{
			Name:     string(p.Module),
			Revision: int32(irev),
			Language: string(p.Language),
		})
		if err != nil {
			ie.SendHTTPError(w, ie.ParseDBError(err))
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("ETag", fmt.Sprintf("%x", md5.Sum(code)))
	w.WriteHeader(http.StatusOK)
	w.Write(code)
}

// Get the OpenAPI file
func GetOpenAPIFile() ([]byte, error) {
	return decodeSpec()
}

// Update the program listing by queries each DB for content
func UpdateProgramCache() error {
	dbs := postgres.GetAllDB()

	pcache.Begin()

	for _, item := range dbs {
		if item.DB == nil {
			continue
		}

		q := postgres.New(item.DB)
		if q == nil {
			continue
		}

		ctx := context.Background() // FIXME: timeout
		routines, err := q.FindAllRoutineRevisions(ctx)
		if err != nil {
			return nil
		}

		for _, p := range routines {
			pcache.Add(NewProgramRevision(
				item.Domain,
				p.Name,
				p.ProgramUuid,
				p.Type,
				p.Schedule,
				p.Deadline,
				p.Language,
				p.Revision,
				p.Code,
				p.Checksum,
			))
		}
	}

	pcache.Commit()

	return nil
}
