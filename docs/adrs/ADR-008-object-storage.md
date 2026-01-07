# ADR-008: Object Storage for Observability Backends

## Status

Accepted

## Context

Multiple observability and platform tools require S3-compatible object storage:
- **Loki** - Log chunk storage
- **Tempo** - Trace storage
- **Thanos** - Long-term metrics storage
- **Velero** - Cluster backup storage
- **Argo Workflows** - Artifact storage

Need to select an object storage solution for:
1. **Kind (local dev)** - Resource-constrained, ephemeral
2. **Talos (home lab)** - Persistent, limited hardware (N100)
3. **Cloud (future)** - Production workloads

### Requirements

| Requirement | Priority | Notes |
|-------------|----------|-------|
| S3-compatible API | Must | All tools use S3 protocol |
| Runs on Kubernetes | Must | GitOps managed |
| Resource efficient | High | Kind has ~8GB RAM total |
| Operator/Helm chart | High | Easy deployment |
| HA capable | Medium | For home lab |
| Learning value | Medium | Resume relevance |

## Options Considered

### Option 1: MinIO

**What it is:** Purpose-built S3-compatible object storage, cloud-native.

**Pros:**
- Industry standard for self-hosted S3
- Lightweight single-binary deployment
- Operator for Kubernetes (MinIO Operator)
- Excellent S3 compatibility
- Active development, good docs
- Can scale from single-node to distributed
- Native Prometheus metrics

**Cons:**
- Still requires resources (~512MB-1GB RAM minimum)
- Operator adds complexity for simple use cases
- License changed to AGPL (fine for internal use)

**Resource usage:** ~256MB-1GB RAM depending on mode

### Option 2: Rook-Ceph (RGW)

**What it is:** Full storage platform with S3 gateway via RADOS Gateway.

**Pros:**
- Enterprise-grade, battle-tested
- Unified block, file, and object storage
- Strong HA and replication
- CNCF graduated project

**Cons:**
- Very heavy (~3GB+ RAM for minimal cluster)
- Complex to operate
- Overkill for observability use case
- Requires multiple nodes for proper deployment

**Resource usage:** 3-6GB RAM minimum

### Option 3: SeaweedFS

**What it is:** Lightweight distributed file system with S3 gateway.

**Pros:**
- Very lightweight (~100MB RAM)
- Fast for small files
- Simple architecture
- S3 compatible

**Cons:**
- Less S3 feature coverage than MinIO
- Smaller community
- Less documentation
- Not as widely recognized

**Resource usage:** ~100-256MB RAM

### Option 4: Filesystem Backend (No Object Storage)

**What it is:** Use local PVCs instead of object storage.

**Pros:**
- Zero additional resources
- Simplest possible setup
- Good for ephemeral Kind clusters

**Cons:**
- Not all tools support it (Thanos requires S3)
- No S3 API learning
- Different config for dev vs prod
- Limited scalability

**Supported by:**
- Loki: Yes (filesystem mode)
- Tempo: Yes (local backend)
- Thanos: No (requires object storage)
- Velero: No (requires object storage)

### Option 5: Cloud Provider (S3/GCS/Azure Blob)

**What it is:** Use actual cloud object storage.

**Pros:**
- Zero cluster resources
- Production-ready
- Highly durable
- Pay-per-use

**Cons:**
- Requires cloud account and credentials
- Ongoing cost
- Network latency from local clusters
- Not self-contained lab

**Cost:** ~$0.023/GB/month (S3 Standard)

## Comparison Matrix

| Factor | MinIO | Rook-Ceph | SeaweedFS | Filesystem | Cloud S3 |
|--------|-------|-----------|-----------|------------|----------|
| **RAM usage** | 256MB-1GB | 3-6GB | 100-256MB | 0 | 0 |
| **S3 compatibility** | Excellent | Good | Good | N/A | Native |
| **Complexity** | Medium | High | Low | None | Low |
| **Thanos support** | Yes | Yes | Yes | No | Yes |
| **Velero support** | Yes | Yes | Yes | No | Yes |
| **Learning value** | High | High | Medium | Low | Medium |
| **Resume relevance** | High | High | Low | N/A | High |
| **HA capable** | Yes | Yes | Yes | No | Yes |
| **Kind suitable** | Yes | No | Yes | Yes | Yes |
| **Home lab suitable** | Yes | Marginal | Yes | No | No |

## Decision

**Use SeaweedFS** for self-hosted object storage.

### Why Not MinIO (December 2025)

MinIO entered **maintenance mode** in December 2025:
- Community edition no longer accepting new features
- Admin UI removed, LDAP/OIDC removed
- Enterprise version costs $96k+/year
- License: AGPL v3 (changed from Apache 2.0 in 2021)

This "rug pull" pattern makes MinIO unsuitable for long-term learning investments.

### Why SeaweedFS

| Factor | SeaweedFS |
|--------|-----------|
| **License** | Apache 2.0 (permissive, stable) |
| **Architecture** | Facebook Haystack-inspired, O(1) lookups |
| **Small files** | Excellent (packs into 32GB volumes) |
| **RAM usage** | 2-4GB (lighter than MinIO) |
| **Admin UI** | Included |
| **Enterprise** | $1/TB/month, first 25TB free |
| **Status** | Active development |

### Environment Strategy

| Environment | Solution | Rationale |
|-------------|----------|-----------|
| **Kind (tutorials)** | Filesystem | Zero resources for Loki/Tempo basics |
| **Kind (SeaweedFS tutorial)** | SeaweedFS | Learn the architecture |
| **Kind (Thanos/Velero)** | SeaweedFS | These require S3 |
| **Talos (home lab)** | SeaweedFS | Persistent, HA capable |
| **Cloud** | Native S3/GCS | Use managed services |

## Consequences

### Positive
- Apache 2.0 license - no future rug-pull risk
- Learning unique architecture (Haystack design)
- Lower resource usage than MinIO
- Active community and development

### Negative
- Less S3 feature parity than MinIO (but sufficient for our tools)
- Smaller community (but growing post-MinIO-maintenance-mode)
- Different architecture to learn (but that's educational)

### Migration Path
- SeaweedFS is S3-compatible
- Loki/Tempo/Thanos configs work with any S3 backend
- Can switch to cloud S3 without app changes

## References

- [SeaweedFS GitHub](https://github.com/seaweedfs/seaweedfs)
- [Facebook Haystack Paper](https://www.usenix.org/legacy/event/osdi10/tech/full_papers/Beaver.pdf)
- [MinIO Maintenance Mode (Dec 2025)](https://www.infoq.com/news/2025/12/minio-s3-api-alternatives/)
- [MinIO vs SeaweedFS Comparison](https://github.com/chrislusf/seaweedfs/issues/1515)
- [Loki Storage](https://grafana.com/docs/loki/latest/storage/)
- [Thanos Storage](https://thanos.io/tip/thanos/storage.md/)
