// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package services

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
