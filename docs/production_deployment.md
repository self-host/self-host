# Deployment

To deploy the system, you are going to need three things;

- A PostgreSQL DBMS server with at least one Self-host database (Domain).
- One or several instances of the API server (selfserv).
- An instance of the Program Manager (selfpmgr).
- One or several instances of the Program Worker (selfpwrk).


# Docker

...


# Kubernetes

...


# Configuration files

The name of the configuration file is declared with the environment variable `CONFIG_FILENAME`. The file is then looked for in;

- /etc/selfhost
- $HOME/.config/selfhost"
- . (current directory)

Make sure to place the file in any of these three locations.


## API server

...


## Program Manager

The Program Manager needs network access to all DBs for which it should manage programs. It also needs network access over HTTP to the Program Workers.

A typical configuration file for a Program Manager looks like this;

```yaml
-- example.conf.yaml
domainfile: domains.yaml

listen:
  host: "172.16.0.1"
  port: 80
```

The `listen.host` parameter can be either IP or hostname.

The `domainfile` parameter points to a YAML file with connection information to all databases.

A typical `domains.yaml` file can look like this:

```yaml
domains:
  test0: postgresql://postgres:secret@myhost.mydomain:5432/selfhost-test0
  test1: postgresql://postgres:secret@myhost.mydomain:5432/selfhost-test1
  test2: postgresql://postgres:secret@myhost.mydomain:5432/selfhost-test2
  test3: postgresql://postgres:secret@myhost.mydomain:5432/selfhost-test3
```

# Program Worker

A Program Worker needs access to the Program Manager and, depending on the situation, especially if HTTP request to the Internet at large is required, access the Internet.

**NOTE**: For a program to interact with the Self-host API server (selfserv) using the HTTP module is a requirement.

It is also a requirement to have access to the Program Manager, as the Worker has to keep the Manager updated on its state constantly. The Program Manager is also the `Library Manager`, which supplies the Worker (on request) with the source code to `modules`.

A typical configuration file for a Program Worker looks like this;

```yaml
listen:
  host: 172.16.0.151
  port: 80

cache:
  library_timeout: 10
  program_timeout: 5

module_library:
  scheme: http
  authority: 172.16.0.1:80

program_manager:
  scheme: http
  authority: 172.16.0.1:80
```

