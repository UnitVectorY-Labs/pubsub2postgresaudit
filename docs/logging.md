---
layout: default
title: Logging
nav_order: 4
permalink: /logging
---

# Logging
{: .no_toc }

## Table of Contents
{: .no_toc .text-delta }

- TOC
{:toc}

---

## Overview

pubsub2postgresaudit uses Go's `log/slog` with a JSON handler for all logging output. Logs are written to stdout.

## Message Processing Events

Each message processing attempt produces a structured log event with the following fields:

| Field | Description |
|-------|-------------|
| `subscription` | Full Pub/Sub subscription name |
| `message_id` | Pub/Sub message ID |
| `publish_time` | Message publish timestamp (RFC 3339) |
| `db_schema` | Target database schema |
| `db_table` | Target database table |
| `outcome` | Processing result (see below) |
| `error` | Error details (for error outcomes only) |

## Outcome Values

| Outcome | Description | Message Ack? |
|---------|-------------|-------------|
| `inserted` | Row was successfully inserted | Yes |
| `duplicate` | Message ID already exists (conflict on primary key) | Yes |
| `invalid_json` | Message data was not valid JSON | Yes |
| `db_error` | Database operation failed | No (nack) |
| `pubsub_error` | Pub/Sub receive error | N/A |

## Privacy

- Logs **never** include the full message payload (`data` field)
- For invalid JSON messages, the log includes the parse error and byte size of the payload, but not the payload content
