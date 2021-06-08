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
package postgresql

//go:generate sqlc generate

import (
	"database/sql"
	"fmt"
	"errors"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

var (
	dbCache map[string]*sql.DB
	dbCacheMux sync.RWMutex
)

type DomainDB struct {
	Domain string
	DB *sql.DB
}

func init() {
	dbCache = make(map[string]*sql.DB)

	// Ensure connections every 30 seconds
	go func() {
		for {
			select {
			case <-time.After(30 * time.Second):
				failed := make([]string, 0)

				dbCacheMux.RLock()
				for domain, conn := range dbCache {
					err := conn.Ping()
					if err != nil {
						s := domain
						failed = append(failed, s)
					}
				}
				dbCacheMux.RUnlock()

				dbCacheMux.Lock()
				for _, domain := range failed {
					delete(dbCache, domain)
				}
				dbCacheMux.Unlock()
			break
			}
		}
	}()
}

func AddDB(domain, connectionInfo string) error {
	dbCacheMux.RLock()
	c, ok := dbCache[domain]
	dbCacheMux.RUnlock()

	if ok {
		err := c.Ping()
		if err != nil {
			dbCacheMux.Lock()
			delete(dbCache, domain)
			dbCacheMux.Unlock()
		}
		return nil
	}

	conn, err := sql.Open("pgx", connectionInfo)
	if err != nil {
		dbCacheMux.Lock()
		dbCache[domain] = nil
		dbCacheMux.Unlock()
		return err
	} else {
		err = conn.Ping()
		if err != nil {
			dbCacheMux.Lock()
			delete(dbCache, domain)
			dbCacheMux.Unlock()
			return err
		}
	}

	dbCacheMux.Lock()
	dbCache[domain] = conn
	dbCacheMux.Unlock()

	return nil
}

func GetDB(domain string) (*sql.DB, error) {
	dbCacheMux.RLock()
	res, ok := dbCache[domain]
	dbCacheMux.RUnlock()

	if ok == false {
		return nil, errors.New(fmt.Sprintf("no such domain '%v'", domain))
	}

	return res, nil
}

func GetAllDB() []DomainDB {
	dbs := make([]DomainDB, 0)

	dbCacheMux.RLock()
	for domain, db := range dbCache {
		dbs = append(dbs, DomainDB{
			Domain: domain,
			DB: db,
		})
	}
	dbCacheMux.RUnlock()

	return dbs
}
