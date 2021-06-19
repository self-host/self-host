// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package aapije

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/self-host/self-host/api/aapije/rest"
	ie "github.com/self-host/self-host/internal/errors"
	"github.com/self-host/self-host/internal/services"
	"github.com/self-host/self-host/pkg/util"
)

func (ra *RestApi) FindTsdataByQuery(w http.ResponseWriter, r *http.Request, p rest.FindTsdataByQueryParams) {
	timezone := "UTC"
	aggregate := "avg"
	precision := "microseconds"

	if time.Time(p.End).Sub(time.Time(p.Start)) > 31556952*time.Second {
		// FIXME: these errors should be declared in one location
		ie.SendHTTPError(w, ie.NewBadRequestError(fmt.Errorf("start to end range exceeds limit")))
		return
	}

	if p.Timezone != nil {
		timezone = string(*p.Timezone)
	}

	if p.Aggregate != nil {
		aggregate = string(*p.Aggregate)
	}

	if p.Precision != nil {
		precision = string(*p.Precision)
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

	svc := services.NewTimeseriesService(db)
	policySvc := services.NewPolicyCheckService(db)

	var uuids []uuid.UUID

	if p.Tags == nil && p.Uuids == nil {
		ie.SendHTTPError(w, ie.NewBadRequestError(fmt.Errorf("requires on of `tags` or `uuids`")))
		return

	} else if p.Tags != nil {
		timeSvc := services.NewTimeseriesService(db)

		offset := int64(0)
		limit := int64(10) // Match at most 10 timeseries

		params := services.NewFindByTagsParams(
			[]byte(domaintoken.Token),
			*p.Tags,
			&limit,
			&offset,
		)

		timeseries, err := timeSvc.FindByTags(r.Context(), params)
		if err != nil {
			ie.SendHTTPError(w, ie.ParseDBError(err))
			return
		}

		uuids = make([]uuid.UUID, 0)
		for _, t := range timeseries {
			u, err := uuid.Parse(t.Uuid)
			if err != nil {
				continue // This _should_ not occur, unless someone breaks something
			}

			// Filter all items the User does not have access to
			//
			// While FindByTags does filter on permissions, it can
			// only do it on timeseries/{uuid} and not on timeseries{uuid}/data,
			// which is what we want in this instance.
			resource := fmt.Sprintf("timeseries/%v/data", t.Uuid)
			ok, err = policySvc.UserHasAccessViaToken(r.Context(), []byte(domaintoken.Token), "read", resource)
			if err != nil {
				ie.SendHTTPError(w, ie.ParseDBError(err))
				return
			} else if ok == false {
				continue // Skip
			}

			uuids = append(uuids, u)
		}
	} else if p.Uuids != nil {
		uuids, err = util.StringSliceToUuidSlice([]string(*p.Uuids))
		if err != nil {
			ie.SendHTTPError(w, ie.NewBadRequestError(fmt.Errorf("uuids has invalid format")))
			return
		}

		// Generate check rules for access control
		resources := make([]string, 0)
		for _, id := range uuids {
			resources = append(resources, fmt.Sprintf("timeseries/%v/data", id.String()))
		}

		// Ensure that the User has access to all requested items
		ok, err = policySvc.UserHasManyAccessViaToken(r.Context(), []byte(domaintoken.Token), "read", resources)
		if err != nil {
			ie.SendHTTPError(w, ie.ParseDBError(err))
			return
		} else if ok == false {
			// Access denied to one or more requested resources
			ie.SendHTTPError(w, ie.ErrorForbidden)
			return
		}
	}

	params := services.QueryMultiSourceDataParams{
		Uuids:       uuids,
		Start:       time.Time(p.Start),
		End:         time.Time(p.End),
		GreaterOrEq: (*float32)(p.Ge),
		LessOrEq:    (*float32)(p.Le),
		Aggregate:   aggregate,
		Precision:   precision,
		Timezone:    timezone,
	}

	data, err := svc.QueryMultiSourceData(r.Context(), params)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
	return
}
