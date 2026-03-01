---
layout: default
title: Database
nav_order: 3
permalink: /database
---

# Database Design
{: .no_toc }

## Table of Contents
{: .no_toc .text-delta }

- TOC
{:toc}

---

## Table Schema

The destination table stores one row per Pub/Sub message. The minimum required schema is:

| Column | Type | Constraint | Description |
|--------|------|-----------|-------------|
| `message_id` | `TEXT` | `PRIMARY KEY` | Pub/Sub message ID (used for deduplication) |
| `publish_time` | `TIMESTAMPTZ` | `NOT NULL` | Pub/Sub message publish timestamp |
| `attributes` | `JSONB` | `NOT NULL` | JSON object of all Pub/Sub message attributes |
| `data` | `JSONB` | `NOT NULL` | Parsed JSON object from the message data payload |

### SQL

```sql
CREATE TABLE IF NOT EXISTS "public"."audit_log" (
    message_id    TEXT        PRIMARY KEY,
    publish_time  TIMESTAMPTZ NOT NULL,
    attributes    JSONB       NOT NULL,
    data          JSONB       NOT NULL
);
```

## Field Mapping

| Table Column | Pub/Sub Source |
|-------------|---------------|
| `message_id` | `message.MessageId` |
| `publish_time` | `message.PublishTime` |
| `attributes` | `message.Attributes` (serialized as JSON object) |
| `data` | `message.Data` (must be valid JSON) |

## Deduplication

- `message_id` is the primary key and sole deduplication mechanism
- Inserts use `ON CONFLICT (message_id) DO NOTHING`
- Duplicate messages are acknowledged without error
- This makes the service safe to run as multiple replicas

## Table Creation

Tables can be created in two ways:

1. **`migrate` subcommand**: Run `pubsub2postgresaudit migrate` to create the table. This is idempotent and safe to run repeatedly.
2. **`--create-table` flag**: When enabled, the consumer service creates the table at startup if it does not exist.

Both methods use `CREATE TABLE IF NOT EXISTS` with the schema shown above.
