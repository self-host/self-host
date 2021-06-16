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

package selfpmgr

type Worker struct {
	URI       string
	Languages []string
	load      uint64
}

func (w *Worker) SetLoad(l uint64) {
	w.load = l
}

func (w *Worker) GetLoad() uint64 {
	return w.load
}

func NewWorker(uri string, langs []string) *Worker {
	return &Worker{
		URI:       uri,
		Languages: langs,
	}
}
