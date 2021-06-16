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
//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=types.cfg.yaml openapiv3.yaml
//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=server.cfg.yaml openapiv3.yaml

package selfpmgr

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/self-host/self-host/pkg/workforce"
	ie "github.com/self-host/self-host/internal/errors"
	pg "github.com/self-host/self-host/postgres"
)

type RestApi struct{}

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

func (ra *RestApi) WorkerSubscribe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// We expect a NewSubscriber object in the request body.
	var sub NewSubscriber
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	workforce.Add(sub.Uuid, NewWorker(
		fmt.Sprintf("%v://%v", sub.Scheme, sub.Authority),
		sub.Languages))

	w.WriteHeader(http.StatusCreated)
}

func (ra *RestApi) WorkerUnsubscribe(w http.ResponseWriter, r *http.Request, id UuidParam) {
	workforce.Delete(string(id))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func (ra *RestApi) ForwardWebhook(w http.ResponseWriter, r *http.Request, dom DomainPathParam, id UuidParam) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (ra *RestApi) CheckWorker(w http.ResponseWriter, r *http.Request, id UuidParam) {
	if workforce.Exists(string(id)) == false {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

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

func (ra *RestApi) GetModuleAtRevision(w http.ResponseWriter, r *http.Request, p GetModuleAtRevisionParams) {
	db, err := pg.GetDB(string(p.Domain))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	q := pg.New(db)
	var code []byte

	if string(p.Revision) == "latest" {
		code, err = q.GetNamedModuleCodeAtHead(r.Context(), pg.GetNamedModuleCodeAtHeadParams{
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

		code, err = q.GetNamedModuleCodeAtRevision(r.Context(), pg.GetNamedModuleCodeAtRevisionParams{
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

func GetOpenAPIFile() ([]byte, error) {
	return decodeSpec()
}

func UpdateProgramCache() error {
	dbs := pg.GetAllDB()

	pcache.Begin()

	for _, item := range dbs {
		if item.DB == nil {
			continue
		}

		q := pg.New(item.DB)
		if q == nil {
			continue
		}

		ctx := context.Background() // FIXME: timeout
		routines, err := q.FindAllRoutineRevisions(ctx)
		if err != nil {
			return nil
		}

		for _, p := range routines {
			rev, err := NewProgramRevision(
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
			)
			if err != nil {
				continue
				// Ignore, log? do not return
			}

			pcache.Add(rev)
		}
	}

	pcache.Commit()

	return nil
}
