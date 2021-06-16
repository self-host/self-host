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

package aapije

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/self-host/self-host/api/aapije/rest"
	ie "github.com/self-host/self-host/internal/errors"
	"github.com/self-host/self-host/internal/services"
)

func (ra *RestApi) AddUser(w http.ResponseWriter, r *http.Request) {
	// We expect a NewUser object in the request body.
	var newUser rest.NewUser
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	s := services.NewUserService(db)

	// Add the user
	user, err := s.AddUser(r.Context(), newUser.Name)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (ra *RestApi) AddNewTokenToUser(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	user_uuid, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	// We expect a NewUser object in the request body.
	var newToken rest.NewToken
	if err := json.NewDecoder(r.Body).Decode(&newToken); err != nil {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	s := services.NewUserService(db)

	// Add token to User
	user, err := s.AddTokenToUser(r.Context(), user_uuid, newToken.Name)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (ra *RestApi) Whoami(w http.ResponseWriter, r *http.Request) {
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

	s := services.NewUserService(db)
	id, err := s.GetUserUuidFromToken(r.Context(), []byte(domaintoken.Token))
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	user, err := s.FindUserByUuid(r.Context(), id)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (ra *RestApi) FindUsers(w http.ResponseWriter, r *http.Request, p rest.FindUsersParams) {
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

	s := services.NewUserService(db)
	users, err := s.FindAll(r.Context(), []byte(domaintoken.Token), (*int64)(p.Limit), (*int64)(p.Offset))
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func (ra *RestApi) FindUserByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	user_uuid, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	s := services.NewUserService(db)
	user, err := s.FindUserByUuid(r.Context(), user_uuid)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (ra *RestApi) FindTokensForUser(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	user_uuid, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	s := services.NewUserService(db)

	tokens, err := s.FindTokensForUser(r.Context(), user_uuid)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokens)
}

func (ra *RestApi) FindPoliciesForUser(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	user_uuid, err := uuid.Parse(string(id))
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

	policies, err := srv.FindByUser(r.Context(), user_uuid)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(policies)
}

func (ra *RestApi) UpdateUserByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	user_uuid, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	// Check if user exits
	svc := services.NewUserService(db)
	_, err = svc.FindUserByUuid(r.Context(), user_uuid)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	// We expect a UpdateUser object in the request body.
	var updUser rest.UpdateUser
	if err := json.NewDecoder(r.Body).Decode(&updUser); err != nil {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	// We do not allow groups and (groups_add or groups_remove)
	if updUser.Groups != nil && (updUser.GroupsAdd != nil || updUser.GroupsRemove != nil) {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	if updUser.Name != nil && len(*updUser.Name) > 3 {
		_, err := svc.SetUserName(r.Context(), user_uuid, *updUser.Name)
		if err != nil {
			ie.SendHTTPError(w, ie.ParseDBError(err))
			return
		}
	}

	// Use multiple DB requests to set each parameter
	// Use a Transaction!
	if updUser.Groups != nil {
		group_uuids := make([]uuid.UUID, 0)
		for _, item := range *updUser.Groups {
			uid, err := uuid.Parse(item)
			if err != nil {
				ie.SendHTTPError(w, ie.ErrorMalformedRequest)
				return
			}
			group_uuids = append(group_uuids, uid)
		}

		// FIXME: handle count value
		_, err := svc.SetUserGroups(r.Context(), user_uuid, group_uuids)
		if err != nil {
			ie.SendHTTPError(w, ie.ParseDBError(err))
			return
		}

	} else if updUser.GroupsAdd != nil || updUser.GroupsRemove != nil {
		add_group_uuids := make([]uuid.UUID, 0)
		remove_group_uuids := make([]uuid.UUID, 0)

		if updUser.GroupsAdd != nil {
			for _, item := range *updUser.GroupsAdd {
				uid, err := uuid.Parse(item)
				if err != nil {
					ie.SendHTTPError(w, ie.ErrorMalformedRequest)
					return
				}
				add_group_uuids = append(add_group_uuids, uid)
			}
		}

		if updUser.GroupsRemove != nil {
			for _, item := range *updUser.GroupsRemove {
				uid, err := uuid.Parse(item)
				if err != nil {
					ie.SendHTTPError(w, ie.ErrorMalformedRequest)
					return
				}
				remove_group_uuids = append(remove_group_uuids, uid)
			}
		}

		// FIXME: Should we validate the group uuids?
		// FIXME: handle count value
		_, err := svc.AddRemoveUserToGroups(r.Context(), user_uuid, add_group_uuids, remove_group_uuids)
		if err != nil {
			ie.SendHTTPError(w, ie.ParseDBError(err))
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (ra *RestApi) SetRequestRateForUser(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (ra *RestApi) DeleteUserByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	user_uuid, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	s := services.NewUserService(db)

	_, err = s.DeleteUser(r.Context(), user_uuid)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (ra *RestApi) DeleteTokenForUser(w http.ResponseWriter, r *http.Request, id rest.UuidParam, token_id string) {
	user_uuid, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	token_uuid, err := uuid.Parse(token_id)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	s := services.NewUserService(db)

	_, err = s.DeleteTokenFromUser(r.Context(), user_uuid, token_uuid)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
