# Security

## Current state

- Strict payload validation in `ingestor`.
- Request body size limit on the ingestion endpoint.
- HTTP timeouts on control and load-generation services.

## Current gaps

- No authentication/authorization in `admin-api` and `simulator`.
- No endpoint rate limiting.
- No encryption for sensitive fields at application layer.

## Recommended next steps

1. Add authentication to control endpoints.
2. Add per-route rate limiting.
3. Add administrative action audit logging.
4. Manage secrets per environment.
