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

// Add a new group
func (ra *RestApi) AddGroup(w http.ResponseWriter, r *http.Request) {
	// We expect a NewGroup object in the request body.
	var n rest.NewGroup
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	s := services.NewGroupService(db)

	// Add the group
	group, err := s.AddGroup(r.Context(), n.Name)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(group)
}

// List groups
func (ra *RestApi) FindGroups(w http.ResponseWriter, r *http.Request, p rest.FindGroupsParams) {
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

	gsrv := services.NewGroupService(db)
	groups, err := gsrv.FindAll(r.Context(), []byte(domaintoken.Token), (*int64)(p.Limit), (*int64)(p.Offset))
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(groups)
}

// Find a specific group by its UUID
func (ra *RestApi) FindGroupByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	groupUUID, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	gsrv := services.NewGroupService(db)
	group, err := gsrv.FindGroupByUuid(r.Context(), groupUUID)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(group)
}

// List all policies beloning to a specific group
func (ra *RestApi) FindPoliciesForGroup(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	groupUUID, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	srv := services.NewPolicyService(db)

	policies, err := srv.FindByGroup(r.Context(), groupUUID)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(policies)
}

// Update a specific group by its UUID
func (ra *RestApi) UpdateGroupByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	groupUUID, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	// We expect a UpdateUser object in the request body.
	var obj rest.UpdateGroup
	if err := json.NewDecoder(r.Body).Decode(&obj); err != nil {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	svc := services.NewGroupService(db)

	count, err := svc.UpdateGroupNameByUuid(r.Context(), groupUUID, obj.Name)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		// FIXME: log
		return
	} else if count == 0 {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Delete a specific group by its UUID
func (ra *RestApi) DeleteGroupByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	groupUUID, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	svc := services.NewGroupService(db)

	count, err := svc.DeleteGroup(r.Context(), groupUUID)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	} else if count == 0 {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
