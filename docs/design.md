# Design

The system is split up into several separate softwares.

- Database (PostgreSQL)
- Self-host API (selfserv)
- Program Manager (selfpmgr)
- Program Worker (selfpwrk)
- Primus (primus)
- CLI tool (selfctl)

[design drawing]

... explain the drawing

# Database

The only supported database is PostgreSQL version 12 or newer.


# Self-host API

The main component of the system. It exposed a `domain` database as a set of API endpoints.

For details about the API specification see: [openapiv3.yaml](https://github.com/self-host/self-host/blob/master/api/selfserv/rest/openapiv3.yaml) file.

Key components of the Self-host API are;

- `Users`: A `User` account is required to access the API.
- `Groups`: A `User` can belong to one or several different `Groups`.
- `Policies`: A `Policy` is applied to a `Group` and grants; CREATE, READ, UPDATE or DELETE permission.
- `Timeseries`: A series of data-points representing a single entity.
- `Things`: A `Thing` is an object used as an (optional) way to group `Timeseries` into logical structures.
- `Datasets`: A `Dataset` can be used for multiple different purposes. It is a good place to store complex data-series that can't be represented by `Timeseries`.
- `Programs`: A `Program` is an easy way to deploy trivial code to handle repeated tasks.


# Primus (Work in Progress)

This database contains (probably) only a single table where domains are declared (with PostgreSQL URI).

Whenever an instance of the API server starts up, it queries the Primus database to establish the available domains. Whenever a new domain is added, the Primus database will send out a NOTIFY event so that all instances of the API server can update their config.

The Self-host should only make few assumptions (demands) on the structure of the Primus database.

From configuration (file, env, etc) the Self-host should establish a few items;

- The connection info to connect to the Primus database
- A configurable SQL query to create a new domain.
- A configurable SQL query to read all (or a subset) of domains.
- A configurable SQL query to update the domain info for a domain.
- A configurable SQL query to delete a domain.

The return values from these queries must always be defined for the Golang code to interpret the result. The code should of-course come with excelent examples and SQL code to create a table if none exists. Maybe this should be part of the `selfctl` cli program?

