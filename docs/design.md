# Design

The system is split up into several separate softwares.

- Database (PostgreSQL)
- Self-host API (selfserv)
- Program Manager (selfpmgr)
- Program Worker (selfpwrk)
- CLI tool (selfctl)

[design drawing]


# Note on Configuration

We decided on using config files instead of Environment variables as a way to "avoid" leaking data to the environment.

Environment variables are good for things like INSTANCES=10 and TIMEOUT=3600. But not so good for DBURL=postgres://postgres:secret@mydbms.com:5432/mydatabase since it exposes secret information.


# Database

The only supported database is PostgreSQL version 12 or newer.


# Primus database (Work in Progress)

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


# Authentication (Users)

To handle authentication, the API used both [RFC-7617](https://tools.ietf.org/html/rfc7617) (The 'Basic' HTTP Authentication Scheme) and [RFC-8959](https://tools.ietf.org/html/rfc8959) (The 'secret-token' URI Scheme).

Example;  
`Authorization: Basic c2VjcmV0LXRva2VuOm15ZG9tYWluLllhNGJkNHph`.

The Base64 encoded part in the example above decodes to:  
`mydomain:secret-token.Ya4bd4za`.

The access token is composed of three parts;

* domain (*mydomain*)
* secret-token-scheme (*secret-token*)
* secret (*Ya4bd4za*)

A single colon (":") separates the domain and token parts, and a single dot (".") separates the secret-token-scheme and secret parts. This format effectively makes the domain the username and the "secret-token-scheme.secret" the password from RFC-7617.

The domain part is used to check the secret against the correct domain as each access token is unique to a specific domain. We chose this structure as it is a straightforward extension of a well-established authentication process.

**NOTE**: The value of secret-token-scheme must always be *secret-token*. This value aims to ease automatic identification and prevention of keys in source code.


## Questions and Answers

**Q**: Why not use JWT (JSON Web Token)? Is it an IETF standard? Also, it is an excellent fit for REST API:s.  
**A**: While it is true that JWT is a proposed standard by the IETF, and it is also true that it removes the requirement to authenticate each request. JWT carries a significant transport overhead as each request holds a simple token and an entire encrypted blob of meta-data. We think that this overhead outweighs the benefits of having to authenticate each request. Even more so, since the only way to disown a client is to use a global blacklist. As each request has to be validated against this list, we re-introduce the same overhead we have when authenticating each request. This overhead renders JWT pointless from a performance perspective. 


# Access Control (Policies)

A typical REST API would use RBAC (Role-Based Access Control). Implemented in such a way that, for example, if we have an endpoint 'orders', we may have three different groups;

- order_manager: can create, read, update and delete orders.
- order_editor: can create, read and update orders, but not delete.
- order_inspector: can only read orders.

RBAC works well for most cases and has been around since 1992. But for the sake of not choosing the first thing you find when you search the web, let's take a look at what alternatives exist.

ACLs or Access Control Lists, used in traditional discretionary access-control systems, assigns permission to low-level objects. In contrast, RBAC systems give permissions to specific operations with meaning (business logic) within the organization.

ABAC or Attribute-based Access Control is a model that evolves from RBAC to consider additional attributes to roles and groups. Characteristics such as context, e.g. time, location, IP or actions such as create and delete. ABAC is policy-based in the sense that it uses policies rather than static permissions to define what is allowed or what is not allowed.

So what is best?

To know that we need to take a step back for a moment and ask ourselves what we need;

- Endpoint level permissions. To limit who can create and list resources. For example, to create a new time series.
- Resource level permissions. To limit who can read, update and delete resources. For example, to delete a specific time series.

Just to clarify; An endpoint (in this context) is for example `/v2/timeseries` while a resource is for example `/v2/timeseries/588793c6-4377-4ba1-ae35-d762495d0ad5`.

New time series are created using a POST request to /v1/timeseries in this example. This call will return a new object with a generated UUID. We can then use this new UUID to perform updates (PUT), read (GET) and delete (DELETE) operations on the time series.

Performing a read request (GET) on /v1/timeseries should present the user with a list of all timeseries.

Can we solve this using RBAC? Yes, we can. By declaring a set of roles for all the different operations the system will need to support. However, this can quickly become quite complicated as the number of interface surfaces expands. Even more so, if the demand arises, that a particular user needs read access to a specific resource. Yet, at the same time should not have access to anything else. Depending on how dynamic the RBAC system is in its implementation, this can be as simple as adding a row to a table, or it can result in a significant rewrite of the entire software.

What about ACLs? Sure, one can use ACLs. With one drawback. Not everything is a resource, and each user has to be assigned the correct resources.

So, ABAC then? Maybe. By breaking access control into a set of policies where we can allow or deny access to CRUD actions based on resource location. Thus we create a versatile access control system. To make it easier to manage, we can introduce groups as the policyholder instead of users. Making the process of adding a user to a set of policies "super easy, barely an inconvenience".

Compared to ACL or RBAC, ABAC does require additional computation to resolve the permission from the policy rules. There are ways to optimize this, but as long as the set computed set of policies is kept as few as possible, there should be a limited impact on performance.

Take, for example, a situation where a single user needs access to all but one resource (present and future). In this case, it makes sense to assign a policy rule of "allow all", then a policy rule to "deny" access to the specific resource. As this only produces two policy rows. Whereas the alternative would be to assign one "allow rule" to the user for each accessible resource.

**Q**: Why role your own permission system? PostgreSQL already has an excellent RBAC permission system with ROW level security.  
**A**: This is a good point, and we considered the existing POLICY system in PostgreSQL. However, some cases where a query will return an empty set instead of a "permission denied" exception caused a little too much headache. We may very well revisit this in the future.  


# Data partitioning (Time Series)

In PostgreSQL, there are two (primary) ways to handle data partitioning. Using INHERITANCE or using PARTITIONS. The former is the older solution (from around version 7), and the latter was introduced in PostgreSQL 10.

Choosing one or the other is a choice between different drawbacks. Neither solutions are a complete automated solution as they both lack support for automatic creation of partition tables. The partition table must exist before we can insert any data into it. 

But, there are ways to achieve an automated solution anyway. Although, with some limitations.

When using INHERITANCE, it is possible to use a before trigger on the parent table. This trigger then takes care of creating the table if it is missing and inserting data to it. The problem with this is that the trigger has to return NULL; otherwise, the new row will also be inserted into the parent table, resulting in duplicate rows. When the trigger returns NULL, it will appear as if the INSERT statement returned no new rows.

When using PARTITIONS, it is impossible to assign a trigger to the parent table; we can only assign triggers to the partitions, which will not help us. Thus, the only logical solution is to use a separate function to insert the data to the correct table. This function is very similar to the before trigger one would use when using INHERITANCE.

Below is an example of such a function;

```sql
CREATE OR REPLACE FUNCTION public.measurement_insert(
    city_id integer,
    logdate timestamp with time zone,
    peaktemp integer,
    unitsales integer)
    RETURNS SETOF measurement 
    LANGUAGE 'plpgsql'
    RETURNS NULL ON NULL INPUT
AS $BODY$
BEGIN
    RETURN QUERY
    INSERT INTO measurement(
        city_id, logdate, peaktemp, unitsales)
    VALUES(city_id, logdate, peaktemp, unitsales)
    RETURNING *;

EXCEPTION
    WHEN check_violation THEN
    EXECUTE
    'CREATE TABLE IF NOT EXISTS' || 
    ' measurement_y' || to_char(logdate, 'YYYYmMM') ||
    ' PARTITION OF measurement FOR VALUES FROM (''' ||
    date_trunc('month', logdate) ||
    ''') TO (''' || 
    (date_trunc('month', logdate) + '1month'::interval) || 
    ''')';

    RETURN QUERY
    INSERT INTO measurement(
        city_id, logdate, peaktemp, unitsales)
    VALUES(city_id, logdate, peaktemp, unitsales)
    RETURNING *;
END;
$BODY$;
```

What this function does is that it tries to inserts the data. If it fails because the table does not exist, it creates the table and tries to insert the data once again. This way, we should avoid any overhead by establishing that the table exists before inserting data.


# Units (Time Series)

While it makes sense from an academic viewpoint to only use the seven SI base units (second, meter, kilogram, ampere, kelvin, mole and candela) for storing time series data. It does not make as much sense from a performance perspective.

One needs to take the most common use cases into account to avoid introducing rules that will always result in performance penalties.

While it makes sense to store temperature in kelvin as a way to always have a common unit for temperature, it makes less sense in a real world scenario as the kelvin scale is hardly ever used outside of the laboratory.

Having an indoor temperature value on the kelvin scale will always result in convertion to Celsius for all countries except the United States, Belize, Palau, the Bahamas and the Cayman Islands, where Fahrenheit is used.

Most people want to work with value ranges they are confortable with, an indoor temperature of `294.15 K` makes little sense to most people while `21 C` or `80 F` is much clearer.


