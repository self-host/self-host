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
