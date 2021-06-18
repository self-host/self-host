// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package juvuln

import (
	"errors"
	"sync"
)

type ProgramCache struct {
	sync.RWMutex
	m map[string]*ProgramRevision

	deletes map[string]struct{}
}

func (p *ProgramCache) Add(obj *ProgramRevision) {
	p.Lock()
	defer p.Unlock()

	id := obj.GetId()

	// Remove id from deletion list
	delete(p.deletes, id)

	if a, ok := p.m[id]; ok {
		if a.Equals(obj) {
			// Already exists
			return
		} else {
			// Stop and delete current
			a.Stop()
			delete(p.m, id)
		}
	}

	// Start background routine
	obj.Start()

	// Add cache item
	p.m[id] = obj
}

func (p *ProgramCache) Begin() {
	// Mark all items for deletion
	p.Lock()
	defer p.Unlock()

	p.deletes = make(map[string]struct{})
	for k := range p.m {
		p.deletes[k] = struct{}{}
	}
}

func (p *ProgramCache) Commit() {
	// Remove all items still marked for deletion
	p.Lock()
	defer p.Unlock()

	for k := range p.deletes {
		if obj, ok := p.m[k]; ok {
			obj.Stop()
		}
		delete(p.m, k)
	}
	p.deletes = make(map[string]struct{})
}

func (p *ProgramCache) GetModule(domain, name string, revision int32) (*ProgramRevision, error) {
	p.RLock()
	defer p.RUnlock()

	var pr *ProgramRevision

	for _, item := range p.m {
		if revision == -1 &&
			item.Domain == domain &&
			item.Name == name &&
			(pr == nil || pr.Revision > item.Revision) {
			v := item
			pr = v
		} else if item.Domain == domain &&
			item.Name == name &&
			item.Revision == revision &&
			revision >= 0 {
			v := item
			pr = v
			break
		}
	}

	if pr == nil {
		return nil, errors.New("no such module")
	}

	return pr, nil
}
