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
package worker

import (
	"errors"
	"sync"
)

type WorkerCache struct {
	sync.RWMutex
	workers map[string]*Worker
}

type Worker struct {
	URI       string
	Languages []string
}

var wcache *WorkerCache

func init() {
	wcache = NewWorkerCache()
}

func NewWorkerCache() *WorkerCache {
	return &WorkerCache{
		workers: make(map[string]*Worker),
	}
}

func NewWorker(uri string, langs []string) *Worker {
	return &Worker{
		URI:       uri,
		Languages: langs,
	}
}

func (c *WorkerCache) Add(k string, w *Worker) {
	c.Lock()
	defer c.Unlock()

	c.workers[k] = w
}

func (c *WorkerCache) Delete(k string) {
	c.Lock()
	defer c.Unlock()

	delete(c.workers, k)
}

func (c *WorkerCache) Clear() {
	c.Lock()
	defer c.Unlock()

	c.workers = make(map[string]*Worker)
}

func (c *WorkerCache) Exists(id string) bool {
	c.RLock()
	defer c.RUnlock()
	_, ok := c.workers[id]
	return ok
}

func (c *WorkerCache) GetAvailable() (string, error) {
	c.RLock()
	defer c.RUnlock()

	// FIXME: pick the most suitable worker
	for _, worker := range c.workers {
		return worker.URI, nil
	}

	return "", errors.New("no available worker")
}

/*
 * Global functions
 */
func Add(k string, w *Worker) {
	wcache.Add(k, w)
}

func Delete(k string) {
	wcache.Delete(k)
}

func Clear() {
	wcache.Clear()
}

func GetAvailable() (string, error) {
	return wcache.GetAvailable()
}

func Exists(id string) bool {
	return wcache.Exists(id)
}
