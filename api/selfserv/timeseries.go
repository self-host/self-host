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

package selfserv

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/self-host/self-host/api/selfserv/rest"
	ie "github.com/self-host/self-host/internal/errors"
	"github.com/self-host/self-host/internal/services"
)

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

	created_by_uuid, err := u.GetUserUuidFromToken(r.Context(), []byte(domaintoken.Token))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	params := &services.NewTimeseriesParams{
		CreatedBy: created_by_uuid,
		Name:      n.Name,
		SiUnit:    n.SiUnit,
	}
	if n.Tags != nil {
		params.Tags = *n.Tags
	}

	if n.ThingUuid != nil {
		thing_uuid, err := uuid.Parse(*n.ThingUuid)
		if err != nil {
			ie.SendHTTPError(w, ie.ErrorInvalidUUID)
			return
		}

		params.ThingUuid = thing_uuid
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

func (ra *RestApi) AddDataToTimeseries(w http.ResponseWriter, r *http.Request, id rest.UuidParam, p rest.AddDataToTimeseriesParams) {
	ts_uuid, err := uuid.Parse(string(id))
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
	created_by, err := u.GetUserUuidFromToken(r.Context(), []byte(domaintoken.Token))
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
		Uuid:      ts_uuid,
		Points:    points,
		CreatedBy: created_by,
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

func (ra *RestApi) QueryTimeseriesForData(w http.ResponseWriter, r *http.Request, id rest.UuidParam, p rest.QueryTimeseriesForDataParams) {
	ts_uuid, err := uuid.Parse(string(id))
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
	ok, err := svc.Exists(r.Context(), ts_uuid)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	} else if ok == false {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	if time.Time(p.End).Sub(time.Time(p.Start)) > 31556952*time.Second {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	params := services.QueryDataParams{
		Uuid:        ts_uuid,
		Start:       time.Time(p.Start),
		End:         time.Time(p.End),
		GreaterOrEq: (*float32)(p.Ge),
		LessOrEq:    (*float32)(p.Le),
		Unit:        (*string)(p.Unit),
	}

	data, err := svc.QueryData(r.Context(), params)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

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

func (ra *RestApi) FindTimeSeriesByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	ts_uuid, err := uuid.Parse(string(id))
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
	timeseries, err := svc.FindByUuid(r.Context(), ts_uuid)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(timeseries)
}

func (ra *RestApi) UpdateTimeseriesByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	ts_uuid, err := uuid.Parse(string(id))
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
		Uuid:   ts_uuid,
		Name:   obj.Name,
		SiUnit: obj.SiUnit,
		Tags:   obj.Tags,
	}

	if obj.ThingUuid != nil {
		thing_uuid, err := uuid.Parse(*obj.ThingUuid)
		if err != nil {
			ie.SendHTTPError(w, ie.ErrorMalformedRequest)
			return
		}
		params.ThingUuid = &thing_uuid
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

func (ra *RestApi) DeleteTimeSeriesByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	ts_uuid, err := uuid.Parse(string(id))
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

	count, err := svc.DeleteTimeseries(r.Context(), ts_uuid)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	} else if count == 0 {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (ra *RestApi) DeleteDataFromTimeSeries(w http.ResponseWriter, r *http.Request, id rest.UuidParam, p rest.DeleteDataFromTimeSeriesParams) {
	ts_uuid, err := uuid.Parse(string(id))
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
	ok, err := svc.Exists(r.Context(), ts_uuid)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	} else if ok == false {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	if time.Time(p.End).Sub(time.Time(p.Start)) > 31556952*time.Second {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	params := services.DeleteTsDataParams{
		Uuid:        ts_uuid,
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
