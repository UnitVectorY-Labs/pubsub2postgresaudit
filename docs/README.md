---
layout: default
title: pubsub2postgresaudit
nav_order: 1
permalink: /
---

# pubsub2postgresaudit

Persist messages from one Google Cloud Pub/Sub pull subscription into one PostgreSQL audit table. Each row retains the message ID, publish time, attributes, and JSON payload.

## Start here

1. [Installation](INSTALL.md): prepare PostgreSQL, Pub/Sub, credentials, and the application.
2. [Usage](USAGE.md): create the table, run the consumer, and check delivery.
3. [Configuration](CONFIGURATION.md): flags, environment variables, and compression.
4. [Examples](EXAMPLES.md): publish a sample event and query stored records.

See [Database](DATABASE.md) for the storage contract and [Logging](LOGGING.md) for processing outcomes and troubleshooting.

## Related applications

- [ghook2pubsub](https://github.com/UnitVectorY-Labs/ghook2pubsub) receives GitHub webhooks and publishes them to Pub/Sub; this consumer can archive those messages.
- [ghook4action](https://github.com/UnitVectorY-Labs/ghook4action) consumes GitHub events from Pub/Sub and adds matching opened issues and pull requests to GitHub organization Projects. Give it a separate subscription on the same topic to process events independently of auditing.

The consumer also accepts JSON messages from other publishers. Invalid JSON and decompression failures are acknowledged without storage; database failures are retried. Plan monitoring around these [outcomes](LOGGING.md).
