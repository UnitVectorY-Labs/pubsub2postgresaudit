[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT) [![Work In Progress](https://img.shields.io/badge/Status-Work%20In%20Progress-yellow)](https://guide.unitvectorylabs.com/bestpractices/status/#work-in-progress) 

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

## Related projects

- [ghook2pubsub](https://github.com/UnitVectorY-Labs/ghook2pubsub) receives GitHub webhooks and publishes them to Pub/Sub; pubsub2postgresaudit can archive those events in PostgreSQL.
- [ghook4action](https://github.com/UnitVectorY-Labs/ghook4action) consumes GitHub events from Pub/Sub and adds matching opened issues and pull requests to GitHub organization Projects. Use separate subscriptions on the same topic to audit events and update Projects independently.

Start with the [documentation index](docs/README.md), [installation](docs/INSTALL.md), and [usage guide](docs/USAGE.md).
