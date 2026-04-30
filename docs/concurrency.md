# Concurrency

## Simulator

- Uses `atomic` for run state.
- Uses per-run `context.Context` (`runCtx` / `runCancel`).
- `/stop` cancels only the active run.
- Uses a **worker pool** (10 workers + queue) for parallel POST requests to `ingestor`; HTTP client timeout is 5s.

## Processor and Sender

- **Semaphore:** buffered `chan struct{}` with capacity = `max_concurrent_messages` (`processor.max_concurrent_messages` / `sender.max_concurrent_messages` in `config.yaml`; default 50 when <= 0).
- The `Subscribe` callback attempts to acquire a slot; once acquired, work runs in a goroutine and releases the slot with `defer`.
- If no slot is available within 100ms, the message receives `Nak` (avoids blocking the NATS callback indefinitely).
- Each message runs with `context.WithTimeout(..., 10s)`.
- Manual ACK/NAK (JetStream) controls redelivery.
- **Metrics:** `processor_messages_concurrent`, `sender_messages_concurrent`.

## Notification-service

- **Fan-out:** one goroutine per provider, `WaitGroup` + channel to aggregate `ProviderResult` (`SendToAllProviders`).
- Consumes `messages.sent` via JetStream consumer with a 10s operation timeout.

## Limits

- Explicit retry/backoff policy at application layer is not implemented yet (redelivery currently comes from JetStream via `Nak` and consumer configuration).
