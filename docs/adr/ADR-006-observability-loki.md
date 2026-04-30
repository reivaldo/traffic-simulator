# ADR-006: Observability Stack – Prometheus + Grafana Loki

**Date:** April 24, 2026  
**Status:** Accepted  
**Supersedes:** None  
**Superseded by:** None

## Context

Observability is a critical requirement for the Traffic Simulator. The system must provide:

- **Metrics collection** - numeric data (latency, throughput, CPU, memory)
- **Log aggregation** - structured events and errors
- **Seamless integration** between metrics and logs (correlation)
- **Low operational overhead** and easy deployment
- **Cloud-native, scalable architecture**
- **Cost-effective** for both local development and production
- **On-premises compatible** (K3s cluster)

The challenge: **No single tool does both metrics AND logs well**. Therefore, we need:

1. A metrics solution (Prometheus)
2. A logging solution (Grafana Loki)
3. A visualization layer that unifies both (Grafana)

## Problem

How to implement observability that balances:

- **Completeness** - both metrics and logs
- **Cloud-native design** - works with Kubernetes/K3s
- **Resource efficiency** - lightweight for on-premises
- **Developer experience** - easy to run locally
- **Integration** - all tools work together seamlessly
- **Cost** - free/open-source
- **Portfolio value** - demonstrates modern patterns
- **Operational simplicity** - not too many moving parts

## Options Considered

### Option A: Comprehensive but Heavy

**ELK Stack (Elasticsearch, Logstash, Kibana) + Prometheus**

- ❌ **Too heavy**: ~4GB RAM just for ELK
- ❌ **Not integrated**: Prometheus + Kibana don't correlate
- ❌ **Complex**: Multiple components, hard to manage
- ❌ **Overkill**: Over-engineered for demo/learning
- ❌ **Not cloud-native**: Elasticsearch JVM awkward in K3s

**Not selected**: Too much overhead for on-premises

---

### Option B: Cloud Lock-in

**Cloud Provider Observability (DataDog, Splunk, CloudWatch)**

- ❌ **Vendor lock-in**: AWS/GCP/Azure dependent
- ❌ **Can't run locally**: Need cloud account
- ❌ **Costs money**: SaaS pricing
- ❌ **Not portable**: Tied to one provider
- ❌ **Not suitable**: For on-premises K3s cluster

**Not selected**: Contradicts on-premises strategy

---

### Option C: Modern, Lightweight, Integrated (Chosen) ✅

**Prometheus (metrics) + Grafana Loki (logs) + Grafana (dashboards)**

#### Prometheus: Metrics Collection

**What it does:**

- Collects numeric data (latency, throughput, error rates, resource usage)
- Time-series database for metrics
- Built-in alerting (AlertManager)
- ~100MB RAM footprint

**Why Prometheus:**

- ✅ Industry standard for metrics
- ✅ Native Kubernetes/K3s support
- ✅ Lightweight and cloud-native
- ✅ All Go services export Prometheus metrics by default
- ✅ Excellent alerting capabilities
- ✅ 15+ years of production use

#### Grafana Loki: Log Aggregation

**What it does:**

- Collects, indexes, and queries structured logs
- Label-based indexing (not full-text)
- ~100-300MB RAM footprint
- Simple operational model

**Why Loki (not ELK):**

- ✅ Cloud-native design (by Grafana Labs)
- ✅ 50% lighter than ELK (~300MB vs 4GB)
- ✅ Native Grafana integration (one dashboard for both)
- ✅ Works perfectly with Prometheus
- ✅ Label-based (efficient for structured logs)
- ✅ Horizontal scaling works smoothly
- ✅ Designed for Kubernetes/K3s environments

#### Grafana: Unified Dashboards

**What it does:**

- Visualizes both Prometheus metrics AND Loki logs
- Correlates logs and metrics on same timeline
- 400+ built-in dashboards
- Alerting rules visible alongside metrics

**Unified Stack:**

```
Prometheus (metrics)  ─────┐
                            ├─→ Grafana (single source of truth)
Grafana Loki (logs)   ─────┘

Both query the same timeline, same Grafana dashboard
```

### Pros (Prometheus + Loki)

- ✅ **Complete observability** - metrics + logs + dashboards
- ✅ **Cloud-native** - designed for Kubernetes/K3s
- ✅ **Lightweight** - ~100MB each, not ~4GB
- ✅ **Unified** - Grafana integrates both seamlessly
- ✅ **Open-source** - no vendor lock-in
- ✅ **Easy local dev** - Docker Compose setup
- ✅ **Powerful alerting** - from Prometheus
- ✅ **Great correlation** - logs + metrics on same dashboard
- ✅ **Industry standard** - proven in production
- ✅ **Portfolio value** - shows modern observability knowledge

### Cons

- Two tools to manage (but lightweight)
- Loki not full-text search (uses labels instead)
- Smaller community than ELK
- Newer than Elasticsearch/Kibana (but mature enough)

## Decision

**We choose Prometheus + Grafana Loki for observability**

### Implementation Architecture

```
┌──────────────────────────────────────────────────────┐
│             K3s Cluster (On-Premises)               │
├──────────────────────────────────────────────────────┤
│                                                      │
│  Service 1 (Go)     Service 2 (Go)     Service 3 (PHP) │
│      │                  │                   │        │
│      └──Prometheus──────┴───────────────────┘        │
│           Scrape every 15s (metrics)                 │
│                  ↓                                     │
│          ┌───────────────┐                           │
│          │  Prometheus   │                           │
│          │ Time-Series DB│ (~100MB RAM)              │
│          └───────────────┘                           │
│                  ↑                                     │
│      ┌──────────┴─────────┐                          │
│      │                    │                          │
│   Logs via         Metrics via                       │
│   Loki/Promtail    Prometheus                        │
│      │                    │                          │
│      ↓                    ↓                           │
│  ┌──────────┐       ┌──────────────┐               │
│  │  Loki    │       │  AlertManager│               │
│  │  (~200MB)│       │              │               │
│  └──────────┘       └──────────────┘               │
│      ↑                    ↑                          │
│      └────────┬───────────┘                          │
│               ↓                                      │
│         ┌──────────────┐                            │
│         │   Grafana    │                            │
│         │ (dashboards) │                            │
│         └──────────────┘                            │
│              ↑                                       │
│              │ (query both sources)                 │
│              │                                      │
│         http://localhost:3000                       │
│                                                      │
└──────────────────────────────────────────────────────┘
```

## Rationale

1. **Complete Observability**
   - Metrics alone can't tell you _why_ latency spiked (need logs)
   - Logs alone can't tell you system health (need metrics)
   - Both together = true root cause analysis

2. **On-Premises Optimization**
   - Lightweight enough for K3s cluster on commodity hardware
   - No JVM overhead (Go/Rust based)
   - Operational overhead minimal

3. **Kubernetes/K3s Native**
   - Prometheus: standard K8s monitoring
   - Loki: designed for containerized environments
   - Grafana: universal visualization

4. **Developer Experience**
   - Docker Compose setup with 3 containers
   - Fast startup (<30s)
   - Same in local dev and production

5. **Learning & Portfolio Value**
   - Modern observability stack (2024 best practices)
   - Demonstrates understanding of metrics vs logs
   - Loki knowledge valuable in industry
   - Correlation capability impresses interviewers

## Implementation Details

### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: "traffic-simulator"
    static_configs:
      - targets: ["localhost:8080"]

  - job_name: "processor"
    static_configs:
      - targets: ["localhost:8081"]

  - job_name: "sender"
    static_configs:
      - targets: ["localhost:8082"]
```

### Loki Configuration

```yaml
# loki-config.yml
auth_enabled: false

ingester:
  chunk_idle_period: 3m
  max_chunk_age: 1h
  chunk_retain_period: 1m

schema_config:
  configs:
    - from: 2024-01-01
      store: filesystem
      object_store: filesystem
      schema: v11
      index:
        prefix: loki_index_
        period: 24h

storage_config:
  filesystem:
    directory: /loki/chunks
```

### Grafana Dashboard Example

**Single Grafana dashboard showing:**

```
┌─────────────────────────────────────────────────┐
│  Traffic Simulator Dashboard                    │
├─────────────────────────────────────────────────┤
│                                                 │
│  [Prometheus] Requests/sec:  │  1,234 req/s   │
│  [Prometheus] Latency (p95): │  234ms         │
│  [Prometheus] Errors/sec:    │  12 err/s      │
│                                                 │
│  ─────────────────────────────────────────────  │
│                                                 │
│  [Loki] Recent Logs:                           │
│  10:30:01 [ERROR] "Template timeout"           │
│  10:30:02 [ERROR] "Circuit breaker opened"     │
│  10:30:05 [WARN]  "Queue saturation 95%"      │
│                                                 │
│  ─────────────────────────────────────────────  │
│                                                 │
│  Correlation: Spike in latency happened        │
│  exactly when logs show template timeouts!     │
│                                                 │
└─────────────────────────────────────────────────┘
```

## Related Decisions

- ADR-001: Technology stack (all decisions)
- ADR-005: PostgreSQL schema (stores event data)
- ADR-004: Event streaming (NATS event log source)
- ADR-007: K3s orchestration (where Prometheus + Loki run)

---

**Approved by:** Architecture Review Board  
**Date:** April 24, 2026  
**Review Date:** Quarterly
