# ADR-003: Worker Pools and Backpressure for Context Services

**Date:** April 24, 2024  
**Status:** Accepted

## Context

Core contexts (`processor`, `sender`) consume asynchronous events and perform I/O-heavy work. Unbounded concurrency would hide saturation and risk resource exhaustion.

The project needs predictable runtime behavior to support both learning and architecture demonstration.

## Decision

Adopt bounded worker-pool processing with explicit backpressure signals in event-consuming contexts.

## Rationale

1. **Operational predictability**
   - Bounded queues and worker counts control memory and connection usage.

2. **DDD alignment**
   - Application services consume events and coordinate domain operations within controlled throughput limits.

3. **Observability**
   - Queue depth, saturation, and retry counters become first-class indicators of architectural health.

4. **Interview-grade explainability**
   - Trade-off between throughput and safety is explicit and measurable.

## Consequences

### Positive

- Safer behavior under load spikes
- Clear backpressure semantics between contexts
- Easier capacity discussions and tuning

### Negative

- Requires tuning per workload profile
- Can reject or delay work when saturation thresholds are reached

## Implementation Direction for Refactor Phase

- Keep worker pools in infrastructure/application orchestration layers.
- Keep domain logic independent from concurrency primitives.
- Expose pool metrics through each service metrics endpoint.
