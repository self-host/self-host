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

package services_test

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/ory/dockertest/v3"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	pgVersion = "12"
	dbname    = "selfhost_services_test"
	rootToken = "root"
)

var db *sql.DB

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run("postgres", pgVersion, []string{"POSTGRES_PASSWORD=secret"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	} else {
		resource.Expire(60) // Tell docker to hard kill the container in 60 seconds
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		var err error
		db, err = sql.Open("postgres", fmt.Sprintf("postgres://postgres:secret@localhost:%s/postgres?sslmode=disable", resource.GetPort("5432/tcp")))
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE \"%v\" WITH ENCODING 'UTF-8'", dbname))
	if err != nil {
		log.Fatalf("Could not create database: %s", err)
	}

	db.Close()

	pgUrl := fmt.Sprintf("postgres://postgres:secret@localhost:%s/%s?sslmode=disable", resource.GetPort("5432/tcp"), dbname)

	db, err = sql.Open("postgres", pgUrl)
	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	mig, err := migrate.New("file://../../postgres/migrations", pgUrl)
	if err != nil {
		log.Fatal(err)
	}
	err = mig.Up()
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}
