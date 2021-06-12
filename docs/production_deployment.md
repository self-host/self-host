# Deployment

To deploy the system, you are going to need three things;

- A PostgreSQL DBMS server with at least one Self-host database (Domain).
- One or several instances of the API server (selfserv).
- An instance of the Program Manager (selfpmgr).
- One or several instances of the Program Worker (selfpwrk).

If you want the quick and dirty way to deploy a test environment, then take a look at the [five to fifteen-minute deployment](https://github.com/self-host/self-host/blob/main/docs/test_deployment.md).


# Configuration files (applies to selfserv, selfpmgr, selfpwrk)

The name of the configuration file is declared with the environment variable `CONFIG_FILENAME`. The file is then looked for in;

- /etc/selfhost
- $HOME/.config/selfhost"
- . (current directory)

Make sure to place the file in any of these three locations.


# Docker (using Portainer)

For details on how to deploy using Docker, see the following [guide](https://github.com/self-host/self-host/blob/main/docs/portainer_deployment.md).


# Kubernetes

For details on how to deploy using Kubernetes, see the following [guide](https://github.com/self-host/self-host/blob/main/docs/k8s_deployment.md).


## API server

The API server (seflserv) needs network access to all DBs for which it should host an API interface. It also needs network access over HTTP to the Program Manager to forward webhook calls.

A typical configuration file for an API server looks like this;

```yaml
-- selfserv.conf.yaml
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

Once running, the server software listens for changes to the file declared by the `domainfile` keyword. Thus, allowing for changes to be made to the list of domains without restarting the software.


## Program Manager

The Program Manager needs network access to all DBs for which it should manage programs. It also needs network access over HTTP to the Program Workers.

A typical configuration file for a Program Manager looks like this;

```yaml
-- selfpmgr.conf.yaml
domainfile: domains.yaml

listen:
  host: "172.16.0.2"
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

Once running, the server software listens for changes to the file declared by the `domainfile` keyword. Thus, allowing for changes to be made to the list of domains without restarting the software.


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

The Program Worker does not access the database in any direct way.


# Multiple instances

## API server

If you have a load demand that requires more than one API server, you can easily spin up several instances and then place an HTTP proxy in front. How to do this depends heavily on your specific situation and the environment you are using.

For Kubernetes, the "HTTP proxy" is called an [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/).

For Docker, there are as many different solutions as there are Ingress Controllers with Kubernetes. One that we've had great success with is [nginx-proxy](https://github.com/nginx-proxy/nginx-proxy).


## Program Worker

Depending on how many Self-host programs you have running, how often they run and for how long they run, you might want to have more than one Program Worker.

Deploying more Program Workers is as simple as starting more instances of the `selfpwrk` server, having the configuration pointing to the same Program Manager.

The Program Manager will take care of the rest and distribute the load across all workers.


## DBMS and Domains (databases)

The Self-host design allows you to deploy as many DBMSs as you need with as many databases as can fit in each. Then, if the need arises, you to take a "heavy" domain and move it to a dedicated host to free up resources.

We recommend that you place a connection proxy in front of the DBMSs as a way to avoid running out of connections. As a solution, we highly recommend [pgbouncer](https://www.pgbouncer.org/) as we have had great success using it in production for several years.
