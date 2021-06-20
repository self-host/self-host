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
[![Docker](https://github.com/self-host/self-host/actions/workflows/docker.yml/badge.svg)](https://github.com/self-host/self-host/actions/workflows/docker.yml)

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
    + You can host domains on separate PostgreSQL machines.
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

- One (or more) instances of the `Self-host API server`. This server provides the public REST API and is the only exposed surface from the internet.
- One instance of the `Program Manager`. This server manages all programs and schedules them.
- One (or more) instances of the `Program Worker`. This server accepts work from the `Program Manager`.
- One (or more) DBMS to host all Self-host databases (Domains). PostgreSQL 12+ with one (or more) Self-host `Domains`.

The `API server` can accept client request from either the internet or from the intranet. The `Domain databases` backs all `API server` requests.

The `Program Worker` can execute requests to external services on the internet or internal services, for example, the `API server`.

An `HTTP Proxy` may be used in front of the `API server` depending on the deployment scenario.


# Project structure

- `api`:
    + `aapije`: REST API interface for the Self-host public facing API server.
    + `juvuln`: REST API interface for the internal API of the Program Manager.
    + `malgomaj`: REST API interface for the internal API of the Program Worker.
- `cmd`:
    + `selfctl`: Self Control; self-host CLI program.
    + `aapije`: Aapije is the Self-host public facing API server.
    + `juvuln`: Juvuln is the Self-host Program Manager.
    + `malgomaj`: Malgomaj is the Self-host Program Worker.
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

- Good knowledge of Docker.
- Some knowledge of PostgreSQL.
- Some knowledge of GNU+Linux or Unix environments in general.

Hardware and software required;

- Computer with Docker installed.

[Five to fifteen-minute deployment](https://github.com/self-host/self-host/blob/main/docs/test_deployment.md)


# Documentation

- [Glossary](https://github.com/self-host/self-host/blob/main/docs/glossary.md)
- [Tools of the trade](https://github.com/self-host/self-host/blob/main/docs/tools_of_the_trade.md)
- [Design](https://github.com/self-host/self-host/blob/main/docs/design.md)
    + [Authentication](https://github.com/self-host/self-host/blob/main/docs/authentication.md)
    + [Access control](https://github.com/self-host/self-host/blob/main/docs/access_control.md)
    + [Data partitioning](https://github.com/self-host/self-host/blob/main/docs/data_partitioning.md)
    + [Rate control](https://github.com/self-host/self-host/blob/main/docs/rate_control.md)
    + [Program Manager and Workers](https://github.com/self-host/self-host/blob/main/docs/program_manager_worker.md)
    + [Unit handling](https://github.com/self-host/self-host/blob/main/docs/unit_handling.md)
    + [Exernal services](https://github.com/self-host/self-host/blob/main/docs/external_services.md)
- [Benchmark](https://github.com/self-host/self-host/blob/main/docs/benchmark_overview.md)
- [Public-facing API specification](https://petstore.swagger.io/?url=https://raw.githubusercontent.com/self-host/self-host/main/api/aapije/rest/openapiv3.yaml)
- [Notes on deploying to production](https://github.com/self-host/self-host/blob/main/docs/production_deployment.md)
    + [Docker](https://github.com/self-host/self-host/blob/main/docs/docker_deployment.md)
    + [Kubernetes](https://github.com/self-host/self-host/blob/main/docs/k8s_deployment.md)


:hearts: Like this project? Want to improve it but unsure where to begin? Check out the [issue tracker](https://github.com/self-host/self-host/issues).


[vimpel]: https://raw.githubusercontent.com/self-host/self-host/main/docs/assets/vimpel.svg "Vimpel"
[fig1]: https://raw.githubusercontent.com/self-host/self-host/main/docs/assets/overview.svg "Overview"
