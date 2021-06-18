// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package workforce

import (
	"fmt"
	"sort"
	"sync"
)

type Worker interface {
	Alive() bool
	SetLoad(uint64)
	GetLoad() uint64
}

type ByLoad []Worker

func (l ByLoad) Len() int {
	return len(l)
}
func (l ByLoad) Less(i, j int) bool {
	return l[i].GetLoad() < l[j].GetLoad()
}
func (l ByLoad) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

type Workforce struct {
	sync.RWMutex
	workers map[string]Worker
}

var (
	wforce *Workforce
)

func init() {
	wforce = NewWorkforce()
}

func NewWorkforce() *Workforce {
	return &Workforce{
		workers: make(map[string]Worker),
	}
}

func (c *Workforce) Add(k string, w Worker) {
	c.Lock()
	defer c.Unlock()

	c.workers[k] = w
}

func (c *Workforce) Delete(k string) {
	c.Lock()
	defer c.Unlock()

	delete(c.workers, k)
}

func (c *Workforce) Clear() {
	c.Lock()
	defer c.Unlock()

	c.workers = make(map[string]Worker)
}

func (c *Workforce) Exists(id string) bool {
	c.RLock()
	defer c.RUnlock()
	_, ok := c.workers[id]
	return ok
}

func (c *Workforce) GetAvailable() (Worker, error) {
	c.RLock()
	defer c.RUnlock()

	workers := make([]Worker, 0)

	for _, worker := range c.workers {
		if worker.Alive() {
			w := worker
			workers = append(workers, w)
		}
	}

	if len(workers) > 0 {
		sort.Sort(ByLoad(workers))
		return workers[0], nil
	}

	return nil, fmt.Errorf("no available worker")
}

func (c *Workforce) SetLoad(id string, l uint64) {
	c.Lock()
	defer c.Unlock()

	for k, worker := range c.workers {
		if k == id && worker != nil {
			worker.SetLoad(l)
		}
	}
}

func (c *Workforce) ClearInactive() []Worker {
	c.Lock()
	defer c.Unlock()

	d := make([]Worker, 0)

	for k, worker := range c.workers {
		if worker.Alive() == false {
			d = append(d, worker)
			delete(c.workers, k)
		}
	}

	return d
}

/*
 * Global functions
 */
func Add(k string, w Worker) {
	wforce.Add(k, w)
}

func Delete(k string) {
	wforce.Delete(k)
}

func Clear() {
	wforce.Clear()
}

func GetAvailable() (Worker, error) {
	return wforce.GetAvailable()
}

func Exists(id string) bool {
	return wforce.Exists(id)
}

func SetLoad(id string, l uint64) {
	wforce.SetLoad(id, l)
}

func ClearInactive() []Worker {
	return wforce.ClearInactive()
}
