# Kubernetes deployment

When deploying to Kubernetes, there are a few things we recommend you to do;

- Use [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) when deploying selfserv, selfpmgr and selfpwrk.
- Store all configuration in [ConfigMaps](https://kubernetes.io/docs/concepts/configuration/configmap/)
- Use an [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) for HTTP traffic from the internet.
- Use [PgBouncer](https://www.pgbouncer.org/) as a Connection Proxy to the DBMS(s). A pre-built container image can be found at [ganehag/pgbouncer:latest](https://hub.docker.com/r/ganehag/pgbouncer
- Only scale up `selfserv` and `selfpwrk` to more replicas if the load requires it.


## Examples

### Selfserv

Below we give an example of the YAML files required to deploy the `selfserv` program. We leave it as an exercise to the reader to add/modify/extend these as suited for their environment, along with the files required to deploy `selfpmgr` and `selfpwrk`.

#### ConfigMaps

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: selfhost-domains-conf
data:
  domains.yaml: |-
    domains:
      test: postgresql://postgres:secret@pg13.selfserv:5432/selfhost-test
```

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: selfserv-conf
data:
  selfserv.conf.yaml: |-
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
  name: selfserv-deployment
  labels:
    app: selfserv
spec:
  replicas: 1
  selector:
    matchLabels:
      app: selfserv
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  template:
    metadata:
      labels:
        app: selfserv
    spec:
      containers:
        - name: tengil
          image: selfhoster/selfserv:main
          env:
          - name: CONFIG_FILENAME
            value: selfserv.conf.yaml
          ports:
          - containerPort: 80
          volumeMounts:
            - name: selfserv-conf
              mountPath: /etc/selfhost
              readOnly: true
            - name: domains
              mountPath: /etc/selfhost/domains
              readOnly: true
      volumes:
        - name: selfserv-conf
          configMap:
            name: selfserv-conf
        - name: domains
          configMap:
            name: selfhost-domains-conf
```
