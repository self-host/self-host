# Alerts

The alert system is a trivial solution for external clients to store and query messages sent by themselves or other external clients.

It is not a rule engine to build and execute alarm rules. But the location where those alerts end up. This solution does not aim to replace solutions such as;

- Alerta
- Logstash
- Grafana Loki
- Logwatch
- Syslog-ng
- etc.

It is merely a way to achieve the bare minimum without external dependencies.


## Design

Taking inspiration from Alerta, we have decided on the following schema;

### Alerts

| Attribute           | Format      | Description                                                            |
|---------------------|-------------|------------------------------------------------------------------------|
| id                  | UUID        | unique id for this alert.                                              |
| resource            | text        | resource under alarm, deliberately not host-centric.                   |
| event               | text        | event name, e.g. Crashed, DEADLINE: EXCEEDED.                          |
| environment         | text        | effected environment used to namespace the resource.                   |
| severity            | int         | the severity of an alert (default normal). See Severities table.       |
| status              | int         | status of alert (default open). See Status table.                      |
| service             | []text      | list of affected services.                                             |
| value               | text        | event value eg. 45%, Down, PingFail, >500ms.                           |
| description         | text        | freeform text description.                                             |
| tags                | []text      | set of tags in any format eg. myTag, valDouble:Tag, a:Map=Tag.         |
| origin              | text        | name of monitoring component that generated the alert.                 |
| created             | timestamptz | date-time the alert was generated in RFC-3339 format with a time zone. |
| timeout             | int         | number of seconds before an alert is considered stale.                 |
| rawdata             | bytea       | unprocessed data, e.g. full log message or Exception trace.            |
| duplicate           | int         | the number of times this event has been received for a resource.       |
| previous\_severity  | int         | the previous severity of the same event for this resource.             |
| last\_receive\_time | timestamptz | the last time we received this alert.                                  |


### Severities

| Severity      | Description                                                       |
|---------------|-------------------------------------------------------------------|
| security      | Security related alert.                                           |
| critical      | Things that need fixing ASAP.                                     |
| major         | A huge problem, but it can wait until after lunch.                |
| minor         | Not good doesn't need fixing right now. It can wait until tomorrow. |
| warning       | This issue has to be handled... someday.                            |
| informational | For your information.                                             |
| debug         | Testing, testing. Can you hear me?                                |
| trace         | Tracking an entity/event throughout several layers of the stack.  |
| indeterminate | Sorry Dave, I don't know that one.                                |


### Status

| Status      | Description                                                                       |
|-------------|-----------------------------------------------------------------------------------|
| open        | The alert is active.                                                              |
| close       | The alert has been closed by external action.                                     |
| expire      | The alert expired due to `timeout`. The next occurrence will result in a new alert.   |
| shelve      | The alert was put on hold by external action. A shelved alert will never expire.  |
| acknowledge | We acknowledged the alert.                                                       |
| unknown     | The alert is in an unknown state.                                                 |

## Adding new alerts

Whenever we add a new alert to the system, the system first determines if a matching alert already exists. An alert matches if;

- It has the same resource and,
- It has the same environment and,
- It has the same event and,
- It has the same origin and,
- It is open or acknowledged

If a match exists, the `duplicate` value will increase, along with the update of `previous_severity` and `last_receive_time`.

If no match can be found, then a new alert post is added to the system.

## Searching for alerts

There are several items to filter on when searching for alerts.

- resource: string
- environment: string
- event: string
- origin: string
- status: string
- severity (Less or equal to, Greater or equal to, Equal to): string
- tags: []string
- service []string

For example; if one would like to find all the `open` alerts in the `Production` environment, then use;

`status=open` and `environment=Production`.

Or if one wants to find all the `critical` alerts affecting the `web` service one can use;

`severity=critical` and `service=[web]`.
