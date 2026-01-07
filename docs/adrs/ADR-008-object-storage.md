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

**Architecture:**
- Erasure coding (Reed-Solomon) for durability across drives/nodes
- Metadata stored inline with data, no separate metadata service
- Each object = file on disk (directory scanning at millions of objects)
- Strong consistency (read-after-write guaranteed)

**Pros:**
- Best-in-class S3 API compatibility
- Erasure coding provides durability without replication overhead
- Massive ecosystem: docs, tutorials, integrations
- Battle-tested at scale

**Cons:**
- Entered maintenance mode (Dec 2025) - no new community features
- Admin UI removed from community edition
- Each object as file means LOSF problem at scale
- Enterprise push ($96k+/year)

**Best for:** Large objects, strict S3 compatibility needs, erasure coding durability

**Resource usage:** ~256MB-1GB RAM

### Option 2: Rook-Ceph (RGW)

**What it is:** Full storage platform with S3 gateway via RADOS Gateway.

**Architecture:**
- CRUSH algorithm distributes data (no central metadata bottleneck)
- Objects stored in OSDs with per-object overhead
- Strong consistency
- Unified block, file, and object from same cluster

**Pros:**
- Enterprise-grade, petabyte-scale proven
- Strong consistency guarantees
- Mixed workload support (block + object + file)
- CNCF graduated project

**Cons:**
- Very heavy (~3GB+ RAM for minimal cluster)
- Complex operations (dedicated team territory)
- Per-object overhead hurts small file performance
- Overkill for observability use case

**Best for:** Petabyte scale, mixed block/object/file workloads, enterprise ops teams

**Resource usage:** 3-6GB RAM minimum

### Option 3: SeaweedFS

**What it is:** Facebook Haystack-inspired object storage with O(1) lookups.

**Architecture:**
- Objects packed into ~30GB volume files (not individual files on disk)
- Index kept in memory: `file_id → volume_id → offset`
- Master tracks volumes, not individual files
- Filer adds optional directory/S3 semantics on top
- Replication-based durability (not erasure coding)

**Pros:**
- O(1) lookups regardless of file count (Haystack design)
- Excellent for billions of small files
- Simple architecture, easy to understand
- Active development, Apache 2.0 license
- Lower resource usage than MinIO

**Cons:**
- Less S3 feature coverage than MinIO
- Replication (not erasure coding) means more storage overhead
- Smaller community than MinIO (but growing)
- Eventual consistency by default

**Best for:** Billions of small files, read-heavy workloads, simpler operations

**Resource usage:** ~100-256MB RAM

### Option 4: Garage

**What it is:** Lightweight, geo-distributed object storage written in Rust.

**Architecture:**
- CRDT-based conflict resolution (Dynamo-style)
- Distributed hash table for metadata
- Designed for multi-site active-active deployments
- Eventual consistency (crdt-based)

**Pros:**
- Built for geo-distribution from the start
- Low resource requirements
- Self-healing across sites
- Modern Rust codebase

**Cons:**
- Younger project, less mature
- Partial S3 compatibility (improving)
- Smaller community
- Less documentation

**Best for:** Geo-distributed edge deployments, multi-site active-active, low-resource nodes

**Resource usage:** ~100-200MB RAM

### Option 5: Filesystem Backend (No Object Storage)

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

### Option 6: Cloud Provider (S3/GCS/Azure Blob)

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

## Technical Comparison

### Architecture Summary

| Solution | Data Layout | Metadata | Consistency | Durability |
|----------|-------------|----------|-------------|------------|
| **MinIO** | 1 object = 1 file | Inline | Strong | Erasure coding |
| **Ceph** | Objects in OSDs | Distributed (CRUSH) | Strong | Replication or EC |
| **SeaweedFS** | Packed into volumes | Master + in-memory index | Eventual | Replication |
| **Garage** | DHT-distributed | CRDT-based | Eventual | Replication |

### What Matters for Observability (Loki/Tempo/Thanos)

These tools write **medium-sized blocks** (~1-100MB), not billions of tiny files:

| Concern | Best Option | Notes |
|---------|-------------|-------|
| S3 API completeness | MinIO | Most complete implementation |
| Erasure coding durability | MinIO, Ceph | Less storage overhead than replication |
| Operational simplicity | SeaweedFS, Garage | Fewer moving parts |
| Geo-distribution | Garage | Built-in multi-site |
| Massive scale (PB+) | Ceph | Enterprise proven |
| Small files (if needed) | SeaweedFS | O(1) Haystack lookups |

**Bottom line:** For Loki/Tempo/Thanos specifically, MinIO works fine technically. SeaweedFS wins on simplicity and has an interesting architecture worth learning.

## Comparison Matrix

| Factor | MinIO | Ceph | SeaweedFS | Garage | Filesystem | Cloud S3 |
|--------|-------|------|-----------|--------|------------|----------|
| **RAM usage** | 256MB-1GB | 3-6GB | 100-256MB | 100-200MB | 0 | 0 |
| **S3 compatibility** | Excellent | Good | Good | Partial | N/A | Native |
| **Complexity** | Medium | High | Low | Low | None | Low |
| **Small file perf** | Degrades | Poor | Excellent | Good | N/A | Good |
| **Consistency** | Strong | Strong | Eventual | Eventual | N/A | Strong |
| **Durability model** | Erasure | Both | Replication | Replication | N/A | Erasure |
| **Geo-distribution** | Manual | Manual | Manual | Native | N/A | Native |
| **Thanos support** | Yes | Yes | Yes | Yes | No | Yes |
| **Kind suitable** | Yes | No | Yes | Yes | Yes | Yes |
| **Status (2025)** | Maint. mode | Active | Active | Active | N/A | N/A |

## Decision

**Use SeaweedFS** for self-hosted object storage in this lab.

### MinIO is Technically Fine

MinIO entered **maintenance mode** in December 2025:
- Community edition no longer accepting new features
- Admin UI removed, LDAP/OIDC removed
- Enterprise version costs $96k+/year

**However:** MinIO still works. Existing deployments are stable. For pure S3 compatibility, it remains the best option. The maintenance mode concern is about future trajectory, not current functionality.

### Why SeaweedFS for This Lab

| Factor | Reasoning |
|--------|-----------|
| **Architecture** | Haystack design is genuinely interesting and worth learning |
| **License** | Apache 2.0 (stable, permissive) vs MinIO's AGPL |
| **Tutorial value** | O(1 lookups, volume packing - concepts that transfer |
| **Active development** | New features still coming |
| **Simplicity** | Easier to understand than MinIO's erasure coding |

The tradeoff: SeaweedFS has less S3 feature coverage and uses replication (not erasure coding) for durability. For Loki/Tempo/Thanos, this doesn't matter.

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
- [Garage - Distributed S3 Storage](https://garagehq.deuxfleurs.fr/)
- [MinIO Maintenance Mode (Dec 2025)](https://www.infoq.com/news/2025/12/minio-s3-api-alternatives/)
- [MinIO vs SeaweedFS Comparison](https://github.com/chrislusf/seaweedfs/issues/1515)
- [Loki Storage](https://grafana.com/docs/loki/latest/storage/)
- [Thanos Storage](https://thanos.io/tip/thanos/storage.md/)
