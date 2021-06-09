# Deploying a test environment on a local Docker environment


# 1) Deploy the DBMS

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


# 2) Deploy the DB schema (Manually without selfctl)

*Estimated time required: 1-3 minutes*

Download the source code from; https://github.com/self-host/self-host

The folder we are interested in is "postgres/migrations".

> docker run -v {{ migration_dir }}:/migrations --network selfhost migrate/migrate -path=/migrations/ -database postgresql://postgres:mysecretpassword@pg13.selfhost:5432/selfhost-test?sslmode=disable up

The database is now ready.


# 3) Deploy the Self-host API (selfserv)

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


# 4) Deploy the Self-host Program Manager (selfpmgr)

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


# 5) Deploy the Self-host Program Worker (selfpwrk)

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


# 6) Done

The system is now deployed in a test environment.

The default secret key is "root" and belongs to the "root" user.

The API server listens for connections on port 8080 on localhost.

Using a browser, visit; `http://127.0.0.1:8080/static/swagger-ui/`, and the API documentation page should greet you.

When authenticating, use "test" as the username and "root" as the password.

