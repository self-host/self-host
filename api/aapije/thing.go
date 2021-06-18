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

func (ra *RestApi) AddThing(w http.ResponseWriter, r *http.Request) {
	// We expect a NewThing object in the request body.
	var n rest.NewThing
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

	author, err := u.GetUserUuidFromToken(r.Context(), []byte(domaintoken.Token))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidAPIKey)
		return
	}

	s := services.NewThingService(db)

	params := &services.AddThingParams{
		Name:      n.Name,
		Type:      n.Type,
		CreatedBy: &author,
	}
	if n.Tags != nil {
		params.Tags = *n.Tags
	}

	// Add the thing
	thing, err := s.AddThing(r.Context(), params)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(thing)
}

func (ra *RestApi) FindThings(w http.ResponseWriter, r *http.Request, p rest.FindThingsParams) {
	var err error
	var things []*rest.Thing

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

	svc := services.NewThingService(db)

	if p.Tags != nil {
		params := services.NewFindByTagsParams(
			[]byte(domaintoken.Token),
			*p.Tags,
			(*int64)(p.Limit),
			(*int64)(p.Offset))

		if params.Limit.Value == 0 {
			params.Limit.Value = 20
		}

		things, err = svc.FindByTags(r.Context(), params)
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

		things, err = svc.FindAll(r.Context(), params)
		if err != nil {
			ie.SendHTTPError(w, ie.ParseDBError(err))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(things)
}

func (ra *RestApi) FindThingByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	thingUUID, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	svc := services.NewThingService(db)
	things, err := svc.FindThingByUuid(r.Context(), thingUUID)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(things)
}

func (ra *RestApi) FindTimeSeriesForThing(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	thingUUID, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	srv := services.NewTimeseriesService(db)

	timeseries, err := srv.FindByThing(r.Context(), thingUUID)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(timeseries)
}

func (ra *RestApi) FindDatasetsForThing(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	thingUUID, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	s := services.NewThingService(db)
	if ok, err := s.Exists(r.Context(), thingUUID); err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	} else if ok == false {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	srv := services.NewDatasetService(db)
	datasets, err := srv.FindByThing(r.Context(), thingUUID)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(datasets)
}

func (ra *RestApi) UpdateThingByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	thingUUID, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	svc := services.NewThingService(db)

	// We expect a UpdateThing object in the request body.
	var obj rest.UpdateThing
	if err := json.NewDecoder(r.Body).Decode(&obj); err != nil {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	params := services.UpdateThingParams{
		Uuid:  thingUUID,
		Name:  obj.Name,
		Type:  obj.Type,
		State: (*string)(obj.State),
		Tags:  obj.Tags,
	}

	count, err := svc.UpdateByUuid(r.Context(), params)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	} else if count == 0 {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (ra *RestApi) DeleteThingByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	thingUUID, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	svc := services.NewThingService(db)

	count, err := svc.DeleteThing(r.Context(), thingUUID)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	} else if count == 0 {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
