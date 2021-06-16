# Kubernetes deployment

When deploying to Kubernetes, there are a few things we recommend you to do;

- Use [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) when deploying aapije, juvuln and malgomaj.
- Store all configuration in [ConfigMaps](https://kubernetes.io/docs/concepts/configuration/configmap/)
- Use an [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) for HTTP traffic from the internet.
- Use [PgBouncer](https://www.pgbouncer.org/) as a Connection Proxy to the DBMS(s). A pre-built container image can be found at [ganehag/pgbouncer:latest](https://hub.docker.com/r/ganehag/pgbouncer)
- Only scale up `aapije` and `malgomaj` to more replicas if the load requires it.


## Examples

### Selfserv

Below we give an example of the YAML files required to deploy the `aapije` program. We leave it as an exercise to the reader to add/modify/extend these as suited for their environment, along with the files required to deploy `juvuln` and `malgomaj`.

#### ConfigMaps

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: selfhost-domains-conf
data:
  domains.yaml: |-
    domains:
      test: postgresql://postgres:secret@pg13.aapije:5432/selfhost-test
```

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: aapije-conf
data:
  aapije.conf.yaml: |-
    listen:
      host: 0.0.0.0
      port: 80

    domainfile: domains/domains.yaml
```

#### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aapije-deployment
  labels:
    app: aapije
spec:
  replicas: 1
  selector:
    matchLabels:
      app: aapije
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  template:
    metadata:
      labels:
        app: aapije
    spec:
      containers:
        - name: aapije
          image: selfhoster/aapije:latest
          env:
          - name: CONFIG_FILENAME
            value: aapije.conf.yaml
          ports:
          - containerPort: 80
          volumeMounts:
            - name: aapije-conf
              mountPath: /etc/selfhost
              readOnly: true
            - name: domains
              mountPath: /etc/selfhost/domains
              readOnly: true
      volumes:
        - name: aapije-conf
          configMap:
            name: aapije-conf
        - name: domains
          configMap:
            name: selfhost-domains-conf
```
