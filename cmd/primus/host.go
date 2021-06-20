// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package primus

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	_ "github.com/jackc/pgx/v4/stdlib"
	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic("zap.NewProduction " + err.Error())
	}
}

type PrimusConfig struct {
	ConnInfo    string
	CreateQuery string
	ReadQuery   string
	UpdateQuery string
	DeleteQuery string
}

type PrimusHost struct {
	sync.RWMutex
	cfg       *PrimusConfig
	databases map[string]*sql.DB
	done      chan struct{}
	wg        sync.WaitGroup
	started   bool
}

func NewPrimusHost(cfg *PrimusConfig) *PrimusHost {
	return &PrimusHost{
		cfg:       cfg,
		databases: make(map[string]*sql.DB),
		done:      make(chan struct{}),
		started:   false,
	}
}

func (p *PrimusHost) Start() {
	if p.started == true {
		return
	}

	go func() {
		p.started = true
		p.wg.Add(1)
		select {

		case <-p.done:
			p.wg.Done()
			return
		}
	}()
}

func (p *PrimusHost) Stop() error {
	if p.started == false {
		return errors.New("not started")
	}

	close(p.done)

	p.wg.Wait()

	return nil
}

func (p *PrimusHost) Update(ctx context.Context) error {
	conn, err := sql.Open("pgx", p.cfg.ConnInfo)
	if err != nil {
		return err
	}

	defer func() {
		if err := conn.Close(); err != nil {
			logger.Error("DB error", zap.Error(err))
		}
	}()

	err = conn.Ping()
	if err != nil {
		return err
	}

	rows, err := conn.QueryContext(ctx, p.cfg.ReadQuery)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var domainName string
		var connectionInfo string
		if err := rows.Scan(&domainName, &connectionInfo); err != nil {
			return err
		}

		p.RLock()
		c, ok := p.databases[domainName]
		p.RUnlock()

		if ok {
			err = c.Ping()
			if err != nil {
				p.Lock()
				delete(p.databases, domainName)
				p.Unlock()
			}
		}

		conn, err := sql.Open("pgx", connectionInfo)
		if err != nil {
			p.Lock()
			p.databases[domainName] = nil
			p.Unlock()

			logger.Error("DB Error", zap.Error(err))
		} else {
			p.Lock()
			p.databases[domainName] = conn
			p.Unlock()

			logger.Info("db", zap.String("domainName", domainName), zap.String("connectionInfo", connectionInfo))
		}

		// FIXME: How do we remove databases no longer part of the index?
	}

	return nil
}

func (p *PrimusHost) GetDB(domain string) (*sql.DB, error) {
	p.RLock()
	res, ok := p.databases[domain]
	p.RUnlock()

	if ok == false {
		return nil, errors.New(fmt.Sprintf("no such domain '%v'", domain))
	}

	return res, nil
}
