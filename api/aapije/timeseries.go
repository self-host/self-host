// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package aapije

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/self-host/self-host/api/aapije/rest"
	ie "github.com/self-host/self-host/internal/errors"
	"github.com/self-host/self-host/internal/services"
)

// AddTimeSeries adds a new time series
func (ra *RestApi) AddTimeSeries(w http.ResponseWriter, r *http.Request) {
	// We expect a NewUser object in the request body.
	var n rest.NewTimeseries
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

	createdByUUID, err := u.GetUserUuidFromToken(r.Context(), []byte(domaintoken.Token))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	params := &services.NewTimeseriesParams{
		CreatedBy: createdByUUID,
		Name:      n.Name,
		SiUnit:    n.SiUnit,
	}
	if n.Tags != nil {
		params.Tags = *n.Tags
	}

	if n.ThingUuid != nil {
		thingUUID, err := uuid.Parse(*n.ThingUuid)
		if err != nil {
			ie.SendHTTPError(w, ie.ErrorInvalidUUID)
			return
		}

		params.ThingUuid = thingUUID
	}

	if n.LowerBound != nil {
		params.LowerBound.Scan(*n.LowerBound)
	}

	if n.UpperBound != nil {
		params.UpperBound.Scan(*n.UpperBound)
	}

	s := services.NewTimeseriesService(db)

	// Add the time series
	thing, err := s.AddTimeseries(r.Context(), params)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(thing)
}

// AddDataToTimeseries adds data to a specific time series
func (ra *RestApi) AddDataToTimeseries(w http.ResponseWriter, r *http.Request, id rest.UuidParam, p rest.AddDataToTimeseriesParams) {
	tsUUID, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
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
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	svc := services.NewTimeseriesService(db)

	// Allow max of 5 MB read from body
	r.Body = http.MaxBytesReader(w, r.Body, 5242880)

	// We expect a NewTsData object in the request body.
	var obj rest.NewTsData
	if err := json.NewDecoder(r.Body).Decode(&obj); err != nil {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	if len(obj) == 0 {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	points := make([]services.DataPoint, len(obj))
	for i, element := range obj {
		points[i] = services.DataPoint{
			Value:     float64(element.V),
			Timestamp: element.Ts,
		}
	}

	count, err := svc.AddDataToTimeseries(r.Context(), services.AddDataToTimeseriesParams{
		Uuid:      tsUUID,
		Points:    points,
		CreatedBy: createdBy,
		Unit:      (*string)(p.Unit),
	})
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	} else if count == 0 {
		// No rows where inserted due to boundary checks
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// QueryTimeseriesForData returns data from a specific time series
func (ra *RestApi) QueryTimeseriesForData(w http.ResponseWriter, r *http.Request, id rest.UuidParam, p rest.QueryTimeseriesForDataParams) {
	tsUUID, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	svc := services.NewTimeseriesService(db)

	// Ensure the timeseries exists
	ok, err := svc.Exists(r.Context(), tsUUID)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	} else if ok == false {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	if time.Time(p.End).Sub(time.Time(p.Start)) > 31622401*time.Second {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	params := services.QuerySingleSourceDataParams{
		Uuid:        tsUUID,
		Start:       time.Time(p.Start),
		End:         time.Time(p.End),
		GreaterOrEq: (*float32)(p.Ge),
		LessOrEq:    (*float32)(p.Le),
		Unit:        (*string)(p.Unit),
	}

	if p.Timezone != nil {
		params.Timezone = string(*p.Timezone)
	} else {
		params.Timezone = "UTC"
	}

	if p.Aggregate != nil {
		params.Aggregate = string(*p.Aggregate)
	} else {
		params.Aggregate = "avg"
	}

	if p.Precision != nil {
		params.Precision = string(*p.Precision)
	} else {
		params.Precision = "microseconds"
	}

	data, err := svc.QuerySingleSourceData(r.Context(), params)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

// FindTimeSeries lists all time series
func (ra *RestApi) FindTimeSeries(w http.ResponseWriter, r *http.Request, p rest.FindTimeSeriesParams) {
	var err error
	var timeseries []*rest.Timeseries

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

	srv := services.NewTimeseriesService(db)

	if p.Tags != nil {
		params := services.NewFindByTagsParams(
			[]byte(domaintoken.Token),
			*p.Tags,
			(*int64)(p.Limit),
			(*int64)(p.Offset))

		if params.Limit.Value == 0 {
			params.Limit.Value = 20
		}

		timeseries, err = srv.FindByTags(r.Context(), params)
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

		timeseries, err = srv.FindAll(r.Context(), params)
		if err != nil {
			ie.SendHTTPError(w, ie.ParseDBError(err))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(timeseries)
}

// FindTimeSeriesByUuid returns a specific time series by its UUID
func (ra *RestApi) FindTimeSeriesByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	tsUUID, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	svc := services.NewTimeseriesService(db)
	timeseries, err := svc.FindByUuid(r.Context(), tsUUID)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(timeseries)
}

// UpdateTimeseriesByUuid updates a specific time series by its UUID
func (ra *RestApi) UpdateTimeseriesByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	tsUUID, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	svc := services.NewTimeseriesService(db)

	// We expect a UpdateTimeseries object in the request body.
	var obj rest.UpdateTimeseries
	if err := json.NewDecoder(r.Body).Decode(&obj); err != nil {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	params := services.UpdateTimeseriesParams{
		Uuid:   tsUUID,
		Name:   obj.Name,
		SiUnit: obj.SiUnit,
		Tags:   obj.Tags,
	}

	if obj.ThingUuid != nil {
		thingUUID, err := uuid.Parse(*obj.ThingUuid)
		if err != nil {
			ie.SendHTTPError(w, ie.ErrorMalformedRequest)
			return
		}
		params.ThingUuid = &thingUUID
	}

	if obj.LowerBound != nil {
		var v sql.NullFloat64
		err = v.Scan(*obj.LowerBound)
		if err != nil {
			ie.SendHTTPError(w, ie.ErrorMalformedRequest)
			return
		}
		params.LowerBound = &v
	}

	if obj.UpperBound != nil {
		var v sql.NullFloat64
		err = v.Scan(*obj.UpperBound)
		if err != nil {
			ie.SendHTTPError(w, ie.ErrorMalformedRequest)
			return
		}
		params.UpperBound = &v
	}

	count, err := svc.UpdateTimeseries(r.Context(), params)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	} else if count == 0 {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteTimeSeriesByUuid deletes a specific time series by its UUID
func (ra *RestApi) DeleteTimeSeriesByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	tsUUID, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	svc := services.NewTimeseriesService(db)

	count, err := svc.DeleteTimeseries(r.Context(), tsUUID)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	} else if count == 0 {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteDataFromTimeSeries deletes data from a time series
func (ra *RestApi) DeleteDataFromTimeSeries(w http.ResponseWriter, r *http.Request, id rest.UuidParam, p rest.DeleteDataFromTimeSeriesParams) {
	tsUUID, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	svc := services.NewTimeseriesService(db)

	// Ensure the timeseries exists
	ok, err := svc.Exists(r.Context(), tsUUID)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	} else if ok == false {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	if time.Time(p.End).Sub(time.Time(p.Start)) > 31622401*time.Second {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	params := services.DeleteTsDataParams{
		Uuid:        tsUUID,
		Start:       time.Time(p.Start),
		End:         time.Time(p.End),
		GreaterOrEq: (*float32)(p.Ge),
		LessOrEq:    (*float32)(p.Le),
	}

	_, err = svc.DeleteTsData(r.Context(), params)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
