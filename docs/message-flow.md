# Message Flow

Current flow implemented in code:

1. `simulator` sends `POST /messages` to `ingestor`.
2. `ingestor` validates and publishes to `messages.incoming`.
3. `processor` consumes `messages.incoming`.
4. `processor` persists with status `processed_pending_publish`.
5. `processor` publishes to `messages.processed`.
6. `processor` marks final status `processed`.
7. `sender` consumes `messages.processed`.
8. `sender` publishes to `messages.sent`.
9. `notification-service` consumes `messages.sent` and applies fan-out to provider HTTP endpoints.
10. Provider HTTP calls are handled by PHP `template-service` (`POST /send`) in the default local setup.

## Consumption semantics

- Invalid payload: `Ack` (do not redeliver poison messages).
- Processing error: `Nak` (allow redelivery).
- Success: `Ack`.

## Current limitations

- No explicit terminal failure event on the bus.
- No outbox between persistence and event publication.
