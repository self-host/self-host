# README

## Requirements

```
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.14.1
```

## Local PostgreSQL

```
docker run \
  -d \
  -e POSTGRES_HOST_AUTH_METHOD=trust \
  -e POSTGRES_USER=user \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=dbname \
  -p 5432:5432 \
  postgres:12.5-alpine
```

## Migrations

Create a new up/down set:

NOTE: Transactions must be wrapped in BEGIN COMMIT.

```sql
BEGIN;

CREATE TYPE enum_mood AS ENUM (
	'happy',
	'sad',
	'neutral'
);
ALTER TABLE users ADD COLUMN mood enum_mood;

COMMIT;
```

```
migrate create -ext sql -dir pkg/db/migrations/ <migration name>
```

Run:

```
migrate -path pkg/db/migrations/ -database postgres://user:password@localhost:5432/dbname?sslmode=disable up
```

