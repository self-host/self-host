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

package postgres

//go:generate sqlc generate

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"go.uber.org/zap"
)

var (
	logger     *zap.Logger
	dbCache    map[string]*DBConnection
	dbCacheMux sync.RWMutex
)

type DBConnection struct {
	C                *sql.DB
	Domain           string
	ConnectionString string

	quit chan struct{}
}

func (c *DBConnection) Equals(d *DBConnection) bool {
	return c.ConnectionString == d.ConnectionString
}

func (c *DBConnection) Connect() {
	if c.quit != nil {
		return
	}

	c.quit = make(chan struct{})

	go func() {
		for {
			// No connection
			if c.C == nil {
				conn, err := sql.Open("pgx", c.ConnectionString)
				if err != nil {
					logger.Error("unable to connect to DB", zap.String("domain", c.Domain), zap.Error(err))
				}
				c.C = conn
			}

			select {
			case <-time.After(30 * time.Second):
				// Active connection (maybe)
				if c.C != nil {
					err := PingDB(c.C)
					if err != nil {
						// Unable to ping the DB, set the DB handle to nil
						c.C = nil

						// Report error
						logger.Error("unable to access DB", zap.String("domain", c.Domain), zap.Error(err))
					}
				}
			case <-c.quit:
				return
			}
		}
	}()
}

func (c *DBConnection) Close() {
	if c.quit != nil {
		close(c.quit)
		c.quit = nil
	}
}

type DomainDB struct {
	Domain string
	DB     *sql.DB
}

func init() {
	var err error

	logger, err = zap.NewProduction()
	if err != nil {
		panic("zap.NewProduction " + err.Error())
	}

	dbCache = make(map[string]*DBConnection)
}

func AddDB(domain, connectionInfo string) error {
	dbCacheMux.RLock()
	c, found := dbCache[domain]
	dbCacheMux.RUnlock()

	newDBC := &DBConnection{
		Domain:           domain,
		ConnectionString: connectionInfo,
	}

	if found {
		if c.Equals(newDBC) {
			// Already exists
			return nil
		} else {
			// Same domain, but new DB url
			// Close existing background worker
			c.Close()
		}
	}

	// Add to the domain cache (no matter if we can connect or not!)
	dbCacheMux.Lock()
	dbCache[domain] = newDBC
	dbCacheMux.Unlock()

	// Let the DBC sort itself out
	newDBC.Connect()

	return nil
}

func RemoveDB(domain string) error {
	dbCacheMux.RLock()
	c, found := dbCache[domain]
	dbCacheMux.RUnlock()

	if found == false {
		return fmt.Errorf("no such domain '%v'", domain)
	}

	c.Close() // Stop the background worker

	dbCacheMux.Lock()
	delete(dbCache, domain)
	dbCacheMux.Unlock()

	return nil
}

func PingDB(conn *sql.DB) error {
	if conn == nil {
		return fmt.Errorf("DB connection pointer is nil")
	}

	err := conn.Ping()
	if err != nil {
		return err
	}

	return nil
}

func PingDomain(domain string) error {
	conn, err := GetDB(domain)
	if err != nil {
		return err
	}

	err = conn.Ping()
	if err != nil {
		return err
	}

	return nil
}

func GetDB(domain string) (*sql.DB, error) {
	dbCacheMux.RLock()
	res, ok := dbCache[domain]
	dbCacheMux.RUnlock()

	if ok == false {
		return nil, fmt.Errorf("no such domain '%v'", domain)
	}

	if res.C == nil {
		return nil, fmt.Errorf("no active connection to domain '%v'", domain)
	}

	return res.C, nil
}

func GetAllDB() []DomainDB {
	dbCacheMux.RLock()
	dbs := make([]DomainDB, 0)
	for domain, db := range dbCache {
		dbs = append(dbs, DomainDB{
			Domain: domain,
			DB:     db.C,
		})
	}
	dbCacheMux.RUnlock()

	return dbs
}

func GetDomains() []string {
	dbCacheMux.RLock()

	domains := make([]string, 0)
	for domain := range dbCache {
		domains = append(domains, domain)
	}
	dbCacheMux.RUnlock()

	return domains
}
