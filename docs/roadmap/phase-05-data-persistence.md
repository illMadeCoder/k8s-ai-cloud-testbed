## Phase 6: Data & Storage

*Stateful workloads: databases, caching, persistent storage, and disaster recovery.*

### 8.1 PostgreSQL with CloudNativePG

**Goal:** Operate PostgreSQL on Kubernetes with CloudNativePG

**Learning objectives:**
- Understand CloudNativePG operator
- Configure HA PostgreSQL clusters
- Implement backup and recovery

**Tasks:**
- [ ] Create `experiments/scenarios/postgres-tutorial/`
- [ ] Deploy CloudNativePG operator
- [ ] Create PostgreSQL cluster:
  - [ ] Primary + replicas
  - [ ] Synchronous replication
  - [ ] Connection pooling (PgBouncer)
- [ ] Configure storage:
  - [ ] PVC sizing and storage class
  - [ ] WAL archiving to object storage
- [ ] Backup and recovery:
  - [ ] Scheduled backups (to S3/Azure via Crossplane)
  - [ ] Point-in-time recovery (PITR)
  - [ ] Restore to new cluster
- [ ] Monitoring:
  - [ ] pg_stat metrics in Prometheus
  - [ ] Grafana dashboards
  - [ ] Alerting on replication lag
- [ ] Failover testing:
  - [ ] Kill primary, verify promotion
  - [ ] Measure failover time
- [ ] Document PostgreSQL operational patterns

---

### 8.2 Redis with Spotahome Operator

**Goal:** Operate Redis on Kubernetes for caching

**Learning objectives:**
- Understand Redis sentinel vs cluster mode
- Configure persistence and HA
- Implement caching patterns

**Tasks:**
- [ ] Create `experiments/scenarios/redis-tutorial/`
- [ ] Deploy Redis operator (Spotahome or similar)
- [ ] Create Redis deployments:
  - [ ] Standalone (development)
  - [ ] Sentinel (HA failover)
  - [ ] Cluster (horizontal scaling)
- [ ] Configure:
  - [ ] Persistence (RDB/AOF)
  - [ ] Memory limits and eviction
  - [ ] Password authentication
- [ ] Implement caching patterns:
  - [ ] Cache-aside
  - [ ] Write-through
  - [ ] Session storage
- [ ] Monitoring:
  - [ ] Redis metrics in Prometheus
  - [ ] Memory usage tracking
  - [ ] Hit/miss ratio
- [ ] Document Redis patterns for Kubernetes

---

### 8.3 Backup & Disaster Recovery

**Goal:** Implement comprehensive backup and DR strategy

*Requires: Phase 4.3 (MinIO) for backup storage*

**Learning objectives:**
- Understand Velero for cluster backup
- Implement cross-region DR patterns
- Design RTO/RPO strategies

**Tasks:**
- [ ] Create `experiments/scenarios/backup-dr-tutorial/`
- [ ] Deploy Velero:
  - [ ] Configure backup storage (S3/Azure Blob)
  - [ ] Install plugins (AWS, Azure, CSI)
- [ ] Implement backup strategies:
  - [ ] Full cluster backup
  - [ ] Namespace-scoped backup
  - [ ] Label-selected backup
  - [ ] Scheduled backups (hourly/daily)
- [ ] Test restore scenarios:
  - [ ] Restore to same cluster
  - [ ] Restore to different cluster
  - [ ] Partial restore (specific resources)
- [ ] Volume backup:
  - [ ] CSI snapshot integration
  - [ ] Restic for non-CSI volumes
- [ ] Cross-region DR:
  - [ ] Backup replication to secondary region
  - [ ] DR cluster provisioning (Crossplane)
  - [ ] Application failover procedure
- [ ] Document RTO/RPO for different scenarios
- [ ] Create DR runbook
- [ ] **ADR:** Document backup and DR strategy

---

### 8.4 Schema Migration Patterns

**Goal:** Manage database schema changes in Kubernetes deployments

**Learning objectives:**
- Understand schema migration tools
- Implement zero-downtime migrations
- Integrate with GitOps workflows

**Tasks:**
- [ ] Create `experiments/scenarios/schema-migration-tutorial/`
- [ ] Deploy migration tool:
  - [ ] Flyway OR Liquibase
- [ ] Implement migration patterns:
  - [ ] Init container migrations
  - [ ] Kubernetes Job migrations
  - [ ] ArgoCD pre-sync hook migrations
- [ ] Zero-downtime strategies:
  - [ ] Expand-contract pattern
  - [ ] Backward compatible changes
  - [ ] Blue-green database migrations
- [ ] Version management:
  - [ ] Migration versioning
  - [ ] Rollback strategies
  - [ ] Baseline migrations
- [ ] GitOps integration:
  - [ ] Migrations in Git
  - [ ] Sync wave ordering
  - [ ] Migration verification
- [ ] Document migration patterns
- [ ] **ADR:** Document schema migration strategy

---

### 8.5 Storage Cost Optimization

**Goal:** Optimize storage costs while maintaining performance and reliability

*FinOps consideration: Storage costs grow continuously. Implement lifecycle policies and tiering from the start.*

**Learning objectives:**
- Understand storage cost drivers in Kubernetes
- Implement storage tiering and lifecycle policies
- Balance performance, durability, and cost

**Tasks:**
- [ ] Storage class cost analysis:
  - [ ] Compare storage class costs (SSD vs HDD vs object)
  - [ ] IOPS/throughput requirements vs cost
  - [ ] Cloud storage tier comparison (Premium vs Standard)
- [ ] Lifecycle policies:
  - [ ] Data retention policies by age
  - [ ] Archive policies for cold data
  - [ ] Automatic tier migration
- [ ] Database cost optimization:
  - [ ] Right-sizing database instances
  - [ ] Read replica vs primary cost analysis
  - [ ] Reserved capacity for predictable workloads
- [ ] Backup cost management:
  - [ ] Backup retention cost modeling
  - [ ] Incremental vs full backup costs
  - [ ] Cross-region backup costs
- [ ] PVC right-sizing:
  - [ ] Identify over-provisioned PVCs
  - [ ] Volume expansion vs new provisioning
  - [ ] Unused PVC identification
- [ ] Document storage cost patterns
- [ ] **ADR:** Document storage tier and lifecycle strategy

---

