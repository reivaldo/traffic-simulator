# ADR-005: PostgreSQL Schema Design

**Status:** Accepted  
**Date:** April 29, 2026

## Context

`processor` must persist messages with idempotency by `external_id` and record processing progress before and after publishing the `messages.processed` event.

## Decision

The `messages` table uses:

- `id` UUID with `gen_random_uuid()`
- `external_id` unique and required
- `channel`, `recipient`, `payload`
- `status` (`processed_pending_publish` -> `processed`)
- `created_at` and `updated_at`

Schema is applied in `db/migrations/001_create_messages.sql`.

## Consequences

- Reduces duplicate side effects caused by redelivery.
- Improves state traceability during processing.
- It does not fully solve DB/event atomicity yet (outbox remains in the backlog).
