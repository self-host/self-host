// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package malgomaj

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Program interface {
	Equals(b Program) bool
	Compile(context.Context) error
	Run(context.Context) error
	RunWithHTTP(context.Context, http.ResponseWriter, *http.Request) error
	Modules() []string
	Language() string

	Id() string
	Deadline() time.Duration
	Checksum() [16]byte // MD5
}

type ProgramCacheItem struct {
	program Program
	expires time.Time
	timer   *time.Timer
}

func (p *ProgramCacheItem) Expires() time.Time {
	return p.expires
}

func (p *ProgramCacheItem) Start() {
	p.Stop()
	p.timer = time.AfterFunc(time.Until(p.expires), func() {
		// Access the "global" cache
		programCache.mux.Lock()
		defer programCache.mux.Unlock()
		delete(programCache.m, p.GetId())
	})
}

func (p *ProgramCacheItem) Stop() {
	if p.timer != nil {
		p.Stop()
	}
	p.timer = nil
}

func (p *ProgramCacheItem) GetId() string {
	return p.program.Id()
}

type ProgramCache struct {
	timeout time.Duration
	m       map[string]*ProgramCacheItem
	mux     sync.RWMutex
}

var (
	programCache ProgramCache
)

func init() {
	programCache = ProgramCache{
		timeout: 0 * time.Second,
		m:       make(map[string]*ProgramCacheItem),
	}
}

func SetCacheTimeout(seconds int) {
	programCache.mux.RLock()
	defer programCache.mux.RUnlock()
	programCache.timeout = time.Duration(seconds) * time.Second
}

func ProgramCacheAdd(p Program) *ProgramCacheItem {
	programCache.mux.Lock()
	defer programCache.mux.Unlock()
	item := ProgramCacheItem{
		program: p,
		expires: time.Now().Add(programCache.timeout),
	}
	if o, ok := programCache.m[p.Id()]; ok {
		o.Stop()
	}
	programCache.m[p.Id()] = &item
	item.Start()
	return &item
}

func ProgramCacheGet(id string) *ProgramCacheItem {
	p, ok := programCache.m[id]
	if ok == false {
		return nil
	}
	return p
}

func ProgramCacheGetLoad() int64 {
	var load int64
	programCache.mux.Lock()
	defer programCache.mux.Unlock()
	for _, obj := range programCache.m {
		load += int64(obj.program.Deadline() / time.Millisecond)
	}
	return load
}

func NewProgram(domain string, id string, language string, deadline time.Duration, sourceCode []byte) (Program, error) {
	switch language {
	case "tengo":
		return NewTengoProgram(domain, id, deadline, sourceCode), nil
	}

	return nil, errors.New(fmt.Sprintf("language %v is not supported", language))
}
