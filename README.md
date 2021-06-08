<p align="center">
    <img src="https://raw.githubusercontent.com/self-host/self-host/main/docs/assets/logo.svg" width="200" height="200">
    <br>
    <br>
    <quote>&ldquo;If you want something done, do it yourself.&rdquo;</quote>
    <br>
    <i>- A lot of people</i>
</p>

---

[![GPLv3 License](https://img.shields.io/badge/license-GPLv3-blue)](https://github.com/self-host/self-host/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/self-host/self-host)](https://goreportcard.com/report/github.com/self-host/self-host)
[![Docker](https://img.shields.io/docker/cloud/build/selfhoster/selfserv?label=selfserv&style=flat)](https://hub.docker.com/r/selfhoster/selfserv/builds)
[![Docker](https://img.shields.io/docker/cloud/build/selfhoster/selfpmgr?label=selfpmgr&style=flat)](https://hub.docker.com/r/selfhoster/selfpmgr/builds)
[![Docker](https://img.shields.io/docker/cloud/build/selfhoster/selfpwrk?label=selfpwrk&style=flat)](https://hub.docker.com/r/selfhoster/selfpwrk/builds)

# Self-host by NODA

Roll your own NODA API compatible server and host the data yourself.

Made in Sweden :sweden: by skilled artisan software engineers.


# The purpose of this project

To provide clients and potential clients with an alternative platform where they are responsible for hosting the API server and the data store.


# Features

- Small infrastructure.
    + The only dependency is on PostgreSQL.
- Compiled binaries (Go).
- Designed to scale with demand.
    + Domain can be hosted on separate PostgreSQL DBMS.
    + You can scale the Self-host API to as many instances as you require.
    + The Program Manager and its Workers can be scaled as needed.
- Can be hosted in your environment.
    + No requirement on any specific cloud solution.
- It is Free software.
    + License is GPLv3.
    + The API specification is open, and you can implement it if you don't want to use the existing implementation.

# Overview

A typical deployment scenario would look something like this;

![Overview][fig1]

- One or several instances of the Self-host API server.
- One instance of the Program Manager.
- One or server instances of the Program Worker.
- One DBMS to host all Self-host databases (Domains).


# Project structure

- `api`:
    + `selfserv`: REST API interface for the Self-host central API server.
    + `selfpmgr`: REST API interface for the Program Manager.
    + `selfpwrk`: REST API interface for the Program Worker.
- `cmd`:
    + `selfctl`: Self Control; self-host CLI program.
    + `selfserv`: Self Server; self-host main API server.
    + `selfpmgr`: Self Program Manager; self-host Program Manager.
    + `selfpwrk`: Self Program Worker; self-host Program Worker.
- `docs`: Documentation
- `internal`:
    + `services`: Handlers for interfaces to the PostgreSQL backend.
    + `errors`: Custom errors.
- `middleware`: Middleware used by HTTP Servers.
- `postgres`:
    + `migrations`: Database schema.
    + `queries`: Database queries.


# Five to fifteen-minute deployment

Skills required;

- Good knowledge of Docker
- Some knowledge of PostgreSQL
- Some knowledge of GNU+Linux or Unix environments in general.

Hardware and software required;

- Computer with Docker installed

### 1) Deploy the DBMS

*Estimated time required: 1-3 minutes*

Create a separate user-defined network. Calling it 'selfhost'.

> docker network create -d bridge selfhost

Start a container with PostgreSQL 12 or PostgreSQL 13.

> docker run --name pg13 --network selfhost -e POSTGRES_PASSWORD=mysecretpassword -d postgres:13-alpine

Create a new database on the PostgreSQL DBMS.

> docker run -it --rm --network selfhost -e PGPASSWORD=mysecretpassword postgres:13-alpine psql -h pg13 -U postgres

At the prompt;

> CREATE DATABASE "selfhost-test" WITH ENCODING 'UTF-8';

Then type '\q' to close the connection.

### 2) Deploy the DB schema (Manually without selfctl)

*Estimated time required: 1-3 minutes*

Download the source code from; https://github.com/self-host/self-host

The folder we are interested in is "postgres/migrations".

> docker run -v {{ migration_dir }}:/migrations --network selfhost migrate/migrate -path=/migrations/ -database postgresql://postgres:mysecretpassword@pg13.selfhost:5432/selfhost-test?sslmode=disable up

The database is now ready.

### 3) Deploy the Self-host API (selfserv)

*Estimated time required: 1-3 minutes*

Choose a good place in your filesystem to store configuration files, as we will have to mount several files.

Save the following file as `selfserv.conf.yaml`.

```yaml
listen:
  host: 0.0.0.0
  port: 80

domainfile: domains.yaml
```

Also, save the following file as `domains.yaml`
```yaml
domains:
  test: postgresql://postgres:mysecretpassword@pg13.selfhost:5432/selfhost-test
```

Start a container instance of the `selfserv` program.

> docker run --name selfserv --network selfhost -p 127.0.0.1:8080:80 -e CONFIG_FILENAME=selfserv.conf.yaml -v {{ config_dir }}:/etc/selfhost -d selfhoster/selfserv:main

### 4) Deploy the Self-host Program Manager (selfpmgr)

*Estimated time required: 1-3 minutes*

We are using the same configuration location as for step 3. Save the following file as `selfpmgr.conf.yaml`.

```yaml
listen:
  host: 0.0.0.0
  port: 80

domainfile: domains.yaml
```

It is OK in this test scenario to share the same `domains.yaml` file.

Then deploy the container instance.

> docker run --name selfpmgr --network selfhost -e CONFIG_FILENAME=selfpmgr.conf.yaml -v {{ config_dir }}:/etc/selfhost -d selfhoster/selfpmgr:main

### 5) Deploy the Self-host Program Worker (selfpwrk)

*Estimated time required: 1-3 minutes*

We are using the same configuration location as for step 3 and 4. Save the following file as `selfpwrk.conf.yaml`.

```yaml
listen:
  host: 0.0.0.0
  port: 80

cache:
  library_timeout: 3600
  program_timeout: 600

program_manager:
  scheme: http
  authority: selfpmgr.selfhost:80

module_library:
  scheme: http
  authority: selfpmgr.selfhost:80
```

Then deploy the container instance.

> docker run --name selfpwrk --network selfhost -e CONFIG_FILENAME=selfpwrk.conf.yaml -v {{ config_dir }}:/etc/selfhost -d selfhoster/selfpwrk:main

### 6) Done

The system is now deployed in a test environment.

The default secret key is "root" and belongs to the "root" user.

The API server listens for connections on port 8080 on localhost.

Using a browser, visit; `http://127.0.0.1:8080/static/swagger-ui/`, and the API documentation page should greet you.

When authenticating, use "test" as the username and "root" as the password.

# Documentation

- [Glossary](docs/glossary.md)
- [Tools of the Trade](docs/tools_of_the_trade.md)
- [Design](docs/design.md)
- [Resource Requirements](docs/resource_requirements.md)
    + Rate Control: It's own section with explanation and pitfalls. golang.org/x/time/rate
- [Performance Test](docs/performance_test.md)


:hearts: Like this project? Want to improve it but unsure where to begin? Check out the [issue tracker](https://github.com/self-host/self-host/issues).


[fig1]: https://raw.githubusercontent.com/self-host/self-host/main/docs/assets/overview.svg "Overview"
