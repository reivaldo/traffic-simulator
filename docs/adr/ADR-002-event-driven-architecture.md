# ADR-002: Event-Driven Integration Between Bounded Contexts

**Date:** April 24, 2024  
**Status:** Accepted

## Context

The system has independent responsibilities:

- intake external intents
- process canonical messages
- deliver to outbound providers
- expose observability and controls

A synchronous chain would couple these responsibilities and weaken DDD boundaries.

## Decision

Use event-driven integration as the default communication pattern between core bounded contexts:

- `Message Intake Context` -> emits `MessageAccepted`
- `Message Processing Context` -> emits `MessageProcessed`
- `Message Delivery Context` -> emits `MessageSent`

Transport: NATS JetStream subjects (`messages.incoming`, `messages.processed`, `messages.sent`).

## Rationale

1. **Preserves context autonomy**
   - Each context evolves independently behind stable event contracts.

2. **Improves resilience**
   - Consumers can recover asynchronously without blocking upstream request paths.

3. **Supports distributed-system learning goals**
   - Makes backpressure, retries, and lag explicit and observable.

4. **Aligns with DDD strategic design**
   - Context map relationships are expressed as published language through events.

## Consequences

### Positive

- Clear integration seams for ports/adapters
- Reduced direct service-to-service coupling
- Better replay and diagnostics capabilities

### Negative

- Requires explicit event versioning discipline
- Requires idempotent consumers and lifecycle governance

## Governance Rules

- Event names must reflect ubiquitous language.
- Event payload changes must be backward-compatible or versioned.
- Contract-breaking changes require ADR update and migration plan.
