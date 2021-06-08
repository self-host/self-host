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

package selfserv

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/noda/selfhost/api/selfserv/rest"
	ie "github.com/noda/selfhost/internal/errors"
	"github.com/noda/selfhost/internal/services"
)

func (ra *RestApi) AddDatasets(w http.ResponseWriter, r *http.Request) {
	// We expect a NewDataset object in the request body.
	var n rest.NewDataset
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	domaintoken, ok := r.Context().Value("domaintoken").(*services.DomainToken)
	if ok == false {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	u := services.NewUserService(db)
	created_by, err := u.GetUserUuidFromToken(r.Context(), []byte(domaintoken.Token))

	s := services.NewDatasetService(db)

	params := &services.AddDatasetParams{
		Name:      n.Name,
		Format:    string(n.Format),
		CreatedBy: created_by,
	}

	if n.BelongsTo != nil {
		params.BelongsTo, err = uuid.Parse(*n.BelongsTo)
		if err != nil {
			ie.SendHTTPError(w, ie.ErrorMalformedRequest)
			return
		}
	}

	if n.Content != nil {
		params.Content = *n.Content
	}

	// Add the dataset
	dataset, err := s.AddDataset(r.Context(), params)

	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dataset)
}

func (ra *RestApi) FindDatasets(w http.ResponseWriter, r *http.Request, p rest.FindDatasetsParams) {
	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	domaintoken, ok := r.Context().Value("domaintoken").(*services.DomainToken)
	if ok == false {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	svc := services.NewDatasetService(db)
	datasets, err := svc.FindAll(r.Context(), []byte(domaintoken.Token), (*int64)(p.Limit), (*int64)(p.Offset))
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(datasets)
}

func (ra *RestApi) FindDatasetByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	dataset_uuid, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	svc := services.NewDatasetService(db)
	datasets, err := svc.FindDatasetByUuid(r.Context(), dataset_uuid)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(datasets)
}

func (ra *RestApi) UpdateDatasetByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	dataset_uuid, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	// Allow max of 1 MB read from body
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	// We expect a UpdateDataset object in the request body.
	var updDataset rest.UpdateDataset
	if err := json.NewDecoder(r.Body).Decode(&updDataset); err != nil {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	svc := services.NewDatasetService(db)
	params := services.UpdateDatasetByUuidParams{
		Name:    updDataset.Name,
		Content: updDataset.Content,
	}

	if updDataset.Format != nil {
		s := string(*updDataset.Format)
		params.Format = &s
	}

	count, err := svc.UpdateDatasetByUuid(r.Context(), dataset_uuid, params)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	} else if count == 0 {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (ra *RestApi) GetRawDatasetByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	dataset_uuid, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	svc := services.NewDatasetService(db)
	f, err := svc.GetDatasetContentByUuid(r.Context(), dataset_uuid)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}
	if f == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Change Content-Type based on Dataset type
	switch f.Format {
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
	case "ini":
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	case "json":
		w.Header().Set("Content-Type", "application/json")
	case "toml":
		w.Header().Set("Content-Type", "application/toml")
	case "xml":
		w.Header().Set("Content-Type", "application/xml")
	case "yaml":
		w.Header().Set("Content-Type", "application/yaml")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	if len(f.Content) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("ETag", fmt.Sprintf("%x", md5.Sum(f.Content)))
	w.WriteHeader(http.StatusOK)
	w.Write(f.Content)
}

func (ra *RestApi) InitializeDatasetUploadByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (ra *RestApi) DeleteDatasetUploadByKey(w http.ResponseWriter, r *http.Request, id rest.UuidParam, p rest.DeleteDatasetUploadByKeyParams) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (ra *RestApi) ListDatasetPartsByKey(w http.ResponseWriter, r *http.Request, id rest.UuidParam, p rest.ListDatasetPartsByKeyParams) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (ra *RestApi) AssembleDatasetPartsByKey(w http.ResponseWriter, r *http.Request, id rest.UuidParam, p rest.AssembleDatasetPartsByKeyParams) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (ra *RestApi) UploadDatasetContentByKey(w http.ResponseWriter, r *http.Request, id rest.UuidParam, p rest.UploadDatasetContentByKeyParams) {
	// Each part may no exceed 5 MB in size
	r.Body = http.MaxBytesReader(w, r.Body, 5242880) // Max 5MB of data
	w.WriteHeader(http.StatusNotImplemented)
}

func (ra *RestApi) DeleteDatasetByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	dataset_uuid, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	svc := services.NewDatasetService(db)

	count, err := svc.DeleteDataset(r.Context(), dataset_uuid)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	} else if count == 0 {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
