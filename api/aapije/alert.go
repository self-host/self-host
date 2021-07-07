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

// CreateAlert creates a new alert
func (ra *RestApi) CreateAlert(w http.ResponseWriter, r *http.Request) {
	// We expect a NewAlert object in the request body.
	var n rest.NewAlert
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	s := services.NewAlertService(db)

	params := &services.CreateAlertParams{
		Resource:    n.Resource,
		Environment: n.Environment,
		Event:       n.Event,
		Severity:    n.Severity,
		Description: n.Description,
		Origin:      n.Origin,
		Value:       n.Value,
	}

	if n.Status != nil {
		params.Status = *n.Status
	} else {
		params.Status = rest.AlertStatusOpen
	}

	if n.Service != nil {
		params.Service = *n.Service
	} else {
		params.Service = make([]string, 0)
	}

	if n.Tags != nil {
		params.Tags = *n.Tags
	} else {
		params.Tags = make([]string, 0)
	}

	if n.Timeout != nil {
		params.Timeout = *n.Timeout
	} else {
		params.Timeout = 3600
	}

	if n.Rawdata != nil {
		params.Rawdata = *n.Rawdata
	} else {
		params.Rawdata = make([]byte, 0)
	}

	alert, err := s.CreateAlert(r.Context(), params)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(alert)
}

// FindAlerts lists alerts
func (ra *RestApi) FindAlerts(w http.ResponseWriter, r *http.Request, p rest.FindAlertsParams) {
	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	svc := services.NewAlertService(db)

	params := services.FindAllAlertParams{
		Status:     (*rest.AlertStatus)(p.Status),
		SeverityLe: (*rest.AlertSeverity)(p.SeverityLe),
		SeverityGe: (*rest.AlertSeverity)(p.SeverityGe),
		SeverityEq: (*rest.AlertSeverity)(p.Severity),
	}

	if p.Limit != nil {
		params.ArgLimit = int64(*p.Limit)
	} else {
		params.ArgLimit = 20
	}

	if p.Offset != nil {
		params.ArgOffset = int64(*p.Offset)
	}

	if p.Resource != nil {
		params.Resource = string(*p.Resource)
	}
	if p.Environment != nil {
		params.Environment = string(*p.Environment)
	}
	if p.Event != nil {
		params.Event = string(*p.Event)
	}
	if p.Origin != nil {
		params.Origin = string(*p.Origin)
	}
	if p.Service != nil {
		params.Service = []string(*p.Service)
	}
	if p.Tags != nil {
		params.Tags = []string(*p.Tags)
	}

	alerts, err := svc.FindAll(r.Context(), params)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(alerts)
}

// FindAlertByUuid gets the content of a specific alert
func (ra *RestApi) FindAlertByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	alertUUID, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	svc := services.NewAlertService(db)
	alert, err := svc.FindAlertByUuid(r.Context(), alertUUID)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(alert)
}

// UpdateAlertByUuid update an alert with new content
func (ra *RestApi) UpdateAlertByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	alertUUID, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	// We expect a UpdateAlert object in the request body.
	var updAlert rest.UpdateAlert
	if err := json.NewDecoder(r.Body).Decode(&updAlert); err != nil {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	svc := services.NewAlertService(db)
	params := services.UpdateAlertByUuidParams{
		Resource:    updAlert.Resource,
		Environment: updAlert.Environment,
		Event:       updAlert.Event,
		Severity:    updAlert.Severity,
		Status:      updAlert.Status,
		Value:       updAlert.Value,
		Description: updAlert.Description,
		Origin:      updAlert.Origin,
		Service:     updAlert.Service,
		Tags:        updAlert.Tags,
		Timeout:     updAlert.Timeout,
		Rawdata:     updAlert.Rawdata,
	}

	count, err := svc.UpdateAlertByUuid(r.Context(), alertUUID, params)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	} else if count == 0 {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteAlertByUuid deletes an alert
func (ra *RestApi) DeleteAlertByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	alertUUID, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	svc := services.NewAlertService(db)

	count, err := svc.DeleteAlert(r.Context(), alertUUID)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	} else if count == 0 {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
