// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=rest/types.cfg.yaml rest/openapiv3.yaml
//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=rest/server.cfg.yaml rest/openapiv3.yaml
//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=rest/client.cfg.yaml rest/openapiv3.yaml

package aapije

import (
	"database/sql"
	"errors"
	"net/http"
)

type Error struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

type RestApi struct{}

func NewRestApi() *RestApi {
	return &RestApi{}
}

func (ra *RestApi) GetDB(r *http.Request) (*sql.DB, error) {
	db, ok := r.Context().Value("db").(*sql.DB)
	if ok == false {
		return nil, errors.New("database handle missing from context")
	}
	return db, nil
}
