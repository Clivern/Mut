# Mut Helm chart

Installs the Mut server, the worker, a migration job, and NATS into a cluster. PostgreSQL is external.

## Install

```sh
cp examples/my-values.yaml my-values.yaml
cp examples/secrets.example.yaml secrets.yaml   # fill in secrets.yaml
helm install mut . -f my-values.yaml -f secrets.yaml
```

Edit your copies, not `values.yaml`. Upgrade with the same files:

```sh
helm upgrade mut . -f my-values.yaml -f secrets.yaml
```

`secrets.yaml` is git-ignored. Only commit `examples/secrets.example.yaml`.

## What it runs

- A migration job runs `mut migrate up` before each install and upgrade.
- The server is a ClusterIP Service on port 80, with an Ingress when `ingressHost` is set.
- The worker has no Service. Scale it with `replicas.worker`.
- NATS runs in the release unless `nats.url` is set.

## Status

Not yet tested on a cluster. `helm lint` and `helm template` have not run.
