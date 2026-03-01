---
layout: default
title: pubsub2postgresaudit
nav_order: 1
permalink: /
---

# pubsub2postgresaudit

Pulls GCP Pub/Sub messages and persists payload + attributes to a PostgreSQL audit table.

## Overview

pubsub2postgresaudit is a stateless Go service that:

- Consumes messages from a single GCP Pub/Sub subscription via pull
- Writes each message as an audit record into a configured PostgreSQL table
- Uses `message_id` as the primary key for deduplication
- Supports safe multi-replica deployment with `ON CONFLICT DO NOTHING`
- Provides JSON structured logging
- Offers optional HTTP liveness and readiness health check endpoints

The service is designed to run anywhere, commonly containerized and deployed to Kubernetes.
