# ADR-001: Technology Choices for DDD-Oriented Distributed Learning Project

**Date:** April 24, 2024  
**Status:** Accepted

## Context

The project goal is to demonstrate distributed-systems architecture with DDD boundaries, event-driven integration, and operational observability in a portfolio-friendly setup.

Technology choices must support:

- explicit context boundaries
- asynchronous inter-context communication
- local reproducibility for interviews and demos
- low operational overhead

## Decision

Adopt:

- **Go** for core services (`ingestor`, `processor`, `sender`, `simulator`, `admin-api`)
- **NATS JetStream** for inter-context events
- **PostgreSQL** for durable lifecycle state
- **Prometheus + Grafana + Loki** for observability
- **Docker Compose** for local environment
- **K3s** for on-premises Kubernetes target

## Rationale

### Why this stack matches DDD goals

1. **Go**
   - Small binaries and clear service boundaries encourage context ownership.
   - Good concurrency model for event-driven workflows.

2. **NATS JetStream**
   - Subjects map naturally to context integration events.
   - Lightweight enough for local demonstrations.

3. **PostgreSQL**
   - Clear transactional model for lifecycle state and idempotency.
   - Easy to reason about in interviews.

4. **Observability stack**
   - Enables architecture discussion with evidence (metrics/logs), not only diagrams.

5. **Compose + K3s**
   - Fast local onboarding and production-like deployment narrative.

## Consequences

### Positive

- Strong architecture story for interview scenarios
- Clear path to ports/adapters and layered refactor
- Reproducible demos without vendor lock-in

### Negative

- Multiple services increase setup complexity
- Requires explicit schema/event contract governance

## Notes for Next Phase

The deep refactor should preserve these technology choices while aligning code structure to DDD layers (`domain`, `application`, `infrastructure`) inside each bounded context.
