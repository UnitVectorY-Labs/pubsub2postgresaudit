---
layout: default
title: Usage
nav_order: 2
permalink: /usage
---

# Usage
{: .no_toc }

## Table of Contents
{: .no_toc .text-delta }

- TOC
{:toc}

---

## Running the Consumer

The default behavior (no subcommand) runs the consumer service that pulls messages from Pub/Sub and inserts them into PostgreSQL.

```bash
pubsub2postgresaudit \
  --db-host=localhost \
  --db-port=5432 \
  --db-user=postgres \
  --db-password=secret \
  --db-name=mydb \
  --db-table=audit_log \
  --pubsub-subscription=projects/my-project/subscriptions/my-sub
```

## Migrate Command

The `migrate` subcommand creates the required database table and exits. It is idempotent and safe to run repeatedly.

```bash
pubsub2postgresaudit migrate \
  --db-host=localhost \
  --db-port=5432 \
  --db-user=postgres \
  --db-password=secret \
  --db-name=mydb \
  --db-table=audit_log \
  --pubsub-subscription=projects/my-project/subscriptions/my-sub
```

## Configuration

All configuration is available via command line flags or environment variables. Command line flags take precedence over environment variables when both are provided.

### PostgreSQL Configuration

| Flag | Environment Variable | Default | Description |
|------|---------------------|---------|-------------|
| `--db-host` | `DB_HOST` | `localhost` | PostgreSQL server hostname |
| `--db-port` | `DB_PORT` | `5432` | PostgreSQL server port |
| `--db-user` | `DB_USER` | `postgres` | Database user |
| `--db-password` | `DB_PASSWORD` | *(empty)* | Database password |
| `--db-name` | `DB_NAME` | `cert_observatory` | Database name |
| `--db-sslmode` | `DB_SSLMODE` | `disable` | SSL mode (`disable`, `require`, `verify-ca`, `verify-full`) |
| `--db-schema` | `DB_SCHEMA` | `public` | Destination schema |
| `--db-table` | `DB_TABLE` | *(empty, required)* | Destination table name |

### Pub/Sub Configuration

| Flag | Environment Variable | Default | Description |
|------|----------------------|---------|-------------|
| `--pubsub-subscription` | `PUBSUB_SUBSCRIPTION` | *(empty, required)* | Full subscription name (`projects/<p>/subscriptions/<s>`) |

### Additional Options

| Flag | Environment Variable | Default | Description |
|------|----------------------|---------|-------------|
| `--create-table` | `CREATE_TABLE` | `false` | Create table if it does not exist |
| `--health-port` | `HEALTH_PORT` | *(empty)* | Port for health check endpoints (off by default) |

### Google Authentication

The application uses Google Application Default Credentials (ADC):

- Set `GOOGLE_APPLICATION_CREDENTIALS` to point to a service account key file
- Or rely on Workload Identity Federation, Compute Engine metadata, or other ADC-compatible methods

## Health Endpoints

Health endpoints are off by default. Enable them by setting `--health-port` or `HEALTH_PORT`:

```bash
pubsub2postgresaudit \
  --db-table=audit_log \
  --pubsub-subscription=projects/my-project/subscriptions/my-sub \
  --health-port=8080
```

- `GET /healthz` — Liveness probe. Returns `200 OK` when the process is running.
- `GET /readyz` — Readiness probe. Returns `200 OK` only when PostgreSQL is reachable and the Pub/Sub client has been initialized.

## Docker

```bash
docker run \
  -e DB_HOST=postgres \
  -e DB_PORT=5432 \
  -e DB_USER=postgres \
  -e DB_PASSWORD=secret \
  -e DB_NAME=mydb \
  -e DB_TABLE=audit_log \
  -e PUBSUB_SUBSCRIPTION=projects/my-project/subscriptions/my-sub \
  -e GOOGLE_APPLICATION_CREDENTIALS=/creds/sa.json \
  -v /path/to/creds:/creds:ro \
  ghcr.io/unitvectory-labs/pubsub2postgresaudit
```