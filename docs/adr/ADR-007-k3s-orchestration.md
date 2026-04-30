# ADR-007: Container Orchestration – K3s for On-Premises Kubernetes

**Date:** April 24, 2026  
**Status:** Accepted  
**Supersedes:** None  
**Superseded by:** None

## Context

The Traffic Simulator requires a container orchestration platform for:

- Running multiple services (Go services, PHP, NATS, PostgreSQL, etc.)
- Local development and production deployments
- Managing networking, storage, and scaling
- On-premises infrastructure control

The deployment strategy is **self-managed on-premises infrastructure**, not cloud-managed services. This requires:

- Full control over infrastructure
- Lightweight resource footprint
- Easy setup and management
- Support for both development and production

## Problem

Which Kubernetes distribution should we use for on-premises deployments that balances:

- **Resource efficiency** - On-premises servers have fixed capacity
- **Ease of deployment** - Simple to install and maintain
- **Feature completeness** - Must support full Kubernetes features
- **On-premises suitability** - Self-hosted, not cloud-locked
- **Learning value** - Good for portfolio (show Kubernetes expertise)

## Options Considered

### Option A: Vanilla Kubernetes

**Characteristics:**

- Official Kubernetes distribution
- Full feature set
- Maximum control
- Most documentation

**Setup Complexity:**

```bash
# Multiple steps: control plane, worker nodes, networking, storage
# Requires: kubeadm, kubelet, kube-apiserver, etcd management
# ~2GB+ RAM minimum per node
# Manual cluster bootstrap and scaling
```

**Pros:**

- ✓ Official, well-documented
- ✓ Maximum control
- ✓ Full feature set
- ✓ Industry standard

**Cons:**

- ❌ Resource-intensive (heavy on-premises burden)
- ❌ Complex cluster bootstrap (kubeadm, etcd management)
- ❌ Steep learning curve for operators
- ❌ Requires substantial CPU/RAM investment
- ❌ Operational overhead for on-premises
- ❌ Not optimized for small deployments

---

### Option B: Minikube

**Characteristics:**

- Single-node local development
- Lightweight
- Easy to learn
- Good for learning

**Setup Complexity:**

```bash
minikube start
# Single command, ~2GB RAM
# Local development only
```

**Pros:**

- ✓ Very easy to set up
- ✓ Low resource usage
- ✓ Good for learning

**Cons:**

- ❌ Single-node only (not scalable)
- ❌ Not suitable for production
- ❌ Not appropriate for on-premises multi-node cluster
- ❌ Development-only tool

---

### Option C: K3s (Chosen) ✅

**Characteristics:**

- Lightweight Kubernetes distribution (by Rancher)
- Optimized for on-premises and edge deployments
- ~512MB RAM minimum (vs ~2GB for vanilla K8s)
- Single binary with built-in etcd or external etcd support
- Production-grade but lightweight
- Used in production by many organizations

**Setup Complexity:**

```bash
# Server node
curl -sfL https://get.k3s.io | sh -

# Worker nodes
curl -sfL https://get.k3s.io | K3S_URL=https://server-ip:6443 K3S_TOKEN=xxx sh -

# Scales easily, ~30 seconds per node
```

**Pros:**

- ✅ **Lightweight** (~50% resource usage of vanilla Kubernetes)
- ✅ **Simple to deploy** (single binary, one command)
- ✅ **Perfect for on-premises** (designed for this use case)
- ✅ **Built-in etcd or external database support**
- ✅ **Low operational overhead**
- ✅ **Production-ready** (used in production)
- ✅ **Easy scaling** (add nodes with one command)
- ✅ **Full Kubernetes API compatibility**
- ✅ **Great for learning** (cleaner than vanilla K8s)
- ✅ **Excellent for portfolios** (shows modern DevOps practice)

**Cons:**

- Smaller community than vanilla Kubernetes
- Some advanced enterprise features may require customization
- Documentation slightly less extensive than vanilla K8s

---

## Decision

**We choose K3s (Rancher Kubernetes) for on-premises deployments**

### Rationale

1. **On-Premises Optimization**
   - K3s is specifically designed for on-premises and edge deployments
   - Minimal resource overhead (~50% vs vanilla Kubernetes)
   - Perfect fit for self-managed infrastructure
   - Easy to scale across server fleet

2. **Operational Simplicity**
   - Single binary deployment
   - Automatic etcd management (or bring-your-own database)
   - Simple node addition/removal
   - Reduced operational burden vs vanilla K8s

3. **Resource Efficiency**
   - ~512MB RAM minimum (vs 2GB+ for vanilla)
   - Lower CPU requirements
   - Ideal for cost-conscious on-premises deployments
   - Can run on commodity hardware (x86, ARM)

4. **Production-Ready**
   - Deployed in production by major organizations
   - Full Kubernetes API compatibility
   - Automatic updates and security patches available
   - High availability options (HA control plane)

5. **Learning & Portfolio Value**
   - Shows knowledge of Kubernetes variations
   - Demonstrates understanding of on-premises deployment patterns
   - Experience with lightweight Kubernetes valuable in industry
   - Modern DevOps approach (not just cloud-hosted)

6. **Full Feature Parity**
   - Same kubectl commands
   - Same manifests and configurations
   - Can migrate to vanilla K8s if needed (portable)
   - Integrated storage and networking (integrated Traefik ingress)

## Implementation

### K3s Architecture for Traffic Simulator

```
┌─────────────────────────────────────────┐
│        K3s Control Plane (HA)           │
│  - K3s server (etcd embedded)           │
│  - Kubernetes API server                │
│  - Scheduler, controller manager        │
└─────────────────────────────────────────┘
              ↓
┌─────────────────┬─────────────────┬──────────────┐
│  K3s Worker 1   │  K3s Worker 2   │ K3s Worker 3 │
│  - Go services  │  - Go services  │  - Go services│
│  - PHP service  │  - PostgreSQL   │  - Caching   │
│  - NATS cluster │  - Logging      │              │
└─────────────────┴─────────────────┴──────────────┘
```

### Installation (On-Premises)

**Server Setup:**

```bash
# Install K3s server (with etcd)
curl -sfL https://get.k3s.io | sh -

# High-availability setup (multiple servers)
curl -sfL https://get.k3s.io | \
  K3S_DATASTORE_ENDPOINT="postgresql://..." \
  K3S_TOKEN=your-secret-token \
  sh -
```

**Worker Setup:**

```bash
# Get token from server
TOKEN=$(ssh server "cat /var/lib/rancher/k3s/server/node-token")

# Install on worker nodes
curl -sfL https://get.k3s.io | \
  K3S_URL=https://server-ip:6443 \
  K3S_TOKEN=$TOKEN \
  sh -
```

### On-Premises Considerations

**Storage**

- Use local storage for PostgreSQL (fast SSDs recommended)
- Use external storage (NFS, Ceph) for distributed deployments
- K3s integrates with multiple storage backends

**Networking**

- Built-in Traefik ingress controller
- Flannel networking (default)
- Easy to customize network policies

**Monitoring**

- Deploy Prometheus and Grafana (same as cloud)
- K3s exposes standard metrics endpoints
- Full observability possible with Loki

**Scaling**

- Add worker nodes as needed
- K3s scales from 1 to 1000+ nodes
- Designed for heterogeneous hardware

## Migration & Extensibility

- **Same manifests work with vanilla Kubernetes** - Full API compatibility
- **Upgrade path to vanilla Kubernetes** if requirements grow
- **Multi-cluster setup** possible (K3s federation)
- **Air-gapped deployments** supported

## Related Decisions

- ADR-001: Technology stack choices
- ADR-006: Observability (works same on K3s as any K8s)

---

**Approved by:** Architecture Review Board  
**Date:** April 24, 2026  
**Review Date:** Quarterly

## Additional Resources

- [K3s Official Documentation](https://docs.k3s.io/)
- [K3s for On-Premises](https://docs.k3s.io/installation/network-options)
- [K3s HA Setup](https://docs.k3s.io/datastore/ha)
