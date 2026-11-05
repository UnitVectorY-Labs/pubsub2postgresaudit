---
layout: default
title: Installation
nav_order: 2
permalink: /install
---

# Installation

## Prerequisites

- An existing PostgreSQL database and schema, reachable from the consumer. The runtime role needs schema `USAGE` and table `INSERT`; table creation also needs schema `CREATE`.
- An existing Pub/Sub topic and pull subscription. Grant the application's Google identity Pub/Sub Subscriber access to that subscription.
- Google Application Default Credentials: use an attached service identity or workload identity in production, or `gcloud auth application-default login` for local use. A service account file can be selected with `GOOGLE_APPLICATION_CREDENTIALS`.

Use a dedicated subscription for auditing. Consumers sharing a subscription divide messages; separate subscriptions on the same topic each receive their own copy.

## Build from source

Install a Go toolchain meeting the version in [go.mod](../go.mod), then run:

```bash
git clone https://github.com/UnitVectorY-Labs/pubsub2postgresaudit.git
cd pubsub2postgresaudit
go build -o pubsub2postgresaudit .
./pubsub2postgresaudit --help
```

## Container

The release image is `ghcr.io/unitvectory-labs/pubsub2postgresaudit`. Select a published release tag for your deployment. Alternatively build locally:

```bash
docker build -t pubsub2postgresaudit:local .
```

The container runs as UID/GID 65532. Mounted credential files must be readable by that user. Its command is `/server`; to run a subcommand, explicitly select that entrypoint:

```bash
docker run --rm --entrypoint /server pubsub2postgresaudit:local --help
```

Continue with [Usage](USAGE.md) to configure connections and create the audit table.
