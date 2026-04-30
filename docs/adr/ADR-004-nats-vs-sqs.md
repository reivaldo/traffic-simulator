# ADR-004: Event Streaming Platform Selection

**Date:** April 24, 2024  
**Status:** Accepted  
**Supersedes:** None  
**Superseded by:** None

## Context

Event-driven architecture requires a message streaming system. After choosing event-driven patterns (ADR-002), we must evaluate leading event streaming platforms:

- **Apache Kafka** (heavy-weight, battle-tested, industry standard)
- **Amazon SQS** (managed cloud service, AWS ecosystem)
- **RabbitMQ** (message broker, mature, widely adopted)
- **NATS JetStream** (lightweight, cloud-native, modern)

Each platform has fundamentally different design philosophies and trade-offs.

## Problem

Which event streaming platform should we use for a **demonstration/learning project** that balances:

- **Portability** - Run locally, in any cloud, on Kubernetes
- **No vendor lock-in** - Avoid cloud provider dependency
- **Simplicity** - Easy to deploy and manage
- **Learning Value** - Educational for portfolio projects
- **Observability** - Transparent and debuggable
- **Cost** - Reasonable for learning/demo (no production scale costs)
- **Knowledge Demonstration** - Shows understanding of architecture trade-offs

## Options Considered

### Option A: Apache Kafka

**Characteristics:**

- Enterprise-grade event streaming platform
- Persistent distributed log architecture
- Multiple consumers per topic
- Cluster-based (Zookeeper/KRaft coordination)
- 1M+ messages per second
- Industry standard (Netflix, LinkedIn, Uber, etc.)

**Setup Complexity:**

```bash
# Requires: Kafka broker + ZooKeeper (or KRaft)
docker-compose up kafka zookeeper
# ~2-3GB memory minimum
# Complex cluster coordination
```

**Pros:**

- ✓ Industry-proven at massive scale
- ✓ Excellent throughput (1M+ msg/sec)
- ✓ Persistent replay capability
- ✓ Large ecosystem and community
- ✓ Mature monitoring/tooling
- ✓ Multiple consumer groups
- ✓ Exactly-once semantics available

**Cons:**

- ❌ **Complex setup and operations**
- ❌ **Resource-intensive** (JVM, 2-3GB memory)
- ❌ **Steep learning curve** (concepts like offsets, partitions)
- ❌ **Overkill for demo** project
- ❌ **Heavy operational burden**
- ❌ **Not suitable for learning** (hides abstractions)
- ❌ **Cluster coordination complexity** (Zookeeper/KRaft)
- ❌ **Less suitable for laptops**

**Learning Value:**

- ✓ Production patterns (proven battle-tested)
- ✗ Complex to reason about
- ✗ Many operational details
- ✗ Obscures core event-streaming concepts

**Production Readiness:**

- ✓ Battle-tested
- ✓ Scales to extreme scale
- ✓ Used by FAANG companies

---

### Option B: Amazon SQS

**Characteristics:**

- Managed queue service (not streaming platform)
- Fully serverless AWS service
- Automatic scaling
- AWS infrastructure managed
- Pay-per-request model
- Regional failover support

**Setup Complexity:**

```bash
# Create AWS account
# Create SQS queue via AWS Console or Terraform
# Configure IAM credentials
# Update application config
# No local development setup (must use AWS or LocalStack)
```

**Pros:**

- ✓ **Fully managed** (no ops burden)
- ✓ **Auto-scales transparently**
- ✓ **High availability** (AWS redundancy)
- ✓ **AWS ecosystem integration**
- ✓ **No infrastructure maintenance**
- ✓ **Battle-tested at scale**
- ✓ **Good for AWS-native orgs**

**Cons:**

- ❌ **AWS vendor lock-in** (not portable)
- ❌ **Can't run locally** (only LocalStack)
- ❌ **Costs money** even for demo/learning
- ❌ **Not suitable for learning** (hidden complexity)
- ❌ **No local development** (relies on AWS)
- ❌ **Less educational** (managed, not transparent)
- ❌ **Requires AWS credentials** locally
- ❌ **Not portable** to other clouds
- ❌ **Harder to understand** internals
- ❌ **Less suitable for portfolio** (AWS-specific)

**Learning Value:**

- ✗ Hides implementation details
- ✗ Creates AWS dependency
- ✗ Can't run offline
- ✗ Not portable to other clouds

**Production Readiness:**

- ✓ Battle-tested at massive scale
- ✓ Used by many enterprises
- ✗ Locks you into AWS
- ✗ Expensive at scale

---

### Option C: RabbitMQ

**Characteristics:**

- Open-source message broker
- Multiple exchange types (direct, topic, fanout)
- Advanced routing capabilities
- Built-in ack/nack semantics
- Mature and widely adopted
- Transient and persistent modes

**Setup Complexity:**

```bash
docker run -p 5672:5672 -p 15672:15672 rabbitmq:3-management
# ~100-200MB memory
# Simple single-node setup
```

**Pros:**

- ✓ **Mature and stable**
- ✓ **Familiar to many developers**
- ✓ **Good web UI** for management
- ✓ **Multiple exchange types** (flexible routing)
- ✓ **Easy to run locally**
- ✓ **Lower resource usage** than Kafka
- ✓ **Good community**
- ✓ **Open-source**

**Cons:**

- ❌ **Message broker, not streaming**
- ❌ **No built-in replay** capability
- ❌ **No consumer lag concept** (traditional acks)
- ❌ **Not durable by default** (messages lost on broker crash)
- ❌ **Less suitable for event sourcing**
- ❌ **Doesn't demonstrate streaming patterns**
- ❌ **Doesn't teach modern patterns** (events vs messages)
- ❌ **Not cloud-native** (designed for traditional deployments)
- ❌ **Weaker observability** for events

**Learning Value:**

- ✓ Good for understanding message brokers
- ✗ Different paradigm (not streaming)
- ✗ Doesn't teach event-driven patterns
- ✗ Obscures streaming concepts

**Production Readiness:**

- ✓ Proven and stable
- ✓ Good for message-based systems
- ✗ Not designed for event streaming
- ✗ Limited streaming capabilities

---

### Option D: NATS JetStream (Chosen) ✅

**Characteristics:**

- Cloud-native messaging platform
- Modern, minimal design philosophy
- Sub-millisecond latency
- High throughput (1M+ msg/sec)
- Single binary deployment
- Consumer groups with lag tracking
- Durable streams with replay
- Built-in metrics and observability

**Setup Complexity:**

```bash
docker run -p 4222:4222 nats:latest --jetstream
# Single command, ~100MB memory
# Runs locally, in Kubernetes, anywhere
# Zero cluster coordination needed
```

**Pros:**

- ✅ **Simple to deploy** (single binary)
- ✅ **Lightweight** (~100MB memory)
- ✅ **Cloud-native** (Kubernetes native)
- ✅ **Built-in consumer lag** (key observability!)
- ✅ **Auto retry with backoff**
- ✅ **Dead-letter queues** (pattern learning)
- ✅ **Easy to understand** (clear mental model)
- ✅ **Excellent for learning** projects
- ✅ **Low operational overhead**
- ✅ **Excellent documentation**
- ✅ **No vendor lock-in** (open-source, portable)
- ✅ **Shows modern architecture** (portfolio value)
- ✅ **Truly durable streams** (event sourcing capable)

**Cons:**

- Smaller ecosystem than Kafka
- Less widely known in enterprise
- Fewer production deployments (but rapidly growing)
- Less community resources vs Kafka

**Learning Value:**

- ✓ Event-driven patterns clear and understandable
- ✓ Consumer groups intuitive
- ✓ Lag metrics obvious
- ✓ Excellent for teaching streaming concepts
- ✓ Shows internal workings (not abstracted away)
- ✓ Portable across environments

**Production Readiness:**

- ✓ Fast-growing adoption
- ✓ Cloud Foundry uses it
- ✓ Can scale to high throughput
- ✓ Modern, actively developed
- ✓ Good for new projects

---

## Decision

**We choose Option D: NATS JetStream**

### Strategic Rationale

**Primary Goal: Demonstrate Architectural Knowledge**

This project serves as a **portfolio piece** to demonstrate understanding of distributed systems and architecture. By choosing NATS, we intentionally:

- ✅ **Show portability mindset** - Not locked into vendor (AWS SQS)
- ✅ **Demonstrate architecture knowledge** - Understand trade-offs between options
- ✅ **Avoid vendor lock-in** - System runs anywhere (laptop → Kubernetes → any cloud)
- ✅ **Show modern patterns** - NATS represents cloud-native thinking
- ✅ **Prove deployment skills** - Can run locally AND in production-like environments

**Why Not Other Options?**

| Why Not...   | Reason                                                                                        |
| ------------ | --------------------------------------------------------------------------------------------- |
| **Kafka**    | Overkill complexity for learning; hides concepts; JVM overhead; too heavy for laptops         |
| **SQS**      | Creates AWS dependency; can't demo locally; costs money; hidden complexity; not portable      |
| **RabbitMQ** | Message broker paradigm, not event streaming; doesn't teach modern patterns; not cloud-native |

### Technical Rationale

1. **Simplicity** (Key for Learning)
   - Single binary deployment
   - Zero cluster coordination overhead
   - Easy local development with docker-compose
   - Straightforward Kubernetes manifests
   - Clear mental model of streams and consumers

2. **Observability**
   - Consumer lag visible and intuitive
   - Can see pending messages per consumer
   - Built-in metrics (not require separate tools)
   - Perfect for teaching observability concepts
   - Demonstrable saturation patterns

3. **Resource Efficiency**
   - ~100MB memory (vs Kafka ~2GB, SQS managed)
   - Perfect for laptops and small environments
   - Can run multiple instances locally
   - Good for learning environments

4. **Educational Value**
   - Demonstrates real event-driven patterns
   - Clear examples of consumer groups and replay
   - Easy to explain retry/backoff mechanisms
   - Suitable for portfolio projects
   - Shows internal streaming concepts (not a black box)
   - Modern cloud-native design patterns

5. **Cloud-Native & Portable**
   - Kubernetes manifests simple and native
   - Stateless design (easily scalable)
   - Operator-friendly
   - Same code works on laptop, DC, GCP, Azure, AWS
   - Good for learning cloud patterns
   - Not locked into any cloud provider

## Implementation

### Docker Compose (Local Development)

```yaml
services:
  nats:
    image: nats:latest
    command: -js -m 8222
    ports:
      - "4222:4222" # Client
      - "8222:8222" # Monitoring
```

### Kubernetes (Production-like)

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: nats
spec:
  serviceName: nats
  replicas: 3
  template:
    spec:
      containers:
        - name: nats
          image: nats:latest
          args:
            - -c
            - /etc/nats/nats.conf
            - -js
```

### Stream Creation

```go
import "github.com/nats-io/nats.go"

nc, _ := nats.Connect("nats://localhost:4222")
js, _ := nc.JetStream()

// Create stream
js.AddStream(&nats.StreamConfig{
    Name:     "messages",
    Subjects: []string{"messages.incoming", "messages.processed"},
    Storage:  nats.FileStorage,
    MaxAge:   7 * 24 * time.Hour,
})

// Create consumer
js.AddConsumer("messages", &nats.ConsumerConfig{
    Durable:       "processor-consumer",
    DeliverPolicy: nats.DeliverAllPolicy,
    AckPolicy:     nats.AckExplicitPolicy,
    MaxDeliver:    5,
})
```

### Consumer Usage

```go
subscription, _ := js.PullSubscribe("messages.incoming", "processor")

for {
    msgs, _ := subscription.Fetch(100)  // Fetch batch
    for _, msg := range msgs {
        processMessage(msg.Data)
        msg.Ack()  // Explicit ack
    }
}
```

## Comprehensive Comparison Matrix

| Aspect                   | Kafka         | SQS            | RabbitMQ    | NATS      | Notes                       |
| ------------------------ | ------------- | -------------- | ----------- | --------- | --------------------------- |
| **Setup Complexity**     | ⭐⭐⭐⭐      | ⭐             | ⭐⭐        | ⭐        | NATS easiest                |
| **Local Development**    | Difficult     | Not native     | Good        | Excellent | NATS best for laptops       |
| **Memory Usage**         | ~2GB          | Managed        | ~200MB      | ~100MB    | NATS most lightweight       |
| **Latency**              | 1-10ms        | 100ms-1s       | <5ms        | <1ms      | NATS fastest                |
| **Throughput**           | 1M+ msg/s     | 1M+ msg/s      | 100K+ msg/s | 1M+ msg/s | Kafka & NATS similar        |
| **Consumer Lag**         | Via metrics   | Via CloudWatch | None        | Built-in  | NATS most transparent       |
| **Message Replay**       | ✓ Yes         | ✗ No (limited) | ✗ No        | ✓ Yes     | Critical for event sourcing |
| **Learning Curve**       | Steep         | Hidden details | Gentle      | Gentle    | NATS & RabbitMQ best        |
| **Vendor Lock-in**       | None          | AWS only       | None        | None      | SQS creates dependency      |
| **Production Proven**    | ✓ Yes (FAANG) | ✓ Yes (AWS)    | ✓ Yes       | ✓ Growing | All production-capable      |
| **Kubernetes Native**    | Via operator  | Via IAM        | Via charts  | Native    | NATS simplest               |
| **Cost (Demo)**          | Free\*        | Pay-per-use    | Free        | Free      | \*Requires ops              |
| **Portability**          | High          | AWS only       | High        | Highest   | NATS runs everywhere        |
| **Portfolio Value**      | Good          | Limited\*\*    | Good        | Excellent | \*\*AWS-specific only       |
| **Paradigm**             | Streaming     | Queue          | Messaging   | Streaming | NATS modern paradigm        |
| **Operational Overhead** | Medium        | Low            | Low         | Very Low  | NATS minimal ops            |

**Summary:**

- **Kafka:** Battle-tested, complex, overkill for learning
- **SQS:** AWS-locked, unsuitable for portfolio diversity
- **RabbitMQ:** Message broker (different paradigm), not streaming
- **NATS:** ✅ Best for portfolio learning project

## Observable Features

### Consumer Lag

**NATS (simple and direct):**

```bash
$ nats consumer info messages processor-consumer

Pending: 8234 messages
Next Delivery: sequence 12345
```

**Kafka (via JMX/Prometheus):**

```
lag = highWaterMark - committedOffset
# Requires separate monitoring setup (Prometheus, Grafana)
```

**RabbitMQ (limited via management):**

```bash
$ rabbitmqctl list_queues name messages consumers
# Queue depth only, no consumer lag concept
```

**SQS (requires CloudWatch API):**

```
# Must call AWS CloudWatch API
aws sqs get-queue-attributes --queue-url <url> --attribute-names ApproximateNumberOfMessages
# Requires AWS credentials and network call
```

### Monitoring Dashboard

**NATS:**

```
http://localhost:8222/streaming/general
# Built-in web UI, local access
```

**Kafka:**

```
Requires: Kafka UI, Kafdrop, or Prometheus
# Must set up separately
```

**RabbitMQ:**

```
http://localhost:15672
# Built-in web UI with credentials
```

**SQS:**

```
AWS Console (requires login)
# Cloud-only, no local dashboard
```

## Feature Comparison by Scenario

All four platforms support event-driven patterns, but with different capabilities:

| Scenario              | Kafka   | SQS     | RabbitMQ  | NATS   | Notes                        |
| --------------------- | ------- | ------- | --------- | ------ | ---------------------------- |
| **Replay messages**   | ✓ Yes   | ✗ No    | ✗ No      | ✓ Yes  | Critical for event sourcing  |
| **Consumer groups**   | ✓ Yes   | ✓ Yes   | ✓ Yes     | ✓ Yes  | All support it               |
| **Backpressure**      | ✓ Obs.  | ✓ Met.  | ✓ Obs.    | ✓ Obs. | NATS most transparent        |
| **Dead-letter queue** | ✓ DLQ   | ✓ DLQ   | ✓ Via ex. | ✓ DLQ  | All have mechanisms          |
| **Local testing**     | ✓ Hard  | ✗ AWS   | ✓ Easy    | ✓ Easy | RabbitMQ & NATS best         |
| **Multi-cloud**       | ✓ Yes   | ✗ AWS   | ✓ Yes     | ✓ Yes  | SQS AWS-only                 |
| **Cluster scaling**   | Complex | Managed | Medium    | Simple | NATS easiest                 |
| **Message ordering**  | ✓ Yes   | ✓ FIFO  | ✓ Yes     | ✓ Yes  | All support ordered delivery |

## Platform Characteristics Summary

### Kafka

- **Best for:** Large-scale, complex systems with proven ops
- **Worst for:** Learning, laptops, simplicity
- **Use when:** Company standardized on Kafka, massive scale needed

### SQS

- **Best for:** AWS-native ecosystems
- **Worst for:** Learning, portability, local development
- **Use when:** 100% committed to AWS, cost less concern

### RabbitMQ

- **Best for:** Message passing, traditional message broker patterns
- **Worst for:** Event streaming, replay capabilities
- **Use when:** Not doing event sourcing, traditional messaging

### NATS JetStream

- **Best for:** Learning, portability, cloud-native, event streaming
- **Worst for:** Very large enterprise deployments (though improving)
- **Use when:** Building modern, portable, educational systems

## Trade-offs

### NATS Advantages

- ✓ Portable (laptop → Kubernetes → any cloud)
- ✓ No vendor lock-in (crucial for demo/learning)
- ✓ Simpler mental model
- ✓ Easier local development
- ✓ Built-in consumer lag visibility
- ✓ Lower resource usage (~100MB)
- ✓ Cloud-native design
- ✓ Educational (shows internal patterns)

### SQS Advantages

- ✓ Fully managed (no ops)
- ✓ Auto-scales transparently
- ✓ AWS ecosystem integration
- ✓ Battle-tested at scale
- ✗ BUT: Creates AWS dependency
- ✗ BUT: Not portable
- ✗ BUT: Can't run locally
- ✗ BUT: Hidden complexity

**For a learning/portfolio project**: NATS wins decisively  
**For AWS-only production**: SQS is valid choice

## Abstraction & Migration Path

**Key Design Decision:** We abstract the event streaming layer

```go
// Interface-based (not NATS-specific)
type EventStreamer interface {
    Publish(topic string, message []byte) error
    Subscribe(topic string, handler func([]byte)) error
}

// Can be implemented as:
// - NATS adapter ← Current
// - Kafka adapter
// - SQS adapter
// - RabbitMQ adapter
// - Redis Streams adapter
```

**Future Migration Paths:**

If the project evolves:

1. **Staying Local** → Keep NATS (works on laptop and server)
2. **Moving to Kafka** → Implement `KafkaStreamer`, swap adapters
3. **Moving to AWS** → Implement `SQSStreamer`, swap adapters
4. **Multi-cloud** → Choose based on cloud provider

**No service code changes needed** - only adapter swap

This demonstrates good architectural practice: interface-based abstractions enable flexibility without vendor lock-in.

---

## Implementation Flexibility

The Traffic Simulator intentionally stays **vendor-neutral**:

```
Goal: Create a learning project that runs ANYWHERE

✅ Works on macOS/Linux/Windows laptop (NATS in Docker)
✅ Works in Docker Compose (complete local stack)
✅ Works in Kubernetes (cloud-agnostic manifests)
✅ Works in any cloud (GCP, Azure, on-premise)
✗ Does NOT require AWS account
✗ Does NOT depend on cloud infrastructure
```

This makes it suitable for:

- **Learning** (no AWS costs)
- **Portfolio** (shows multi-cloud thinking)
- **Interviews** (portable demo)
- **Open-source** (anyone can run it)

If you later need AWS integration, SQS adapter is straightforward.

---

## Related Decisions

- ADR-001: Technology choices (platform selection)
- ADR-002: Event-driven architecture (why events)
- ADR-003: Worker pools (consumer group behavior)

---

**Approved by:** Architecture Review Board  
**Date:** April 24, 2024  
**Review Date:** Quarterly
