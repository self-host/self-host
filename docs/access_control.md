# Access Control (Policies)

## How access control works in general

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

You can create a new time series using a POST request to /v1/timeseries in this example. This call will return a new object with a generated UUID. We can then use this new UUID to perform updates (PUT), read (GET) and delete (DELETE) operations on the time series.

Performing a read request (GET) on /v1/timeseries should present the user with a list of all timeseries.

Can we solve this using RBAC? Yes, we can. By declaring a set of roles for all the different operations the system will need to support. However, this can quickly become quite complicated as the number of interface surfaces expands. Even more so, if the demand arises, that a particular user needs read access to a specific resource. Yet, at the same time should not have access to anything else. Depending on how dynamic the RBAC system is in its implementation, this can be as simple as adding a row to a table, or it can result in a significant rewrite of the entire software.

What about ACLs? Sure, one can use ACLs. With one drawback. Not everything is a resource, and each user has to be assigned the correct resources.

So, ABAC then? Maybe. By breaking access control into a set of policies where we can allow or deny access to CRUD actions based on resource location. Thus we create a versatile access control system. To make it easier to manage, we can introduce groups as the policyholder instead of users. Making the process of adding a user to a set of policies "super easy, barely an inconvenience".

Compared to ACL or RBAC, ABAC does require additional computation to resolve the permission from the policy rules. There are ways to optimize this, but as long as the set computed set of policies is kept as few as possible, there should be a limited impact on performance.

Take, for example, a situation where a single user needs access to all but one resource (present and future). In this case, assigning a policy rule of "allow all" makes sense, then a policy rule to "deny" access to the specific resource. As this only produces two policy rows. The alternative would be to assign one "allow rule" to the user for each accessible resource.


## How access control works in the Self-host

In our solution, we assign policy rules to `groups`. Each `user` can belong to one or several `groups`, granting or denying the `user` access privileges.

There are four different `actions` defined in a CRUD policy schema.

- CREATE
- READ
- UPDATE
- DELETE

There are two different effects; `allow` and `deny` and a `resource path`, which may include a wild card character `%`.

A typical policy rule can look like this;

```
Policy ID        Group ID           Priority  Effect   Action    Resource
"d8a11...97380"  "86efe...8cb5e2a"  0         "allow"  "create"  "timeseries/%"
```

This rule gives a user belonging to group "86efe...8cb5e2a" the right to perform CREATE actions on any resource matching the resource path "timeseries/%". Because the resource path contains a wild card character, the resource path matches all of the time series and sub-endpoints.

If we want to grant access to all of the time series except a specific one, we can subtract the access-right from that particular resource using a "deny" rule.

```
Policy ID       Group ID        Priority Effect  Action  Resource
"dac12...96fc0" "86efe...8c2a"  10       "deny" "create" "timeseries/924...21/%"
```

Rules are computed such that all `allow` rules are combined, then all `deny` rules are applied to retract access privileges.

For details on which resource paths are relevant, see the [openapiv3.yaml](https://github.com/self-host/self-host/blob/main/api/selfserv/rest/openapiv3.yaml) specification. Look at the `BasicAuth` declaration for each endpoint where the required access privilege is declared on the form; `action:resource_path`.


## Questions

#### Why role your own permission system? PostgreSQL already has an excellent RBAC permission system with ROW level security.

This question is a good point, and we considered the existing POLICY system in PostgreSQL. However, some cases where a query will return an empty set instead of a "permission denied" exception caused a little too much headache. We may very well revisit this in the future.
