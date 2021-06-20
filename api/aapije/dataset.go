// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package aapije

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/self-host/self-host/api/aapije/rest"
	ie "github.com/self-host/self-host/internal/errors"
	"github.com/self-host/self-host/internal/services"
)

// AddDatasets adds a new dataset
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
	createdBy, err := u.GetUserUuidFromToken(r.Context(), []byte(domaintoken.Token))

	s := services.NewDatasetService(db)

	params := &services.AddDatasetParams{
		Name:      n.Name,
		Format:    string(n.Format),
		CreatedBy: createdBy,
	}
	if n.Tags != nil {
		params.Tags = *n.Tags
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

// FindDatasets lists all datasets
func (ra *RestApi) FindDatasets(w http.ResponseWriter, r *http.Request, p rest.FindDatasetsParams) {
	var err error
	var datasets []*rest.Dataset

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

	if p.Tags != nil {
		params := services.NewFindByTagsParams(
			[]byte(domaintoken.Token),
			*p.Tags,
			(*int64)(p.Limit),
			(*int64)(p.Offset))

		if params.Limit.Value == 0 {
			params.Limit.Value = 20
		}

		datasets, err = svc.FindByTags(r.Context(), params)
		if err != nil {
			ie.SendHTTPError(w, ie.ParseDBError(err))
			return
		}
	} else {
		params := services.NewFindAllParams(
			[]byte(domaintoken.Token),
			(*int64)(p.Limit),
			(*int64)(p.Offset))

		if params.Limit.Value == 0 {
			params.Limit.Value = 20
		}

		datasets, err = svc.FindAll(r.Context(), params)
		if err != nil {
			ie.SendHTTPError(w, ie.ParseDBError(err))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(datasets)
}

// FindDatasetByUuid returns a specific dataset by its UUID
func (ra *RestApi) FindDatasetByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	datasetUUID, err := uuid.Parse(string(id))
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
	datasets, err := svc.FindDatasetByUuid(r.Context(), datasetUUID)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(datasets)
}

// UpdateDatasetByUuid updates a dataset by its UUID
func (ra *RestApi) UpdateDatasetByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	datasetUUID, err := uuid.Parse(string(id))
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
		Tags:    updDataset.Tags,
	}

	if updDataset.Format != nil {
		s := string(*updDataset.Format)
		params.Format = &s
	}

	count, err := svc.UpdateDatasetByUuid(r.Context(), datasetUUID, params)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	} else if count == 0 {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetRawDatasetByUuid gets the "file" content from a dataset by its UUID
func (ra *RestApi) GetRawDatasetByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam, p rest.GetRawDatasetByUuidParams) {
	datasetUUID, err := uuid.Parse(string(id))
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
	f, err := svc.GetDatasetContentByUuid(r.Context(), datasetUUID)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}
	if f == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if p.IfNoneMatch != nil && (string)(*p.IfNoneMatch) == f.Checksum {
		w.WriteHeader(http.StatusNotModified)
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

	w.Header().Set("ETag", f.Checksum)
	w.WriteHeader(http.StatusOK)
	w.Write(f.Content)
}

// InitializeDatasetUploadByUuid initiates the upload of a larger dataset
func (ra *RestApi) InitializeDatasetUploadByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	w.WriteHeader(http.StatusNotImplemented)
}

// DeleteDatasetUploadByKey cancels a partialy completed upload
func (ra *RestApi) DeleteDatasetUploadByKey(w http.ResponseWriter, r *http.Request, id rest.UuidParam, p rest.DeleteDatasetUploadByKeyParams) {
	w.WriteHeader(http.StatusNotImplemented)
}

// ListDatasetPartsByKey lists all uploaded parts of the dataset
func (ra *RestApi) ListDatasetPartsByKey(w http.ResponseWriter, r *http.Request, id rest.UuidParam, p rest.ListDatasetPartsByKeyParams) {
	w.WriteHeader(http.StatusNotImplemented)
}

// AssembleDatasetPartsByKey combines all uploaded parts into a new dataset content
func (ra *RestApi) AssembleDatasetPartsByKey(w http.ResponseWriter, r *http.Request, id rest.UuidParam, p rest.AssembleDatasetPartsByKeyParams) {
	w.WriteHeader(http.StatusNotImplemented)
}

// UploadDatasetContentByKey uploads a (max 5MB) part of a new content update to a dataset
func (ra *RestApi) UploadDatasetContentByKey(w http.ResponseWriter, r *http.Request, id rest.UuidParam, p rest.UploadDatasetContentByKeyParams) {
	// Each part may no exceed 5 MB in size
	r.Body = http.MaxBytesReader(w, r.Body, 5242880) // Max 5MB of data
	w.WriteHeader(http.StatusNotImplemented)
}

// DeleteDatasetByUuid deletes a dataset by its UUID
func (ra *RestApi) DeleteDatasetByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	datasetUUID, err := uuid.Parse(string(id))
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

	count, err := svc.DeleteDataset(r.Context(), datasetUUID)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	} else if count == 0 {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
