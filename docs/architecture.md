# System Architecture

## Contexts

- **Message Intake Context** (`ingestor`)
- **Message Processing Context** (`processor`)
- **Message Delivery Context** (`sender`, publishes `messages.sent`)
- **Notification / fan-out context** (`notification-service`, consumes `messages.sent` and calls providers in parallel)
- **Supporting Contexts**: `simulator`, `admin-api`, `admin-ui`
- **Template Rendering Context**: `template-service` (PHP HTTP provider for delivery templates)

## Layers (core services)

`ingestor`, `processor`, and `sender` follow `domain`, `application`, `infrastructure`, and `interfaces`.

## Integration

- Backbone: NATS JetStream
- Active subjects:
  - `messages.incoming`
  - `messages.processed`
  - `messages.sent`

## Persistence

`processor` uses PostgreSQL as the system of record for processed messages, with a unique business key on `external_id`.

## Current state vs planned

Implemented:

- Full asynchronous pipeline up to `messages.sent`
- Runtime HTTP provider integration from `notification-service` to PHP `template-service` (`/send`)
- Baseline observability via Prometheus metrics and logs

Planned:

- Transactional outbox
- Explicit terminal failure flow
- End-to-end distributed tracing
