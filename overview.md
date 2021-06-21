% Self-host: an Overview
% Mikael Ganehag Brorsson @ganehag
% 2021-06-21

### What is the Self-host

The Self-host is;

- A complete environment to run the Self-host API.
- A time-series store and dataset store.
- An open API interface specification.
- Free and open source [software](https://github.com/self-host/self-host).
- Not necessarily the only solution.
- Spartacus.

---

### Why does it exist?

We at Noda want to provide an alternative to our hosted solution as we are experiencing a growing need for customers to be in complete control. As a way to give that level of control, we have come up with the Self-host solution.

As a way to ensure the future of the solution, we decided early on that it should be available to everyone as [free and open source software](#freelibre-software-and-open-source).

---

### The target audience

Medium to large organization with the requirement that solutions should run under their banner and not in someones else's environment.

Examples are;

- Utility companies
- Property owners
- Research institutes
- Integrators

# Components of the system

## Overview

![][fig1]

## Details

- One (or more) instances of the **Self-host API server**. 
- One instance of the **Program Manager**. 
- One (or more) instances of the **Program Worker**.
- One (or more) DBMS to host all Self-host databases (Domains).

---

### The API server (Aapije)

The server accepts client request from either the internet or from the *intranet*. Exposes an interface to the data in the **Domain databases**.

---

### The Program Manager (Juvuln)

The program manager tracks all programs from all domains and submits program execution tasks to instances of the Program Worker.

---

### The Program Worker (Malgomaj) 

Executes program code to perform various tasks like requests to external services on the *internet* or *intranet*, for example, the API server.

---

### Notes

- An **HTTP Proxy** may/should be used in front of the **API server** depending on the deployment scenario.
- The API server (Aapije) and Program Worker (Malgomaj) supports horizontal scaling.
- You can spread domains over several DBMS:s. However, one domain can not be split over several different DBMS:s.
- We recommend pgBouncer as a database connection proxy.


# The API specification

## Everything is on GitHub

#### API specification
[https://github.com/self-host/self-host/blob/main/api/aapije/rest/openapiv3.yaml](https://github.com/self-host/self-host/blob/main/api/aapije/rest/openapiv3.yaml)

#### Swagger UI interface
[Self-host API](https://petstore.swagger.io/?url=https://raw.githubusercontent.com/self-host/self-host/main/api/aapije/rest/openapiv3.yaml)

## It is Open

If you don't feel like using our Self-host implementation, then you are free to implement the REST API specification using any language or system that you please.

*Make it yours.*

# Design

## The abridged version

The Self-host has at the time of writing seven core concepts;

- **Users**: required for access to the API.
- **Groups**: a way to easily manage user access.
- **Policies**: access controls for groups.

---

- **Timeseries**: a series of numerical data points in time.
- **Things**: a way to group time series.
- **Datasets**: files that can store complex data types such as configuration.
- **Programs**: Self-host managed small pieces of code to perform simple tasks.

---

### Database

We designed the Self-host to use PostgreSQL v12+ because;

- We have good experience with it.
- It is free and open-source software.
- It is well maintained.

---

### Go(lang)

We decided to write the Self-host in Go because;

- The source code compiles to a binary.
- Cross-compilation (for different platforms) is a breeze.
- Static typing helps to prevent stupid mistakes.

---

### Tengo

The script language we choose for the program execution environment is called [Tengo](https://github.com/d5/tengo).

It allows one to quickly deploy small pieces of code to perform simple tasks without creating a new software development project.

It is our aim to in the future provide a library index of common programs, making it even easier to set up new programs quickly.

# Deployment

## Prebuilt Containers

To simplify deployment, we provide prebuilt docker containers of the three main components. API server (Aapije), Program Manager (Juvuln) and Program Worker (Malgomaj).

[https://hub.docker.com/u/selfhoster](https://hub.docker.com/u/selfhoster)

## Docker

The containers we provide are built for Docker.

You can use either Docker directly via the CLI or [Docker Desktop](https://www.docker.com/products/docker-desktop) for local development.

For some tips and guidelines, check out our [documentation](https://github.com/self-host/self-host/blob/main/docs/docker_deployment.md).

## Kubernetes

For reliable production deployment, Kubernetes has become the de-facto standard.

Our prebuilt Docker images workes just as well in a Kubernetes environment as in a Docker environment.

For some tips and guidelines, check out our [documentation](https://github.com/self-host/self-host/blob/main/docs/k8s_deployment.md).


# Free/Libre software and Open Source

## What is what?

The terms "free software" and "open source" stand for almost the same range of programs. However, they say profoundly different things about those programs and are based on different values. The open-source idea values mainly practical advantage and does not campaign for principles. While free or libre software aims to provide freedom for the users. The freedom to modify, fix and extend as they see fit.

## So it is essentially the same thing?

It depends on how you look at it. But, yes.


# Questions?

...



[fig1]: https://raw.githubusercontent.com/self-host/self-host/main/docs/assets/overview.svg "Overview"
