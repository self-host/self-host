// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package aapije

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"

	"github.com/self-host/self-host/api/aapije/rest"
	ie "github.com/self-host/self-host/internal/errors"
	"github.com/self-host/self-host/internal/services"
)

// AddProgram adds a new program
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
	createdBy, err := u.GetUserUuidFromToken(r.Context(), []byte(domaintoken.Token))

	s := services.NewProgramService(db)

	// Validate CRON formated schedule
	specParser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	_, err = specParser.Parse(string(newProgram.Schedule))
	if err != nil {
		ie.SendHTTPError(w, ie.NewInvalidRequestError(
			fmt.Errorf("unable to parse schedule: \n%v", err),
		))
		return
	}

	params := services.AddProgramParams{
		Name:      newProgram.Name,
		Type:      string(newProgram.Type),
		State:     string(newProgram.State),
		Schedule:  string(newProgram.Schedule),
		Deadline:  newProgram.Deadline,
		Language:  string(newProgram.Language),
		CreatedBy: createdBy,
	}

	if newProgram.Tags != nil {
		params.Tags = *newProgram.Tags
	}

	prog, err := s.AddProgram(r.Context(), params)

	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(prog)
}

// FindPrograms lists all programs
func (ra *RestApi) FindPrograms(w http.ResponseWriter, r *http.Request, p rest.FindProgramsParams) {
	var err error
	var programs []*rest.Program

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

	svc := services.NewProgramService(db)

	if p.Tags != nil {
		params := services.NewFindByTagsParams(
			[]byte(domaintoken.Token),
			*p.Tags,
			(*int64)(p.Limit),
			(*int64)(p.Offset))

		if params.Limit.Value == 0 {
			params.Limit.Value = 20
		}

		programs, err = svc.FindByTags(r.Context(), params)
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

		programs, err = svc.FindAll(r.Context(), params)
		if err != nil {
			ie.SendHTTPError(w, ie.ParseDBError(err))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(programs)
}

// FindProgramByUuid returns a specific program by its UUID
func (ra *RestApi) FindProgramByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	programUUID, err := uuid.Parse(string(id))
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
	program, err := s.FindProgramByUuid(r.Context(), programUUID)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(program)
}

// UpdateProgramByUuid updates a specific program by its UUID
func (ra *RestApi) UpdateProgramByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	programUUID, err := uuid.Parse(string(id))
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

	// Validate CRON formated schedule
	if updProgram.Schedule != nil {
		specParser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
		_, err = specParser.Parse(string(*updProgram.Schedule))
		if err != nil {
			ie.SendHTTPError(w, ie.NewInvalidRequestError(
				fmt.Errorf("unable to parse schedule: \n%v", err),
			))
			return
		}
	}

	svc := services.NewProgramService(db)
	params := services.UpdateProgramByUuidParams{
		Name:     updProgram.Name,
		Type:     (*string)(updProgram.Type),
		State:    (*string)(updProgram.State),
		Schedule: updProgram.Schedule,
		Deadline: updProgram.Deadline,
		Language: (*string)(updProgram.Language),
		Tags:     updProgram.Tags,
	}

	count, err := svc.UpdateProgramByUuid(r.Context(), programUUID, params)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	} else if count == 0 {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteProgramByUuid deletes a specific program by its UUID
func (ra *RestApi) DeleteProgramByUuid(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	programUUID, err := uuid.Parse(string(id))
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
	count, err := s.DeleteProgram(r.Context(), programUUID)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	} else if count == 0 {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddProgramCodeRevision adds a new revision of code to a program
func (ra *RestApi) AddProgramCodeRevision(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	programUUID, err := uuid.Parse(string(id))
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

	maxUploadSize := 1048576
	contentLength, err := strconv.Atoi(r.Header.Get("Content-Length"))
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorLengthRequired)
		return
	} else if contentLength > maxUploadSize {
		ie.SendHTTPError(w, ie.ErrorRequestEntityTooLarge)
		return
	}

	// Read at most X MB of data from request body
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxUploadSize))

	b, err := io.ReadAll(r.Body)
	if err != nil {
		ie.SendHTTPError(w, ie.ErrorUndefined)
	}

	s := services.NewProgramService(db)
	revision, err := s.AddCodeRevision(r.Context(), services.AddCodeRevisionParams{
		ProgramUuid: programUUID,
		CreatedBy:   createdBy,
		Code:        b,
	})
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(revision)
}

// GetProgramCodeRevisionsDiff returns the difference between two code revisions for a program
func (ra *RestApi) GetProgramCodeRevisionsDiff(w http.ResponseWriter, r *http.Request, id rest.UuidParam, p rest.GetProgramCodeRevisionsDiffParams) {
	programUUID, err := uuid.Parse(string(id))
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
	diff, err := s.DiffProgramCodeAtRevisions(r.Context(), programUUID, p.RevA, p.RevB)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(diff))
}

// GetCodeFromProgram returns the newest, signed code for a program
func (ra *RestApi) GetCodeFromProgram(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	programUUID, err := uuid.Parse(string(id))
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
	code, err := s.GetSignedProgramCodeAtHead(r.Context(), programUUID)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(code))
}

// GetProgramCodeRevisions returns all code revisions for a program
func (ra *RestApi) GetProgramCodeRevisions(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	programUUID, err := uuid.Parse(string(id))
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
	revisions, err := s.FindAllCodeRevisions(r.Context(), programUUID)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(revisions)
}

// ExecuteProgramWebhook forwards a request to a webhook program
func (ra *RestApi) ExecuteProgramWebhook(w http.ResponseWriter, r *http.Request, id rest.UuidParam) {
	w.WriteHeader(http.StatusNotImplemented)
}

// SignProgramCodeRevisions signs a specific code revision
func (ra *RestApi) SignProgramCodeRevisions(w http.ResponseWriter, r *http.Request, id rest.UuidParam, revision int) {
	programUUID, err := uuid.Parse(string(id))
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
	signedBy, err := u.GetUserUuidFromToken(r.Context(), []byte(domaintoken.Token))

	s := services.NewProgramService(db)
	count, err := s.SignCodeRevision(r.Context(), services.SignCodeRevisionParams{
		ProgramUuid: programUUID,
		Revision:    revision,
		SignedBy:    signedBy,
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

// DeleteProgramCodeRevisions deletes a specific code revision
func (ra *RestApi) DeleteProgramCodeRevisions(w http.ResponseWriter, r *http.Request, id rest.UuidParam, revision int) {
	programUUID, err := uuid.Parse(string(id))
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
	count, err := s.DeleteProgramCodeRevision(r.Context(), programUUID, revision)
	if err != nil {
		ie.SendHTTPError(w, ie.ParseDBError(err))
		return
	} else if count == 0 {
		ie.SendHTTPError(w, ie.ErrorNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
