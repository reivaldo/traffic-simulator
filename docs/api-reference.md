# API Reference

This reference separates external application contracts (HTTP APIs) from internal integration contracts (domain/application events).

## Base URLs

- Admin API: `http://localhost:8086`
- Simulator API: `http://localhost:8081`
- Ingestor API: `http://localhost:8082`

## External Application Contracts

### Ingestor

#### `POST /messages`

Accepts a `MessageIntent` and emits `MessageAccepted` internally.

Request:

```json
{
  "external_id": "ext-123",
  "channel": "email",
  "recipient": "user@example.com",
  "payload": {
    "subject": "hello",
    "body": "world"
  }
}
```

Response:

- `202 Accepted` when accepted for asynchronous processing
- `400 Bad Request` for invalid payload
- `500 Internal Server Error` for integration failures

### Simulator

#### `GET /start?duration=<seconds>`

Starts a simulation run.

#### `GET /stop`

Stops current simulation.

#### `GET /status`

Returns runtime simulation status.

### Admin API

#### `GET /simulator/start`

Proxy endpoint that starts simulation through simulator context.

#### `GET /simulator/stop`

Proxy endpoint that stops simulation.

#### `GET /simulator/status`

Proxy endpoint for current simulation state.

#### `GET /metrics/simulator`

Returns simulator metrics payload for dashboard consumption.

## Operational Endpoints

Each core service exposes:

- `GET /healthz`
- `GET /readyz`
- `GET /metrics`

## Internal Event Contracts

### Subject: `messages.incoming`

- Producer: `Message Intake Context`
- Consumer: `Message Processing Context`
- Event type: `MessageAccepted`

### Subject: `messages.processed`

- Producer: `Message Processing Context`
- Consumer: `Message Delivery Context`
- Event type: `MessageProcessed`

### Subject: `messages.sent`

- Producer: `Message Delivery Context`
- Consumer: observability pipelines and downstream analytics
- Event type: `MessageSent`

## Contract Governance

To keep DDD consistency during refactoring:

- HTTP contracts represent application layer use cases.
- Event contracts represent inter-context integration.
- Ubiquitous language terms must be stable across both levels.

Any rename of event names, fields, or status semantics should be treated as an explicit architectural change and tracked in ADRs.
