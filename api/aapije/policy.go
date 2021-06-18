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
	"github.com/self-host/self-host/pkg/util"
)

func (ra *RestApi) AddPolicy(w http.ResponseWriter, r *http.Request) {
	// We expect a NewPolicy object in the request body.
	var newPolicy rest.NewPolicy
	if err := json.NewDecoder(r.Body).Decode(&newPolicy); err != nil {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	groupUUID, err := uuid.Parse(newPolicy.GroupUuid)
	if err != nil {
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

	// Ensure that the User has the right to create a policy with these access rules
	pc := services.NewPolicyCheckService(db)
	canGrant, err := pc.UserHasAccessViaToken(r.Context(), []byte(domaintoken.Token), string(newPolicy.Action), newPolicy.Resource)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	} else if canGrant == false {
		ie.SendHTTPError(w, ie.ErrorForbidden)
		return
	}

	srv := services.NewPolicyService(db)

	params := services.NewPolicyParams{
		GroupUuid: groupUUID,
		Priority:  int32(newPolicy.Priority),
		Effect:    string(newPolicy.Effect),
		Action:    string(newPolicy.Action),
		Resource:  newPolicy.Resource,
	}

	policy, err := srv.Add(r.Context(), params)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(policy)
}

func (ra *RestApi) FindPolicies(w http.ResponseWriter, r *http.Request, p rest.FindPoliciesParams) {
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

	srv := services.NewPolicyService(db)

	params := services.FindAllPoliciesParams{
		Token: []byte(domaintoken.Token),
	}
	if p.Limit != nil {
		i := int64(*p.Limit)
		params.Limit = &i
	}
	if p.Offset != nil {
		i := int64(*p.Offset)
		params.Offset = &i
	}

	if p.GroupUuids != nil {
		groupUUIDs, err := util.StringSliceToUuidSlice(*p.GroupUuids)
		if err != nil {
			ie.SendHTTPError(w, ie.ErrorMalformedRequest)
			return
		}

		params.GroupUuids = &groupUUIDs
	}

	policies, err := srv.FindAll(r.Context(), params)
	if err != nil {
		// FIXME: log
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(policies)
}

func (ra *RestApi) FindPolicyByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	policyUUID, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	s := services.NewPolicyService(db)
	policy, err := s.FindByUuid(r.Context(), policyUUID)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(policy)
}

func (ra *RestApi) UpdatePolicyByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	// We expect a UpdatePolicy object in the request body.
	var updatePolicy rest.UpdatePolicy
	if err := json.NewDecoder(r.Body).Decode(&updatePolicy); err != nil {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	policyUUID, err := uuid.Parse(string(id))
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
	params := services.UpdatePolicyParams{
		Priority: updatePolicy.Priority,
		Effect:   (*string)(updatePolicy.Effect),
		Action:   (*string)(updatePolicy.Action),
		Resource: updatePolicy.Resource,
	}

	if updatePolicy.GroupUuid != nil {
		groupUUID, err := uuid.Parse(*updatePolicy.GroupUuid)
		if err != nil {
			ie.SendHTTPError(w, ie.ErrorMalformedRequest)
			return
		}
		params.GroupUuid = &groupUUID
	}

	_, err = srv.Update(r.Context(), policyUUID, params)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (ra *RestApi) DeletePolicyByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	policyUUID, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	s := services.NewPolicyService(db)
	_, err = s.Delete(r.Context(), policyUUID)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
