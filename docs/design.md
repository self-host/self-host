# Design

The system exists as a set of interconnected parts.

- Database (PostgreSQL)
- Self-host API (aapije)
- Program Manager (juvuln)
- Program Worker (malgomaj)

![Overview][fig1]


# Database

The only supported database is PostgreSQL version 12 or newer.


# Self-host API

The Self-host API is the main component of the system. It exposed a `domain` database as a set of API endpoints.

For details about the API specification see: [openapiv3.yaml](https://github.com/self-host/self-host/blob/master/api/aapije/rest/openapiv3.yaml) file.

Key components of the Self-host API are;

- `Users`: A `User` account is required to access the API.
- `Groups`: A `User` can belong to one or several different `Groups`.
- `Policies`: A `Policy` applies to a `Group` and grants; CREATE, READ, UPDATE or DELETE permission.
- `Timeseries`: A series of data points representing a single entity.
- `Things`: A `Thing` is an object used as an (optional) way to group `Timeseries` into logical structures.
- `Datasets`: We can use a `Dataset` for multiple different purposes. It is an excellent place to store complex data series that we can't represent with `Timeseries`.
- `Programs`: A `Program` is an easy way to deploy trivial code to handle repetitive tasks.


# Program Manager

The purpose of the Program Manager is to keep track of Program Workers and distribute work in the form of `programs`.

For more details, see [Program Manager and Worker](https://github.com/self-host/self-host/blob/main/docs/program_manager_worker.md).


# Program Worker

The purpose of the Program Worker is to execute tasks given to it by the Program Manager. A Worker may implement all or a subset of `languages` depending on implementation limitations.

For more details, see [Program Manager and Worker](https://github.com/self-host/self-host/blob/main/docs/program_manager_worker.md).


[fig1]: https://raw.githubusercontent.com/self-host/self-host/main/docs/assets/overview.svg "Overview"