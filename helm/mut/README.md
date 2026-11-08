# Mut Helm chart

Installs the Mut server, the worker, a migration job, and a NATS server (optional) into a cluster.
PostgreSQL is external: point the chart at a database you already run.

## Install

```sh
helm install mut ./helm/mut \
  --set server.publicUrl=https://mut.example.com \
  --set postgres.host=<db-host> \
  --set postgres.password=<db-password> \
  --set secrets.encryptionKey=<32-byte-key>
```

Keep secrets out of the command line in real installs. Create a secret and reference it:

```sh
kubectl create secret generic mut-db --from-literal=MUT_DATABASE_PASSWORD=<db-password>
helm install mut ./helm/mut --set postgres.existingSecret=mut-db ...
```

## Secrets

- `postgres.existingSecret` must contain `postgres.existingSecretKey` (default `MUT_DATABASE_PASSWORD`).
- `secrets.existingSecret` must contain `MUT_ENCRYPTION_KEY`, `MUT_SERVER_PROM_METRICS_SECRET`,
  `MUT_AI_API_KEY`, `MUT_RESEND_API_KEY`, `MUT_OAUTH_GITHUB_CLIENT_SECRET` and `MUT_OAUTH_GOOGLE_CLIENT_SECRET`.
  Use an empty value for keys you don't use. If `postgres.existingSecret` is empty, the same secret must also hold `MUT_DATABASE_PASSWORD`.

## How it runs

- The migration job runs `mut migrate up` before each install and upgrade.
- The server serves HTTP on port 80 through a ClusterIP Service, with an optional Ingress.
- The worker runs `mut worker` with no Service. Scale it with `worker.replicas`.
- NATS runs as a single core NATS pod unless `nats.enabled=false` and `nats.url` is set.

## Status

Not yet tested against a cluster. The templates parse, but `helm lint` and `helm template` have not run here.
