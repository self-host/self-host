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
