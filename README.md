# pubsub2postgresaudit

Pulls GCP Pub/Sub messages and persists payload + attributes to a PostgreSQL audit table.

## Overview

pubsub2postgresaudit is a stateless Go service that consumes messages from a single GCP Pub/Sub subscription and writes each message as an audit record into a configured PostgreSQL table. It is designed to run as a container in Kubernetes or any Docker-compatible environment.

Key features:
- Single subscription to single table mapping
- JSON structured logging with `log/slog`
- Deduplication via `message_id` primary key
- `ON CONFLICT DO NOTHING` for safe multi-replica deployment
- Optional HTTP health/readiness endpoints
- `migrate` subcommand for table creation
- Configuration via CLI flags or environment variables

## Documentation

See the [documentation site](https://unitvectory-labs.github.io/pubsub2postgresaudit/) for detailed usage, configuration, and database schema information.
