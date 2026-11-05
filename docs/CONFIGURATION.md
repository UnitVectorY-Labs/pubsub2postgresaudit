---
layout: default
title: Configuration
nav_order: 3
permalink: /configuration
---

# Configuration

## Configuration precedence

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

### Message Compression

The consumer supports compressed message payloads. Set the `compression` attribute on a Pub/Sub message to indicate the compression algorithm used for the message body:

| Attribute Value | Algorithm |
|-----------------|-----------|
| `gzip` | gzip compression |
| `zstd` | Zstandard compression |

If the `compression` attribute is absent, the message body is treated as uncompressed JSON.

The message data must contain the raw compressed bytes, not a base64 string. Unknown compression values (including an empty value) and corrupt compressed data are acknowledged and logged with outcome `decompression_error`; no row is stored.

### Additional Options

| Flag | Environment Variable | Default | Description |
|------|----------------------|---------|-------------|
| `--create-table` | `CREATE_TABLE` | `false` | Create table if it does not exist |
| `--health-port` | `HEALTH_PORT` | *(empty)* | Port for health check endpoints (off by default) |

### Google Authentication

The application uses Google Application Default Credentials (ADC):

- Set `GOOGLE_APPLICATION_CREDENTIALS` to point to a service account key file
- Or rely on Workload Identity Federation, Compute Engine metadata, or other ADC-compatible methods


`CREATE_TABLE` enables automatic creation only for the exact environment values `true` or `1`. Use `--create-table=false` to override it. Schema and table identifiers may contain only ASCII letters, digits, and underscores.

The `migrate` command also requires `PUBSUB_SUBSCRIPTION` (or its flag), even though it does not connect to Pub/Sub.

See [Usage](USAGE.md) for startup and health checks.
