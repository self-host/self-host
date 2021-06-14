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