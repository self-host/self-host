// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package juvuln

import (
	"sync"
	"time"
)

type Worker struct {
	sync.Mutex

	Id        string
	URI       string
	Languages []string

	load     uint64
	timeout  time.Duration
	lastSeen time.Time
}

func (w *Worker) SetLoad(l uint64) {
	w.Lock()
	defer w.Unlock()
	w.load = l
	w.lastSeen = time.Now()
}

func (w *Worker) GetLoad() uint64 {
	w.Lock()
	defer w.Unlock()
	return w.load
}

func (w *Worker) Alive() bool {
	w.Lock()
	defer w.Unlock()
	return time.Now().Before(w.lastSeen.Add(w.timeout))
}

func NewWorker(id, uri string, langs []string, timeout time.Duration) *Worker {
	return &Worker{
		Id:        id,
		URI:       uri,
		Languages: langs,
		timeout:   timeout,
		lastSeen:  time.Now(),
	}
}
