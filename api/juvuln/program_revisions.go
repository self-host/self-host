// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package juvuln

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/self-host/self-host/pkg/workforce"
	"go.uber.org/zap"
)

var watch *cron.Cron

func init() {
	watch = cron.New(cron.WithSeconds())
	watch.Start()
}

type ProgramRevision struct {
	Domain      string
	Name        string
	ProgramUuid uuid.UUID
	Type        string
	Schedule    string
	Deadline    int32
	Language    string
	Revision    int32
	Code        []byte
	Checksum    string

	eid cron.EntryID
}

type WorkerTask struct {
	Language    string    `json:"language"`
	Deadline    int       `json:"deadline"`
	Domain      string    `json:"domain"`
	ProgramUuid uuid.UUID `json:"program_uuid"`
	SourceCode  string    `json:"source_code"`
}

func NewProgramRevision(domain string, name string, programUUID uuid.UUID, ptype string, schedule string,
	deadline int32, language string, revision int32, code []byte, checksum string) *ProgramRevision {
	return &ProgramRevision{
		Domain:      domain,
		Name:        name,
		ProgramUuid: programUUID,
		Type:        ptype,
		Schedule:    schedule,
		Deadline:    deadline,
		Language:    language,
		Revision:    revision,
		Code:        code,
		Checksum:    checksum,
	}
}

func (p *ProgramRevision) Equals(b *ProgramRevision) bool {
	return p.Domain == b.Domain &&
		p.Name == b.Name &&
		p.ProgramUuid == b.ProgramUuid &&
		p.Type == b.Type &&
		p.Schedule == b.Schedule &&
		p.Deadline == b.Deadline &&
		p.Language == b.Language &&
		bytes.Compare(p.Code, b.Code) == 0 &&
		p.Checksum == b.Checksum
}

func (p *ProgramRevision) Start() {
	if p.Type != "routine" {
		return
	}

	eid, err := watch.AddFunc(p.Schedule, p.Run)
	if err != nil {
		logger.Error("unable to schedule task", zap.Error(err))
	}
	p.eid = eid
}

func (p *ProgramRevision) Run() {
	err := p.Execute()
	if err != nil {
		logger.Error("unable to execute task", zap.Error(err))
	}
}

func (p *ProgramRevision) Execute() error {
	code := base64.StdEncoding.EncodeToString([]byte(p.Code))
	requestBody, err := json.Marshal(WorkerTask{
		ProgramUuid: p.ProgramUuid,
		Domain:      p.Domain,
		Language:    p.Language,
		Deadline:    int(p.Deadline),
		SourceCode:  code,
	})
	if err != nil {
		return err
	}

	w, err := workforce.GetAvailable()
	if err != nil {
		return nil
	}

	worker, ok := w.(*Worker)
	if ok == false {
		return fmt.Errorf("incorrect format for worker")
	}

	resp, err := http.Post(worker.URI+"/v1/tasks", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (p *ProgramRevision) Stop() {
	if p.Type != "routine" {
		return
	}

	watch.Remove(p.eid)
}

func (p *ProgramRevision) GetId() string {
	return fmt.Sprintf("%v/%v@%v", p.Domain, p.ProgramUuid.String(), p.Revision)
}
