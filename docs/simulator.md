# Simulator

Documentation for the current behavior of the `simulator` service.

## Endpoints

- `GET /healthz`
- `GET /readyz`
- `GET /metrics`
- `GET /start?duration=<seconds>`
- `GET /stop`
- `GET /status`

## Load behavior

When started, the simulator runs a simple escalation profile:

- initial ramp-up
- peak phase
- recovery with residual rate

Default duration: `120` seconds (overridable via query string).

## Reliability and concurrency

- Current execution uses a per-run `context.Context`.
- `/stop` cancels only the active run.
- Sending to `ingestor` uses an `http.Client` with timeout.

## Exposed metrics

- `simulator_messages_generated_total{channel}`
- `simulator_current_rate{phase}`
- `simulator_running`
