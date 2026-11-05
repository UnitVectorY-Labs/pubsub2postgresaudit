---
layout: default
title: Usage
nav_order: 3
permalink: /usage
---

# Usage

## Configure connections

The examples below use the binary built in [Installation](INSTALL.md). Replace the project, subscription, and database values for your environment. Supply `DB_PASSWORD` through your environment or secret manager.

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_NAME=mydb
export DB_SCHEMA=public
export DB_TABLE=audit_log
export PUBSUB_SUBSCRIPTION=projects/my-project/subscriptions/audit-sub
```

See [Configuration](CONFIGURATION.md) for every setting. For remote databases, configure `DB_SSLMODE` to meet your database's TLS requirements; its default is `disable`.

## Create the table

```bash
./pubsub2postgresaudit migrate
```

This creates the table if missing and exits. It does not create the database or schema, modify an existing table, or require a Pub/Sub connection. It still requires the subscription setting for configuration validation.

For automatic creation at consumer startup, use `--create-table` instead. Prefer a separate migration role when the runtime identity should only insert rows.

## Run the consumer

```bash
./pubsub2postgresaudit --health-port=8080
```

The default command pulls messages until SIGINT or SIGTERM. Successful inserts and duplicate IDs are acknowledged. Database errors negatively acknowledge the message so Pub/Sub can redeliver it. Invalid JSON or invalid compression is acknowledged and discarded; monitor the [logs](LOGGING.md).

Each process maps one subscription to one table. Multiple replicas can share that subscription and table; the primary key prevents duplicate rows for a redelivered Pub/Sub message ID. Republishing an event produces a new message ID and may store another row.

## Check health and delivery

```bash
curl --fail http://localhost:8080/healthz
curl --fail http://localhost:8080/readyz
```

Health endpoints are disabled unless `HEALTH_PORT` or `--health-port` is set. `/healthz` reports process liveness. `/readyz` checks the startup readiness flag and PostgreSQL connectivity; it does not verify Pub/Sub permissions, subscription availability, or successful receipt of messages.

Use the [publish and query example](EXAMPLES.md) to verify the full path.

## Run in Docker

Provide the same settings in a private environment file, here named `audit.env`, and mount credentials if your environment does not provide ADC:

```bash
docker run --rm --name audit-consumer \
  --env-file audit.env \
  -e GOOGLE_APPLICATION_CREDENTIALS=/creds/sa.json \
  -e HEALTH_PORT=8080 \
  -p 127.0.0.1:8080:8080 \
  -v /absolute/path/to/sa.json:/creds/sa.json:ro \
  pubsub2postgresaudit:local
```

Set `DB_HOST` to a hostname reachable from the container; `localhost` refers to the container itself. Run migrations first, or include `CREATE_TABLE=true` in the environment file. To migrate with the image, use the same environment and credential mounts with `--entrypoint /server` before the image name and `migrate` after it.
