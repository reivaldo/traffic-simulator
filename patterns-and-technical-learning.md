# Traffic Simulator - Patterns & Technical Learning

Proof of concept: distributed messaging pipeline with DDD, event-driven architecture, and observability.

---

## GLOBAL OVERVIEW

**Pipeline:**

```
simulator → ingestor → [NATS] → processor → [NATS] → sender → [NATS] → notification-service → template-service (HTTP /send)
```

*(Observability: Prometheus/Loki is configured via `docker-compose`; the diagram focuses on event flow.)*

**Characteristics:**

- Event-driven (at-least-once semantics)
- Bounded contexts (DDD): Intake | Processing | Delivery
- Persisted state machine (PostgreSQL)
- Observable (Prometheus + Loki)
- Context-based concurrency control
- Idempotent consumer pattern
- Parallelism patterns: Worker Pool (simulator), Semaphore (processor/sender), Fan-Out (notification-service)

---

## TECHNICAL LEARNING OUTCOMES

This section summarizes what was implemented and the technical rationale behind each decision.

- **DDD in practice (not only theory):** bounded contexts are explicit in code and separated by responsibilities and event contracts.
- **Engineering trade-offs are explicit:** NATS JetStream + manual ACK/NAK were chosen for a lightweight, testable learning architecture with at-least-once semantics.
- **Reliability by design:** idempotency is enforced in PostgreSQL via UPSERT and unique business key (`external_id`), reducing duplicate side effects.
- **Concurrency by constraint:** each pattern was applied to a different bottleneck:
  - Worker Pool (`simulator`) for throughput control.
  - Semaphore (`processor`/`sender`) for protecting shared resources under burst traffic.
  - Fan-Out (`notification-service`) for latency reduction in multi-target delivery.
- **Observability as a first-class concern:** metrics, health/readiness endpoints, and structured logs are integrated in each service.
- **Maintainability:** ports/adapters and dependency injection keep use cases testable and technology choices swappable.
- **Polyglot consistency (Go + PHP):** the PHP `template-service` follows the same architecture intent (value object + use case + port/adapter boundaries) used in Go services.
- **CI/CD readiness:** automated checks run on push/PR via GitHub Actions for both Go modules and PHP PHPUnit suite.

---

## 6 GO DESIGN PATTERNS IMPLEMENTED

### 1. Dependency Injection + Factory Pattern

**Purpose:** Decouple object creation from dependencies.

**Implementation in `src/ingestor/internal/bootstrap/bootstrap.go`:**

```go
func BuildHTTPHandler(nc *nats.Conn) http.Handler {
	// Layer 1: Infrastructure
	publisher := natsinfra.NewIntentPublisher(nc)

	// Layer 2: Application
	acceptUC := application.NewAcceptMessageIntent(publisher)

	// Layer 3: Interfaces
	handler := httpiface.NewHandler(acceptUC)

	return handler.Router()
}
```

**Benefit:** Each layer receives its dependencies. Mocking is trivial in tests.

**Application in `src/processor/internal/bootstrap/bootstrap.go`:**

```go
func Build(js nats.JetStreamContext, db *pgxpool.Pool, maxConcurrent int) (http.Handler, error) {
	repo := postgresinfra.NewMessageRepository(db)
	publisher := natsinfra.NewProcessedPublisher(js)
	useCase := application.NewProcessAcceptedMessage(repo, publisher)
	subscriber := eventsiface.NewSubscriber(
		useCase,
		eventsiface.SubscriberOptions{MaxConcurrentMessages: maxConcurrent},
	)
	if err := subscriber.Subscribe(js); err != nil {
		return nil, err
	}
	return httpiface.NewRouter(), nil
}
```

**Concept:** Composition and inversion of control without framework. Data flows down; control flows up.

---

## PHP SERVICE PATTERNS IMPLEMENTED (`template-service`)

The PHP service is not a placeholder anymore. It is part of the runtime delivery flow as the HTTP provider called by `notification-service`.

### Applied patterns

1. **Value Object (Domain validation)**
   - `src/template-service/src/Domain/TemplateRequest.php`
   - Validates required fields (`external_id`, `to`, `channel`) and normalizes channel input.

2. **Ports and Adapters**
   - Port: `src/template-service/src/Application/Ports/TemplateRendererPort.php`
   - Adapter: `src/template-service/src/Infrastructure/Renderer/SimpleTemplateRenderer.php`
   - Keeps rendering logic swappable without changing use-case orchestration.

3. **Use Case Orchestration**
   - `src/template-service/src/Application/RenderTemplate.php`
   - Encapsulates application flow and returns a stable delivery payload.

4. **HTTP Interface Layer**
   - `src/template-service/src/Interfaces/Http/TemplateController.php`
   - `src/template-service/src/Interfaces/Http/Router.php`
   - Handles transport concerns (`/healthz`, `/readyz`, `/send`) and delegates business logic to the use case.

### Test coverage (PHP)

- `src/template-service/tests/TemplateRequestTest.php`
- `src/template-service/tests/RenderTemplateTest.php`
- `src/template-service/tests/TemplateControllerTest.php`
- Runner: `vendor/bin/phpunit` (from `src/template-service`)

---

### 2. Interface Segregation & Ports (Hexagonal Architecture)

**Purpose:** Define minimal contracts, enable multiple implementations.

**Implementation in `src/ingestor/internal/application/ports.go`:**

```go
type IntentPublisherPort interface {
	PublishAccepted(ctx context.Context, intent domain.MessageIntent) error
}
```

**NATS implementation in `src/ingestor/internal/infrastructure/nats/publisher.go`:**

```go
type IntentPublisher struct {
	nc *nats.Conn
}

func (p *IntentPublisher) PublishAccepted(ctx context.Context, intent domain.MessageIntent) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	data, err := json.Marshal(map[string]any{
		"external_id": intent.ExternalID,
		"channel":     intent.Channel,
		"recipient":   intent.Recipient,
		"payload":     intent.Payload,
	})
	if err != nil {
		return err
	}
	if err := p.nc.Publish("messages.incoming", data); err != nil {
		return err
	}
	return p.nc.Flush()
}
```

**Benefit:** Application doesn't know about NATS. Swap Kafka/RabbitMQ without changing use cases.

**Same pattern in `processor`:**

```go
type MessageRepositoryPort interface {
	SaveProcessed(ctx context.Context, msg domain.MessageEnvelope) error
}

type ProcessedEventPublisherPort interface {
	PublishProcessed(ctx context.Context, msg domain.MessageEnvelope) error
}
```

**Concept:** Ports (interfaces) define protocol. Adapters (implementations) choose technology. Domain stays isolated.

---

### 3. Repository Pattern with Idempotency

**Purpose:** Abstract persistence; ensure idempotent operations via DB constraints.

**Implementation in `src/processor/internal/infrastructure/postgres/repository.go`:**

```go
func (r *MessageRepository) SaveProcessed(ctx context.Context, msg domain.MessageEnvelope) error {
	payload := map[string]any{}
	if len(msg.Payload) > 0 {
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}
	}

	// UPSERT: idempotent by design
	_, err := r.db.Exec(
		ctx,
		`INSERT INTO messages (external_id, channel, recipient, payload, status)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (external_id) DO UPDATE SET
		     channel = EXCLUDED.channel,
		     recipient = EXCLUDED.recipient,
		     payload = EXCLUDED.payload,
		     status = EXCLUDED.status,
		     updated_at = now()`,
		msg.ExternalID,
		msg.Channel,
		msg.Recipient,
		payload,
		"processed_pending_publish",
	)
	return err
}
```

**Key:** `external_id` has `UNIQUE` constraint. Redelivery = safe re-execution.

**Concept:** Idempotency implemented in DB (UPSERT), not in application. Repository abstracts complexity.

**Schema:**

Canonical DDL lives in `db/migrations/001_create_messages.sql`:

```sql
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id VARCHAR(128) NOT NULL UNIQUE,
    channel VARCHAR(32) NOT NULL,
    recipient VARCHAR(256) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'processed_pending_publish',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);
```

---

### 4. Value Objects with Smart Constructors

**Purpose:** Validation and encapsulation of domain rules at creation.

**Implementation in `src/ingestor/internal/domain/message_intent.go`:**

```go
type MessageIntent struct {
	ExternalID string
	Channel    string
	Recipient  string
	Payload    map[string]any
}

func NewMessageIntent(externalID, channel, recipient string, payload map[string]any) (MessageIntent, error) {
	// Normalization
	intent := MessageIntent{
		ExternalID: strings.TrimSpace(externalID),
		Channel:    strings.TrimSpace(channel),
		Recipient:  strings.TrimSpace(recipient),
		Payload:    payload,
	}
	if intent.Payload == nil {
		intent.Payload = map[string]any{}
	}

	// Validation
	if err := intent.Validate(); err != nil {
		return MessageIntent{}, err
	}

	return intent, nil
}

func (m MessageIntent) Validate() error {
	if m.ExternalID == "" {
		return errors.New("external_id is required")
	}
	if m.Channel == "" {
		return errors.New("channel is required")
	}
	if m.Recipient == "" {
		return errors.New("recipient is required")
	}
	return nil
}
```

**Benefit:** Impossible to create invalid `MessageIntent`. Validation happens ONCE.

**Concept:** Constructor enforces invariants. Type is guaranteed valid after construction.

**Usage in Handler:**

```go
func (h *Handler) handleMessages(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ExternalID string         `json:"external_id"`
		Channel    string         `json:"channel"`
		Recipient  string         `json:"recipient"`
		Payload    map[string]any `json:"payload"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Smart constructor: validate here
	intent, err := domain.NewMessageIntent(req.ExternalID, req.Channel, req.Recipient, req.Payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Now intent is GUARANTEED valid
	if err := h.acceptUseCase.Execute(r.Context(), intent); err != nil {
		http.Error(w, "failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
```

---

### 5. Context-Based Lifecycle & Cancellation

**Purpose:** Control concurrency, timeouts, and graceful shutdown in goroutines.

**Implementation in `src/simulator/main.go`:**

Per-run cancellation uses `runCtx` / `runCancel` (mutex-protected), created when `GET /start` runs. `escalationPattern(ctx, durationSeconds)` receives that context; the worker pool and send loop both respect `ctx.Done()`. HTTP uses a shared client with **5s timeout**:

```go
httpClient = &http.Client{Timeout: 5 * time.Second}
// ...
req, err := http.NewRequestWithContext(ctx, http.MethodPost, ingestorURL+"/messages", ...)
```

**Benefit:** Stopping a run cancels the escalation loop and workers; outbound calls remain bounded by the client timeout.

**Concept:** Context is the currency of modern Go. Always pass to functions that can block. Always respond to ctx.Done().

---

### 6. Worker Pool Pattern (Controlled Parallelism)

**Purpose:** Parallelize I/O-bound operations without creating uncontrolled N goroutines.

#### ❌ Problem: Naive Approach (Blocking)

**Previous approach (sequential send — removed):**

```go
// SEQUENTIAL SEND: blocks until response
secondSent := 0
for i := 0; i < rate; i++ {
	msg := generateMessage("")
	if err := sendMessageToIngestor(ctx, msg); err == nil {  // BLOCKS here
		atomic.AddInt64(&totalMessagesGenerated, 1)
		secondSent++
	}
}
time.Sleep(1 * time.Second)
```

**Problem:**

- At 320 msg/sec rate, each `sendMessageToIngestor` takes ~5-10ms (timeout or network latency)
- Total time = 320 × 10ms = 3.2 seconds **just for sending**
- Impossible to achieve 320 msg/sec sequentially

#### ✅ Solution: Worker Pool with Channels

**Implementation in `src/simulator/main.go`:**

```go
// WorkerPool manages N parallel goroutines for sending messages
type WorkerPool struct {
	workers      int
	queue        chan Message
	wg           sync.WaitGroup
	sentCounter  int64
	errorCounter int64
}

func NewWorkerPool(numWorkers int) *WorkerPool {
	return &WorkerPool{
		workers: numWorkers,
		queue:   make(chan Message, numWorkers*2), // buffer = 2x workers
	}
}

// Start initializes N goroutine workers
func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}
}

// worker processes messages from queue
func (wp *WorkerPool) worker(ctx context.Context, id int) {
	defer wp.wg.Done()

	for {
		select {
		case <-ctx.Done():
			logrus.Debugf("Worker %d stopping", id)
			return
		case msg, ok := <-wp.queue:
			if !ok {
				return
			}
			// Send is NON-BLOCKING for pool
			if err := sendMessageToIngestor(ctx, msg); err != nil {
				atomic.AddInt64(&wp.errorCounter, 1)
			} else {
				atomic.AddInt64(&wp.sentCounter, 1)
			}
		}
	}
}

// Submit adds message to queue
func (wp *WorkerPool) Submit(ctx context.Context, msg Message) error {
	select {
	case wp.queue <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Stop waits for all workers to finish
func (wp *WorkerPool) Stop() {
	close(wp.queue)
	wp.wg.Wait()
	logrus.Infof("Worker pool stopped. Sent: %d, Errors: %d",
		atomic.LoadInt64(&wp.sentCounter),
		atomic.LoadInt64(&wp.errorCounter))
}
```

**New code in `escalationPattern`:**

```go
func escalationPattern(ctx context.Context, durationSeconds int) {
	logrus.Infof("Starting escalation pattern for %d seconds", durationSeconds)

	// Create and start worker pool with 10 workers
	workerPool = NewWorkerPool(10)
	workerPool.Start(ctx)
	defer workerPool.Stop()

	// ... rest of code ...

	// Send messages via worker pool (NON-BLOCKING)
	secondSent := 0
	for i := 0; i < rate; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			msg := generateMessage("")
			if err := workerPool.Submit(ctx, msg); err == nil {
				atomic.AddInt64(&totalMessagesGenerated, 1)
				messagesGenerated.WithLabelValues(msg.Channel).Inc()
				secondSent++
			}
		}
	}
	time.Sleep(1 * time.Second)
}
```

**How it works:**

1. **Initialization:** 10 workers started in parallel, waiting on `queue` channel
2. **Submit:** Generation loop sends 320 messages to channel (non-blocking)
3. **Worker Loop:** Each worker extracts message from channel and executes `sendMessageToIngestor`
4. **Parallelism:** 10 workers can execute up to 10 sends in parallel (~100ms total for 320 msg on 10 goroutines)

**Benefits:**

- ✅ Achieving 320 msg/sec really is possible
- ✅ No goroutine per message (avoids memory leak)
- ✅ Easy to control "backpressure" with channel buffer
- ✅ Graceful shutdown via `close(queue)` + `Wait()`

**Metrics:**

```
Before:  320 msg × 10ms = 3.2s per second generated
After:   320 msg ÷ 10 workers = ~32 msg per worker = parallel
```

---

## GO PARALLELISM - ADVANCED PATTERNS (3 IMPLEMENTED)

### Pattern 1: Processor & Sender — Semaphore (Bounded Concurrency) ✅ IMPLEMENTED

**Problem:** NATS burst of 1000 messages → 1000 goroutines try DB → connection pool exhausts → timeouts → cascade failure.

**Solution:** Buffered channel as semaphore (`src/processor/internal/interfaces/events/subscriber.go`):

```go
type SubscriberOptions struct {
    MaxConcurrentMessages int
}

type Subscriber struct {
    useCase application.ProcessAcceptedMessagePort
    sem     chan struct{}  // Capacity = max concurrent (50)
}

func (s *Subscriber) Subscribe(js nats.JetStreamContext) error {
    _, err := js.Subscribe("messages.incoming", func(msg *nats.Msg) {
        select {
        case s.sem <- struct{}{}:        // Acquire slot (blocks if full)
            go s.processMessageAsync(msg)
        case <-time.After(100 * time.Millisecond):
            _ = msg.Nak()
        }
    }, nats.Durable("processor"), nats.ManualAck())
    return err
}

func (s *Subscriber) processMessageAsync(msg *nats.Msg) {
    defer func() { <-s.sem }()  // Release (guaranteed)
    processorConcurrentMessages.Inc()
    defer processorConcurrentMessages.Dec()
    // Process with 10s timeout
}
```

**Config:** `processor.max_concurrent_messages: 50` (tune per environment)  
**Metrics:** `processor_messages_concurrent`, `sender_messages_concurrent`  
**Same pattern in:** `src/sender/internal/interfaces/events/subscriber.go`

---

### Pattern 2: Notification-Service — Fan-Out to Providers ✅ IMPLEMENTED

**Problem:** SMS (200ms) + Email (150ms) + WhatsApp (180ms) = 530ms sequentially.

**Solution:** WaitGroup + Channel fan-out (`src/notification-service/internal/application/send_to_all_providers.go`):

```go
func (uc *SendToAllProviders) Execute(ctx context.Context, notification domain.Notification) ([]domain.ProviderResult, error) {
    var wg sync.WaitGroup
    resultChan := make(chan domain.ProviderResult, len(uc.gateways))
    
    // Spawn 1 goroutine per provider
    for providerName, gateway := range uc.gateways {
        wg.Add(1)
        go func(name string, gw ProviderGateway) {
            defer wg.Done()
            status, err := gw.Send(ctx, notification)
            resultChan <- domain.ProviderResult{Provider: name, Status: status, Error: err}
        }(providerName, gateway)
    }
    
    go func() {
        wg.Wait()
        close(resultChan)
    }()
    
    // Collect + evaluate: success = at least 1 provider succeeds
}
```

**Result:** Latency = max(200,150,180) = 200ms (3x faster!)  
**Metrics:** `notification_service_deliveries_total{status="success|error"}`

---

### Comparison: 3 Parallelism Patterns Implemented

| Pattern | Service | Bounded? | Goroutines | Use Case |
|---------|---------|----------|-----------|----------|
| **Worker Pool** | `simulator` | YES (10) | Pool+queue | Throughput: 320 msg/sec |
| **Semaphore** | `processor`, `sender` | YES (50) | Callbacks | Stability: protect DB |
| **Fan-Out** | `notification-service` | NO (3) | Per-target | Latency: parallel delivery |

---


## WHAT WAS LEARNED

### Architecture

1. **DDD in Go is practical.** Bounded contexts don't need frameworks. Just folder structure + interfaces.

2. **Event-driven reduces temporal coupling.** Broker holds events. Reprocessing guaranteed even with failure.

3. **NATS JetStream is lightweight for demo.** Kafka = heavy for learning. Both solve the same with different trade-offs.

4. **Idempotency in DB > Idempotency in application.** UPSERT with UNIQUE is more reliable than manual check.

5. **Explicit state machines ease debugging.** `processed_pending_publish` → `processed`. No state = black box.

6. **At-least-once semantics require duplicate handling.** Redelivery is guaranteed. Design for tolerance, not perfection.

7. **Controlled parallelism (worker pools) vs. goroutine per task.** Worker pool with N fixed workers reuses goroutines. Goroutine per task = memory leak at scale. Use when load is predictable.

### Go Patterns

1. **Constructor functions as validators.** `NewMessageIntent()` forces business rules. Prevents bugs before system.

2. **Small interfaces are powerful.** A function depending on `PublisherPort` is testable with mock.

3. **Context is mandatory in modern Go.** Thread-safe, allows timeout, allows cancellation. No reason not to use it.

4. **Manual Dependency Injection is simple.** Don't need Spring/Guice. Bootstrap package orchestrates creation.

5. **Error is a data type.** Error wrapping (`%w`) + classification (`errors.Is`) enables decisions by error type.

6. **Worker pools + channels = parallelism without memory leaks.** Don't create goroutine per item. Reuse workers with controlled buffer. Use `WaitGroup` for shutdown.

### Observability

1. **Prometheus (numbers) + Loki (text) is powerful.** Separate concerns: metrics = aggregated, logs = events. Correlation = insights.

2. **Fast healthz endpoints are critical.** >500ms = balancer might shut down by mistake.

3. **Metrics with labels enable drill-down.** `messages_received_total{status="error", channel="email"}` is more useful.

4. **Structured logs + context.Context = traceability.** Correlation IDs throughout stack.

### Operation

1. **Graceful shutdown is critical.** `pkill` or `docker stop` need to stop goroutines without losing state.

2. **Timeout on EVERY external call.** HTTP, NATS, DB: always timeout. Avoids hanging.

3. **Explicit error classification reduces MTTR.** `ErrTransientInfra` vs `ErrInvalidPayload` = different retry strategies.

---

## TECHNICAL ARCHITECTURE SUMMARY

### Layering (by service)

Paths are `src/<service>/internal/{domain,application,infrastructure,interfaces}/`.

```
domain/           → Value objects / structs
application/      → Use cases + ports (interfaces)
infrastructure/   → NATS, PostgreSQL adapters
interfaces/       → HTTP handlers, JetStream subscribers
```

### Data Flow (Ingestor)

```
HTTP Request
    ↓
src/ingestor/internal/interfaces/http/handlers.go
    ↓
src/ingestor/internal/domain/message_intent.go
    ↓
src/ingestor/internal/application/accept_message_intent.go
    ↓
src/ingestor/internal/infrastructure/nats/publisher.go → subject messages.incoming
    ↓
202 Accepted
```

### Data Flow (Processor)

```
NATS messages.incoming
    ↓
src/processor/internal/interfaces/events/subscriber.go
    ↓
src/processor/internal/application/process_accepted_message.go
    ↓
src/processor/internal/infrastructure/postgres/repository.go (UPSERT + MarkPublished)
    ↓
src/processor/internal/infrastructure/nats/publisher.go → messages.processed
```

### Contracts (NATS payloads today)

JSON matches `MessageEnvelope` / `MessageProcessed` (`external_id`, `channel`, `recipient`, `payload`). Status lives in PostgreSQL; it is not embedded in the published JSON structs.

```
messages.incoming  → same shape as HTTP body to ingestor
messages.processed → MessageEnvelope JSON from processor
messages.sent      → MessageProcessed JSON from sender
```

### Data Flow (Notification-service)

```
NATS messages.sent
    ↓
src/notification-service/internal/interfaces/events/subscriber.go
    ↓
src/notification-service/internal/application/send_to_all_providers.go (fan-out)
    ↓
HTTP calls to mock SMS / email / WhatsApp (parallel)
```

---

## GAPS & TECHNICAL BACKLOG

| Item                     | Why                                | Priority        |
| ------------------------ | ---------------------------------- | --------------- |
| **Transactional Outbox** | Resolve race condition DB/broker   | 🔴 CRITICAL     |
| **Error Classification** | Decide ACK/NAK by error type       | 🟡 IMPORTANT    |
| **Circuit Breaker**      | Protection against cascade failure | 🟡 IMPORTANT    |
| **Distributed Tracing**  | OpenTelemetry end-to-end           | 🟢 NICE-TO-HAVE |
| **Dead Letter Queue**    | Unrecoverable messages             | 🟢 NICE-TO-HAVE |
| **Schema Registry**      | Event versioning                   | 🟡 IMPORTANT    |

---

## DECISION POINTS

**Why PostgreSQL and not EventStore?**

- Event Sourcing = version entire history. Here only need current snapshot.
- UPSERT is simpler. Backlog: evaluate EventStore if queries get complex.

**Why NATS JetStream and not Kafka?**

- Kafka = distributed log. NATS = durable queue + stream.
- NATS fits in 1 container. Kafka needs Zookeeper + replicas.
- Learning = prefer lightweight. Production evaluates trade-offs.

**Why no distributed tracing today?**

- With a small set of Go services in this repo, Prometheus + Loki is sufficient before distributed tracing pays off.
- Jaeger = overhead without current value. Backlog for growing architecture.

**Why no auth today?**

- Local/demo environment. Production: oauth2 + RBAC.
- Ports ready for auth middleware.

**Why Worker Pool in Simulator?**

- Sequential sends to the ingestor wait for each HTTP round-trip; at high target rates the loop cannot keep up.
- A fixed pool (10 workers) overlaps I/O so the per-second generation loop can enqueue work without blocking on every response. Throughput still depends on ingestor capacity and network latency.

---

## RECOMMENDED STUDY

**Patterns to Deepen:**

1. **Transactional Outbox** - Resolve DB/broker inconsistency
   - Read: "Transactional Outbox" on microservices.io
   - Apply: Add `outbox_events` table + relay worker

2. **Saga Pattern** - Orchestration of distributed business flow
   - Read: O'Reilly "Building Microservices"
   - Relevance: When adding complex failure flow

3. **CQRS** - Command Query Responsibility Segregation
   - Deepen: When read scale becomes problem

4. **Event Sourcing** - Source of truth = immutable events
   - Deepen: When audit trail becomes critical

**Go Concepts:**

1. **Context and Cancellation** - Master propagation
2. **Error Handling** - Wrapped errors vs sentinel errors
3. **Interface Design** - When to segregate, when to combine
4. **Testing** - Table-driven tests, mocks with interfaces
5. **Channels & Goroutines** - Parallelism patterns

---

## KEY LEARNING POINTS & LESSONS

This project consolidated practical DDD in distributed systems: clear boundaries, explicit contracts, and technology-agnostic application logic.
Concurrency patterns are applied by constraint (throughput, stability, latency), not preference.
Reliability depends on operational discipline — idempotency, ACK/NAK semantics, observability, and graceful shutdown — while architecture quality depends on explicit trade-offs and testability.

- Value Objects protect domain invariants
- Ports separate domain from technology
- Dependency Injection reduces coupling
- Context is mandatory for concurrency
- Idempotency must be in DB, not in app
- Handlers validate at boundary
- Observability is not optional
- Error classification is strategic
- Graceful shutdown is critical
- Explicit state machines > implicit ones
- Worker pools solve controlled parallelism
- Channels with buffer prevent goroutine leaks
- `sync.WaitGroup` + context cancellation = clean coordination
- Semaphores cap concurrency (implemented in `processor` + `sender`)
- Fan-out parallelizes multi-target delivery (implemented in `notification-service`)
- Sequential throughput ≠ parallel throughput
