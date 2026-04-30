# Observability

Current observability implementation.

## Metrics

All Go services expose:

- `GET /metrics`
- `GET /healthz`
- `GET /readyz`

Example metrics:

- `ingestor_messages_received_total`
- `processor_messages_processed_total`
- `sender_messages_sent_total`
- `simulator_messages_generated_total`
- `admin_api_requests_total`

## Logs

- Services use `logrus`.
- Logs are used for functional diagnostics and operational errors.

## Local stack

- Prometheus for metric scraping.
- Grafana for visualization.
- Loki/Promtail for log aggregation (when enabled in the environment).

## Current limits

- No end-to-end distributed tracing.
- No robust alert rules defined in application code.
