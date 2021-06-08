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
package selfpmgr

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/self-host/self-host/api/selfpmgr/worker"
)

type ProgramRevision struct {
	Domain      string
	Name        string
	ProgramUuid uuid.UUID
	Type        string
	Schedule    time.Duration
	Deadline    int32
	Language    string
	Revision    int32
	Code        []byte
	Checksum    string

	terminate chan struct{}
}

type WorkerTask struct {
	Language    string    `json:"language"`
	Deadline    int       `json:"deadline"`
	Domain      string    `json:"domain"`
	ProgramUuid uuid.UUID `json:"program_uuid"`
	SourceCode  string    `json:"source_code"`
}

func AtInterval(d time.Duration) <-chan time.Time {
	t := time.Now().Truncate(d).Add(d).Sub(time.Now())
	return time.After(t)
}

func NewProgramRevision(domain string, name string, program_uuid uuid.UUID, ptype string, schedule string,
	deadline int32, language string, revision int32, code []byte, checksum string) (*ProgramRevision, error) {
	p := &ProgramRevision{
		Domain:      domain,
		Name:        name,
		ProgramUuid: program_uuid,
		Type:        ptype,
		Deadline:    deadline,
		Language:    language,
		Revision:    revision,
		Code:        code,
		Checksum:    checksum,
		terminate:   make(chan struct{}),
	}

	sch, err := time.ParseDuration(schedule)
	if err != nil {
		return nil, err
	}
	p.Schedule = sch

	return p, nil
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

	go func() {
		for {
			select {
			case <-p.terminate:
				return
			case <-AtInterval(p.Schedule):
				err := p.Execute()
				if err != nil {
					logger.Error("unable to execute task", zap.Error(err))
				}
			}
		}
	}()
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

	host, err := worker.GetAvailable()
	if err != nil {
		return nil
	}

	resp, err := http.Post(host+"/v1/tasks", "application/json", bytes.NewBuffer(requestBody))
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

	// FIXME: log?
	close(p.terminate)
}

func (p *ProgramRevision) GetId() string {
	return fmt.Sprintf("%v/%v@%v", p.Domain, p.ProgramUuid.String(), p.Revision)
}
