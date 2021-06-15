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

//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=rest/types.cfg.yaml rest/openapiv3.yaml
//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=rest/server.cfg.yaml rest/openapiv3.yaml
//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=rest/client.cfg.yaml rest/openapiv3.yaml

package selfserv

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
