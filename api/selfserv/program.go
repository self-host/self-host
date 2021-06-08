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
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/noda/selfhost/api/selfserv/rest"
	ie "github.com/noda/selfhost/internal/errors"
	"github.com/noda/selfhost/internal/services"
)

func (ra *RestApi) AddProgram(w http.ResponseWriter, r *http.Request) {
	// We expect a NewProgram object in the request body.
	var newProgram rest.NewProgram
	if err := json.NewDecoder(r.Body).Decode(&newProgram); err != nil {
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

	s := services.NewProgramService(db)
	prog, err := s.AddProgram(r.Context(), services.AddProgramParams{
		Name:      newProgram.Name,
		Type:      string(newProgram.Type),
		State:     string(newProgram.State),
		Schedule:  string(newProgram.Schedule),
		Deadline:  newProgram.Deadline,
		Language:  string(newProgram.Language),
		CreatedBy: created_by,
	})

	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(prog)
}

func (ra *RestApi) FindPrograms(w http.ResponseWriter, r *http.Request, p rest.FindProgramsParams) {
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

	s := services.NewProgramService(db)
	programs, err := s.FindAll(r.Context(), services.FindAllProgramsParams{
		Token:  []byte(domaintoken.Token),
		Limit:  (*int64)(p.Limit),
		Offset: (*int64)(p.Offset),
	})
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(programs)
}

func (ra *RestApi) FindProgramByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	program_uuid, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	s := services.NewProgramService(db)
	program, err := s.FindProgramByUuid(r.Context(), program_uuid)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(program)
}

func (ra *RestApi) UpdateProgramByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	program_uuid, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	// We expect a UpdateProgram object in the request body.
	var updProgram rest.UpdateProgram
	if err := json.NewDecoder(r.Body).Decode(&updProgram); err != nil {
		ie.SendHTTPError(w, ie.ErrorMalformedRequest)
		return
	}

	svc := services.NewProgramService(db)
	params := services.UpdateProgramByUuidParams{
		Name:     updProgram.Name,
		Type:     (*string)(updProgram.Type),
		State:    (*string)(updProgram.State),
		Schedule: updProgram.Schedule,
		Deadline: updProgram.Deadline,
		Language: (*string)(updProgram.Language),
	}

	count, err := svc.UpdateProgramByUuid(r.Context(), program_uuid, params)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	} else if count == 0 {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (ra *RestApi) DeleteProgramByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	program_uuid, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	s := services.NewProgramService(db)
	count, err := s.DeleteProgram(r.Context(), program_uuid)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	} else if count == 0 {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (ra *RestApi) AddProgramCodeRevision(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	program_uuid, err := uuid.Parse(string(id))
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

	max_upload_size := 1048576
	content_length, err := strconv.Atoi(r.Header.Get("Content-Length"))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorLengthRequired)
		return
	} else if content_length > max_upload_size {
		ie.SendHTTPError(w, ie.ErrorRequestEntityTooLarge)
		return
	}

	// Read at most X MB of data from request body
	r.Body = http.MaxBytesReader(w, r.Body, int64(max_upload_size))

	b, err := io.ReadAll(r.Body)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
	}

	s := services.NewProgramService(db)
	revision, err := s.AddCodeRevision(r.Context(), services.AddCodeRevisionParams{
		ProgramUuid: program_uuid,
		CreatedBy:   created_by,
		Code:        b,
	})
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(revision)
}

func (ra *RestApi) GetProgramCodeRevisionsDiff(w http.ResponseWriter, r *http.Request, id rest.UuidParam, p rest.GetProgramCodeRevisionsDiffParams) {
	program_uuid, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	s := services.NewProgramService(db)
	diff, err := s.DiffProgramCodeAtRevisions(r.Context(), program_uuid, p.RevA, p.RevB)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(diff))
}

func (ra *RestApi) GetCodeFromProgram(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	program_uuid, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	s := services.NewProgramService(db)
	code, err := s.GetSignedProgramCodeAtHead(r.Context(), program_uuid)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(code))
}

func (ra *RestApi) GetProgramCodeRevisions(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	program_uuid, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	s := services.NewProgramService(db)
	revisions, err := s.FindAllCodeRevisions(r.Context(), program_uuid)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(revisions)
}

func (ra *RestApi) ExecuteProgramWebhook(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (ra *RestApi) SignProgramCodeRevisions(w http.ResponseWriter, r *http.Request, id rest.UuidParam, revision int) {
	program_uuid, err := uuid.Parse(string(id))
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
	signed_by, err := u.GetUserUuidFromToken(r.Context(), []byte(domaintoken.Token))

	s := services.NewProgramService(db)
	count, err := s.SignCodeRevision(r.Context(), services.SignCodeRevisionParams{
		ProgramUuid: program_uuid,
		Revision:    revision,
		SignedBy:    signed_by,
	})
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	} else if count == 0 {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (ra *RestApi) DeleteProgramCodeRevisions(w http.ResponseWriter, r *http.Request, id rest.UuidParam, revision int) {
	program_uuid, err := uuid.Parse(string(id))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorInvalidUUID)
		return
	}

	db, err := ra.GetDB(r)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
		return
	}

	s := services.NewProgramService(db)
	count, err := s.DeleteProgramCodeRevision(r.Context(), program_uuid, revision)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	} else if count == 0 {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
