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
package library

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"regexp"
	// "strconv"
	"sync"
	"time"
)

var mux sync.RWMutex
var cache map[string]LibraryItem
var cacheTimeout time.Duration
var libraryIndex string

var maxAgeRe = regexp.MustCompile(`max-age=(\d*)`)

type LibraryItem struct {
	Key  string
	Code []byte

	lifetime time.Duration
	timer    *time.Timer
}

func (l *LibraryItem) Start() {
	l.Stop()
	l.timer = time.AfterFunc(l.lifetime, func() {
		mux.Lock()
		defer mux.Unlock()
		delete(cache, l.Key)
	})
}

func (l *LibraryItem) Stop() {
	if l.timer != nil {
		l.Stop()
	}
	l.timer = nil
}

type LibraryError struct {
	Code    int
	Message string
}

func (e *LibraryError) Error() string {
	return e.Message
}

func init() {
	cache = make(map[string]LibraryItem)
	cacheTimeout = 30 * time.Minute
	libraryIndex = "http://127.0.0.1:8000" // v1/library?domain=x&language=y&module=z&revision=r
}

func SetCacheTimeout(seconds int) {
	cacheTimeout = time.Duration(seconds) * time.Second
	mux.Lock()
	defer mux.Unlock()
	for _, v := range cache {
		v.Stop()
		v.lifetime = cacheTimeout
		v.Start()
	}
}

func SetIndexServer(uri string) {
	libraryIndex = uri
}

type LibraryParams struct {
	Domain   string
	Module   string
	Revision string
	Language string
}

func (l *LibraryParams) GetId() string {
	return fmt.Sprintf("%v:%v/%v@%v", l.Language, l.Domain, l.Module, l.Revision)
}

func (l *LibraryParams) QueryString() string {
	q := url.Values{}
	q.Set("domain", l.Domain)
	q.Set("module", l.Module)
	q.Set("language", l.Language)
	q.Set("revision", l.Revision)
	return q.Encode()
}

func Add(p *LibraryParams, content []byte) {
	mux.Lock()
	defer mux.Unlock()

	id := p.GetId()

	// If overwrite, ensure stop first
	if obj, ok := cache[id]; ok {
		obj.Stop()
	}

	l := LibraryItem{
		Key:      id,
		Code:     content,
		lifetime: cacheTimeout,
	}
	cache[id] = l

	l.Start() // Start bg work after adding element to the cache
}

func Get(p *LibraryParams) (*LibraryItem, error) {

	id := p.GetId()

	mux.RLock()
	lifetime := cacheTimeout
	v, ok := cache[id]
	mux.RUnlock()

	if ok == false {
		u, err := url.Parse(libraryIndex)
		if err != nil {
			return nil, err
		}
		u.Path = path.Join(u.Path, "/v1/library")
		u.RawQuery = p.QueryString()

		resp, err := http.Get(u.String())
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == 404 {
			return nil, &LibraryError{
				Code:    404,
				Message: "no such resource",
			}
		} else if resp.StatusCode != 200 {
			return nil, &LibraryError{
				Code:    502,
				Message: "invalid response from library index server",
			}
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		/*
			// Is this a good idea?
			// Using the response from the library server to dictate the cache time

			cache_control := resp.Header.Get("Cache-Control")
			for _, match := range maxAgeRe.FindAllStringSubmatch(cache_control, -1) {
				// max-age=XXXX
				if len(match) == 2 {
					ival, err := strconv.Atoi(match[1])
					if err != nil {
						continue
					}
					lifetime = time.Duration(ival) * time.Second
				}
			}
		*/

		v = LibraryItem{
			Key:      id,
			Code:     body,
			lifetime: lifetime,
		}
		mux.Lock()
		cache[id] = v
		mux.Unlock()

		v.Start() // Start the timer

		// return &v at end of func
	}

	return &v, nil
}

func Delete(p *LibraryParams) {
	mux.Lock()
	defer mux.Unlock()

	delete(cache, p.GetId())
}
