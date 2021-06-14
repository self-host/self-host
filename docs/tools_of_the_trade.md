# Tools of the Trade

These days software is seldom written in a vacuum. Most projects have dependencies on operating systems, external libraries and build environments.

This section explains the different tools and libraries used in this project, why they were selected, and their purpose.


# OpenAPI v3

To ease the effort required when developing a server or client compatible with the API specification, we have decided to use OpenAPI as a declarative way to establish the API. This way, anyone can use a Swagger client to generate skeleton code for their project.

See Swagger Codegen at [Swagger.io](https://swagger.io/tools/swagger-codegen/) for details.

## Questions and Answers

#### Why v3? Why not v2?
We choose OpenAPI v3 (released in 2017) over v2 (released in 2014), as the structure is easier to work with compared to v2. Knowing fully well that there are fewer alternatives when it comes to generating the server code. We still think that v3 is such an improvement over v2 that this limitation is worth it.


# Go (the language)

One can always argue for or against any choice of language. Using Go for an API server has several benefits;

- The language is type-safe. Type safety is a significant improvement over languages such as Python.
- Does (fast) Garbage Collection.
- Testing is part of the main Go set of packages.
- Compiles into a single binary. No need to include X number of dependency packages along with the binary. Looking at you Python and NodeJS.
- Extremely easy to do cross-compilation for other Operating Systems and other target architectures.

## Questions and Answers

#### Why not use Rust? Or Clojure, or Haskell, or Julia, or C# or *X*?
We are not experienced enough in these languages. Each language has its benefits and drawbacks. Go is well suited for an API server.

#### We don't understand Go, and it does not fit well into our current stack.
The specification for the API is open. Feel free to implement your API server using any language you want. Go is the chosen language for this software.


# PostgreSQL 12+

PostgreSQL is an ACID-compliant Open Source SQL relational database management system.

We have used PostgreSQL in production for over ten years. While some corner cases may throw some DevOps people for a loop, it is still a good choice if you want a Relational Database.

Some of the reasons why we prefer PostgreSQL over the alternatives.

- Transactional DDL
- Common Table Expressions
- Declarative Partitioning
- JSONB type with operators
- NOTIFY events

## Questions and Answers
#### Why not Oracle or SQL Server?
As this project is Free Software, we wanted all system components to be Free Software / Open Source.

#### Why not MySQL or MariaDB?
Personal taste. See the list above. Also, we have better knowledge and experience with PostgreSQL.


# Tengo

We need a way to manage server-side code execution of minor tasks or as a way to interact with external service. We have chosen to build a solution around the Tengo Virtual Machine to do so. Tengo is an embeddable, small, dynamic, fast and secure script language for Go. It is relatively easy to extend the VM with modules written in Go, such as an HTTP module or an mqtt-publish module.

Each script (or program) can be extended via other scripts (libraries) and should only run for a small amount of time. Programs are helpful to communicate data at certain time intervals.

[https://github.com/d5/tengo](https://github.com/d5/tengo)


# Docker (For building a container)

Docker is a set of platform as a service (PaaS) products that use OS-level virtualization to deliver software in packages called containers.

https://www.docker.com


# Kubernetes

Kubernetes is an open-source container-orchestration system for automating computer application deployment, scaling, and management.

https://kubernetes.io/


# Go(lang) tools and utilities

## OAPI Codegen

OAPI Codegen is a set of utilities for generating Go boilerplate code for services based on OpenAPI 3.0 API definitions.

[https://github.com/deepmap/oapi-codegen](https://github.com/deepmap/oapi-codegen)

## KIN OpenAPI

A project to handle OpenAPI files. Can generate both OpenAPI JSON and YAML files from Go code.

[https://github.com/getkin/kin-openapi](https://github.com/getkin/kin-openapi)

## CORS net/http middleware

Wthout this, no (current) Web Browser can interact with the Self-host API as a separate service.

For details about CORS [read](https://en.wikipedia.org/wiki/Cross-origin_resource_sharing) Wikipedia.

https://github.com/go-chi/cors

## go-units (For Unit/Dimenssion conversion)

It is an essential requirement of this project to handle conversion between different SI units properly. Each stored value needs a unit, and while the API server internally will manage those units. One can not expect external services to use the same units at all times. Therefore conversion support must exist.

One initial idéa was to use the PostgreSQL extension [postgresql-unit](https://github.com/df7cb/postgresql-unit). Which takes care of units on a database level. However, due to the lack of support for this extension among major cloud platform providers, another solution had to be selected.

Initially the Unit package of Gonum was concidred. However, due to the way Gonum Unit expects the code to be used (no look-up of units), it can not in its current state be used in this project.

As an alternative we decided on using go-units instead.

- Go-units allows us to look up units using `Find(string)`.
- The library allows us to convert between to units.
- Conversion will result in a processing overhead.

For more details about the library;

- https://github.com/bcicen/go-units

## migrate: Database migrations written in Go

When you have a relational database, you always need a way to manage schema updates. We choose to `migrate` because it does what we need it to do, and it is supported by `sqlc`.

https://github.com/golang-migrate/migrate

## sqlc: A SQL Compiler for Go

The tool `sqlc` generates type-safe Go code (models) from SQL definitions. It is compatible with migrate.

https://github.com/kyleconroy/sqlc

## Viper: Go configuration with fangs

An excelent Go library to handle configuration.

https://github.com/spf13/viper

## Dockertest

Use Docker to run your Go language integration tests against third party services on Microsoft Windows, Mac OSX and Linux!

https://github.com/ory/dockertest