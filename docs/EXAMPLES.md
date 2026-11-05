---
layout: default
title: Examples
nav_order: 5
permalink: /examples
---

# Examples

## Publish and verify a JSON event

After [starting the consumer](USAGE.md), publish to the topic attached to its subscription. The Google identity used by this command needs permission to publish; this is separate from the consumer's subscriber identity.

```bash
gcloud pubsub topics publish my-topic \
  --project=my-project \
  --message='{"event":"login","actor":"example-user"}' \
  --attribute=source=manual-test
```

Expect an `inserted` log outcome. Connect to your PostgreSQL database with a role allowed to read the table and run:

```sql
SELECT message_id, publish_time, attributes, data
FROM public.audit_log
ORDER BY publish_time DESC
LIMIT 10;

SELECT message_id, data->>'actor' AS actor
FROM public.audit_log
WHERE data->>'event' = 'login';
```

Substitute your configured schema and table. JSONB retains JSON values, not the original byte formatting.

## Archive GitHub events alongside workflow automation

Point [ghook2pubsub](https://github.com/UnitVectorY-Labs/ghook2pubsub) at a Pub/Sub topic, create a dedicated audit subscription on that topic, and configure this application with its full subscription name. If [ghook4action](https://github.com/UnitVectorY-Labs/ghook4action) also consumes that topic, give it its own subscription.

The audit consumer stores all message attributes and the decoded JSON payload without GitHub-specific configuration. It accepts the `gzip` and `zstd` compression attributes; see [Configuration](CONFIGURATION.md#message-compression).

## Diagnose missing rows

- No processing logs: check the full subscription name, subscriber permissions, credentials, and whether another process shares the subscription.
- `invalid_json` or `decompression_error`: correct the publisher. These messages are acknowledged and will not be retried by this consumer.
- `db_error`: check database connectivity, table existence, schema/table names, and insert permissions. Messages are negatively acknowledged for redelivery.
- `duplicate`: the message ID is already stored; query the existing row rather than expecting another insert.

See [Logging](LOGGING.md) for the full outcome reference.
