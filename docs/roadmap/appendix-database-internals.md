## Appendix J: Database Internals

*Deep dive into how databases actually work — from storage engine fundamentals through distributed systems, operator selection, and benchmarking on Kubernetes. Understanding these transforms database usage from cargo-culting to informed, data-driven decisions.*

*See also: [Appendix E: Distributed Systems](appendix-distributed-systems.md) for theoretical foundations (consensus, CAP, clocks) and [Appendix I: Event-Driven Architecture](appendix-event-driven.md) for saga and outbox patterns in event-driven contexts.*

**Sections:**
| # | Topic | Focus |
|---|-------|-------|
| J.1–J.4 | Storage, Indexing, Queries, Transactions | Single-node fundamentals |
| J.5–J.6 | Replication, Sharding | Scaling fundamentals |
| J.7–J.9 | Connections, Backup, Observability | Operational fundamentals |
| J.10–J.12 | Distributed DB, Scaling Topologies, Sagas | Distributed data |
| J.13–J.14 | Category Selection, NewSQL | Database landscape |
| J.15–J.17 | K8s Operators, Managed vs Self-Hosted, Benchmarking | Kubernetes operations |

### J.1 Storage Engine Fundamentals

**Goal:** Understand how databases store and retrieve data

**Learning objectives:**
- Understand B-tree vs LSM-tree trade-offs
- Know how storage engines impact performance
- Choose appropriate storage engines for workloads

**Tasks:**
- [ ] Create `experiments/storage-engines/`
- [ ] Storage engine role:
  - [ ] Data organization on disk
  - [ ] Index structures
  - [ ] Read/write paths
  - [ ] Recovery mechanisms
- [ ] B-tree storage:
  - [ ] Tree structure and balancing
  - [ ] Page-based organization
  - [ ] In-place updates
  - [ ] Read-optimized characteristics
- [ ] B-tree operations:
  - [ ] Point lookups
  - [ ] Range scans
  - [ ] Insertions and splits
  - [ ] Deletions and merges
- [ ] LSM-tree storage:
  - [ ] Log-structured design
  - [ ] Memtable + SSTables
  - [ ] Compaction strategies
  - [ ] Write-optimized characteristics
- [ ] LSM-tree operations:
  - [ ] Write path (memtable → SSTable)
  - [ ] Read path (merge across levels)
  - [ ] Compaction (leveled, tiered, FIFO)
  - [ ] Bloom filters for efficiency
- [ ] Trade-offs:
  - [ ] Write amplification
  - [ ] Read amplification
  - [ ] Space amplification
  - [ ] Workload fit
- [ ] Storage engines in practice:
  - [ ] PostgreSQL (B-tree heap)
  - [ ] MySQL InnoDB (B-tree)
  - [ ] RocksDB (LSM-tree)
  - [ ] LevelDB (LSM-tree)
  - [ ] Cassandra (LSM-tree)
- [ ] Column stores:
  - [ ] Column-oriented storage
  - [ ] Compression benefits
  - [ ] Analytics workloads
  - [ ] ClickHouse, DuckDB
- [ ] **ADR:** Document storage engine selection criteria

---

### J.2 Indexing Deep Dive

**Goal:** Master database indexing for performance

**Learning objectives:**
- Understand index types and structures
- Design effective indexing strategies
- Troubleshoot index-related performance issues

**Tasks:**
- [ ] Create `experiments/database-indexing/`
- [ ] Index fundamentals:
  - [ ] Index as lookup structure
  - [ ] Primary vs secondary indexes
  - [ ] Clustered vs non-clustered
  - [ ] Index overhead (writes, storage)
- [ ] B-tree indexes:
  - [ ] Structure and navigation
  - [ ] Prefix compression
  - [ ] Index-only scans
  - [ ] Range query efficiency
- [ ] Hash indexes:
  - [ ] O(1) point lookups
  - [ ] No range query support
  - [ ] Memory-resident typically
  - [ ] Use cases
- [ ] Composite indexes:
  - [ ] Multi-column indexes
  - [ ] Column ordering importance
  - [ ] Leftmost prefix rule
  - [ ] Covering indexes
- [ ] Partial indexes:
  - [ ] Conditional indexing
  - [ ] Reduced storage
  - [ ] Query matching requirements
  - [ ] PostgreSQL partial indexes
- [ ] Expression indexes:
  - [ ] Indexing computed values
  - [ ] Function-based indexes
  - [ ] JSON path indexes
- [ ] Full-text indexes:
  - [ ] Text search capabilities
  - [ ] Inverted indexes
  - [ ] Stemming and tokenization
  - [ ] PostgreSQL tsvector
- [ ] Spatial indexes:
  - [ ] R-tree structure
  - [ ] Geographic queries
  - [ ] PostGIS
- [ ] Index maintenance:
  - [ ] Index bloat
  - [ ] Reindexing strategies
  - [ ] Statistics updates
  - [ ] Index monitoring
- [ ] Anti-patterns:
  - [ ] Over-indexing
  - [ ] Unused indexes
  - [ ] Wrong column order
  - [ ] Indexing low-cardinality columns
- [ ] Design indexing strategy for sample schema
- [ ] **ADR:** Document indexing guidelines

---

### J.3 Query Planning & Optimization

**Goal:** Understand how databases execute queries

**Learning objectives:**
- Read and interpret query plans
- Optimize slow queries
- Understand query optimizer decisions

**Tasks:**
- [ ] Create `experiments/query-optimization/`
- [ ] Query processing pipeline:
  - [ ] Parsing
  - [ ] Planning/optimization
  - [ ] Execution
- [ ] Query planner role:
  - [ ] Generate execution plans
  - [ ] Estimate costs
  - [ ] Choose optimal plan
- [ ] Plan operations:
  - [ ] Sequential scan
  - [ ] Index scan
  - [ ] Bitmap scan
  - [ ] Nested loop join
  - [ ] Hash join
  - [ ] Merge join
- [ ] EXPLAIN analysis:
  - [ ] Reading EXPLAIN output
  - [ ] EXPLAIN ANALYZE for actual times
  - [ ] Cost estimates vs actuals
  - [ ] Row estimate accuracy
- [ ] Statistics:
  - [ ] Table statistics
  - [ ] Column statistics (histograms)
  - [ ] Statistics freshness
  - [ ] ANALYZE command
- [ ] Join optimization:
  - [ ] Join order selection
  - [ ] Join algorithm selection
  - [ ] Join hints (when appropriate)
- [ ] Subquery optimization:
  - [ ] Correlated vs uncorrelated
  - [ ] Subquery flattening
  - [ ] Exists vs IN vs JOIN
- [ ] Common issues:
  - [ ] Missing indexes
  - [ ] Stale statistics
  - [ ] Type mismatches
  - [ ] Function calls preventing index use
- [ ] Query rewrites:
  - [ ] Equivalent transformations
  - [ ] Performance improvements
  - [ ] Maintaining correctness
- [ ] PostgreSQL-specific:
  - [ ] pg_stat_statements
  - [ ] auto_explain
  - [ ] Plan caching
- [ ] Optimize real-world queries
- [ ] **ADR:** Document query optimization process

---

### J.4 Transactions & Isolation

**Goal:** Understand transaction guarantees and isolation levels

**Learning objectives:**
- Understand ACID properties deeply
- Choose appropriate isolation levels
- Debug transaction-related issues

**Tasks:**
- [ ] Create `experiments/transactions/`
- [ ] ACID properties:
  - [ ] Atomicity (all or nothing)
  - [ ] Consistency (valid state transitions)
  - [ ] Isolation (concurrent transaction behavior)
  - [ ] Durability (committed = persistent)
- [ ] Isolation anomalies:
  - [ ] Dirty reads
  - [ ] Non-repeatable reads
  - [ ] Phantom reads
  - [ ] Write skew
  - [ ] Lost updates
- [ ] Isolation levels:
  - [ ] Read Uncommitted
  - [ ] Read Committed
  - [ ] Repeatable Read
  - [ ] Serializable
- [ ] Implementation approaches:
  - [ ] Locking (2PL)
  - [ ] MVCC (Multi-Version Concurrency Control)
  - [ ] Optimistic concurrency
- [ ] MVCC deep dive:
  - [ ] Version chains
  - [ ] Snapshot isolation
  - [ ] Garbage collection
  - [ ] Write conflicts
- [ ] PostgreSQL specifics:
  - [ ] Default Read Committed
  - [ ] Serializable Snapshot Isolation (SSI)
  - [ ] Transaction IDs
  - [ ] VACUUM necessity
- [ ] MySQL/InnoDB specifics:
  - [ ] Gap locking
  - [ ] Next-key locking
  - [ ] Deadlock detection
- [ ] Deadlocks:
  - [ ] Causes
  - [ ] Detection
  - [ ] Prevention strategies
  - [ ] Application handling
- [ ] Long transactions:
  - [ ] Problems caused
  - [ ] MVCC bloat
  - [ ] Lock contention
  - [ ] Best practices
- [ ] Demonstrate isolation anomalies
- [ ] **ADR:** Document isolation level selection

---

### J.5 Replication Internals

**Goal:** Understand database replication mechanisms

**Learning objectives:**
- Understand replication protocols
- Configure replication appropriately
- Handle replication lag and failures

**Tasks:**
- [ ] Create `experiments/db-replication/`
- [ ] Replication goals:
  - [ ] High availability
  - [ ] Read scaling
  - [ ] Geographic distribution
  - [ ] Disaster recovery
- [ ] Physical replication:
  - [ ] Byte-level log shipping
  - [ ] WAL (Write-Ahead Log) streaming
  - [ ] Block-level replication
  - [ ] Standby types (hot, warm)
- [ ] Logical replication:
  - [ ] Row-level changes
  - [ ] Schema flexibility
  - [ ] Selective replication
  - [ ] Cross-version replication
- [ ] PostgreSQL streaming replication:
  - [ ] WAL senders and receivers
  - [ ] Synchronous vs asynchronous
  - [ ] Replication slots
  - [ ] Cascading replication
- [ ] PostgreSQL logical replication:
  - [ ] Publications and subscriptions
  - [ ] Logical decoding
  - [ ] Conflict handling
- [ ] MySQL replication:
  - [ ] Binary log replication
  - [ ] GTID (Global Transaction ID)
  - [ ] Group Replication
  - [ ] Semi-synchronous replication
- [ ] Replication lag:
  - [ ] Causes
  - [ ] Monitoring
  - [ ] Impact on reads
  - [ ] Mitigation strategies
- [ ] Failover:
  - [ ] Automatic vs manual
  - [ ] Promotion process
  - [ ] Client reconnection
  - [ ] Split-brain prevention
- [ ] Replication tools:
  - [ ] Patroni (PostgreSQL HA)
  - [ ] Orchestrator (MySQL)
  - [ ] pg_basebackup
- [ ] Set up replication cluster
- [ ] **ADR:** Document replication architecture

---

### J.6 Sharding & Partitioning

**Goal:** Scale databases beyond single node

**Learning objectives:**
- Understand partitioning strategies
- Implement database sharding
- Handle cross-shard operations

**Tasks:**
- [ ] Create `experiments/db-sharding/`
- [ ] Partitioning vs Sharding:
  - [ ] Partitioning (single database)
  - [ ] Sharding (multiple databases)
  - [ ] When each applies
- [ ] Table partitioning:
  - [ ] Range partitioning
  - [ ] List partitioning
  - [ ] Hash partitioning
  - [ ] Composite partitioning
- [ ] PostgreSQL partitioning:
  - [ ] Declarative partitioning
  - [ ] Partition pruning
  - [ ] Partition maintenance
  - [ ] Partition-wise joins
- [ ] Sharding strategies:
  - [ ] Key-based (hash) sharding
  - [ ] Range-based sharding
  - [ ] Directory-based sharding
  - [ ] Geographic sharding
- [ ] Shard key selection:
  - [ ] High cardinality
  - [ ] Query patterns alignment
  - [ ] Even distribution
  - [ ] Access locality
- [ ] Cross-shard challenges:
  - [ ] Distributed queries
  - [ ] Distributed transactions
  - [ ] Referential integrity
  - [ ] Global sequences
- [ ] Sharding solutions:
  - [ ] Vitess (MySQL)
  - [ ] Citus (PostgreSQL)
  - [ ] Application-level sharding
  - [ ] ProxySQL routing
- [ ] Resharding:
  - [ ] Adding shards
  - [ ] Rebalancing data
  - [ ] Online resharding
  - [ ] Minimizing downtime
- [ ] NewSQL alternatives:
  - [ ] CockroachDB
  - [ ] TiDB
  - [ ] YugabyteDB
  - [ ] Automatic sharding
- [ ] Implement sharded database
- [ ] **ADR:** Document sharding strategy

---

### J.7 Connection Management

**Goal:** Optimize database connection handling

**Learning objectives:**
- Understand connection overhead
- Implement connection pooling
- Tune connection settings

**Tasks:**
- [ ] Create `experiments/connection-management/`
- [ ] Connection overhead:
  - [ ] TCP handshake
  - [ ] TLS negotiation
  - [ ] Authentication
  - [ ] Session initialization
  - [ ] Memory per connection
- [ ] Connection pooling:
  - [ ] Pool concepts
  - [ ] Pool sizing
  - [ ] Connection lifecycle
  - [ ] Pool exhaustion handling
- [ ] Application-level pooling:
  - [ ] HikariCP (Java)
  - [ ] pgx pool (Go)
  - [ ] SQLAlchemy pool (Python)
  - [ ] Configuration best practices
- [ ] External poolers:
  - [ ] PgBouncer
  - [ ] Pgpool-II
  - [ ] ProxySQL (MySQL)
- [ ] PgBouncer deep dive:
  - [ ] Pooling modes (session, transaction, statement)
  - [ ] Configuration
  - [ ] Limitations (prepared statements)
  - [ ] Monitoring
- [ ] Pool sizing:
  - [ ] Too small: queuing
  - [ ] Too large: resource exhaustion
  - [ ] Optimal sizing guidelines
  - [ ] Connections vs CPU cores
- [ ] Connection limits:
  - [ ] Database max_connections
  - [ ] Per-user limits
  - [ ] Per-database limits
  - [ ] Monitoring connection usage
- [ ] Kubernetes considerations:
  - [ ] Pod scaling impact
  - [ ] Sidecar poolers
  - [ ] Connection storms on restart
  - [ ] Graceful shutdown
- [ ] Troubleshooting:
  - [ ] Connection leaks
  - [ ] Pool exhaustion
  - [ ] Idle connection timeout
  - [ ] Connection validation
- [ ] Configure PgBouncer for Kubernetes
- [ ] **ADR:** Document connection pooling strategy

---

### J.8 Backup & Recovery

**Goal:** Implement reliable database backup strategies

**Learning objectives:**
- Understand backup types and trade-offs
- Implement point-in-time recovery
- Test recovery procedures

**Tasks:**
- [ ] Create `experiments/db-backup-recovery/`
- [ ] Backup types:
  - [ ] Logical backups (pg_dump, mysqldump)
  - [ ] Physical backups (file-level)
  - [ ] Continuous archiving (WAL)
  - [ ] Snapshots (storage-level)
- [ ] Logical backup:
  - [ ] SQL dump format
  - [ ] Custom format (parallel restore)
  - [ ] Selective backup (tables, schemas)
  - [ ] Restore process
- [ ] Physical backup:
  - [ ] pg_basebackup
  - [ ] File system backup
  - [ ] Consistent snapshots
  - [ ] Faster for large databases
- [ ] Point-in-Time Recovery (PITR):
  - [ ] Continuous WAL archiving
  - [ ] Base backup + WAL replay
  - [ ] Recovery target specification
  - [ ] Recovery timeline
- [ ] WAL archiving:
  - [ ] archive_command
  - [ ] WAL-G, pgBackRest
  - [ ] S3/GCS storage
  - [ ] Retention policies
- [ ] Backup tools:
  - [ ] pgBackRest (PostgreSQL)
  - [ ] Barman (PostgreSQL)
  - [ ] WAL-G (PostgreSQL, MySQL)
  - [ ] Percona XtraBackup (MySQL)
- [ ] Kubernetes backup:
  - [ ] Velero integration
  - [ ] Operator-based backup
  - [ ] Storage snapshots
  - [ ] Cross-cluster backup
- [ ] Recovery testing:
  - [ ] Regular restore tests
  - [ ] Recovery time measurement
  - [ ] Data validation
  - [ ] Runbook maintenance
- [ ] RPO and RTO:
  - [ ] Recovery Point Objective
  - [ ] Recovery Time Objective
  - [ ] Backup frequency alignment
  - [ ] Architecture implications
- [ ] Implement PITR with pgBackRest
- [ ] **ADR:** Document backup strategy

---

### J.9 Database Observability

**Goal:** Monitor and troubleshoot database performance

**Learning objectives:**
- Instrument database monitoring
- Identify performance issues
- Build effective dashboards

**Tasks:**
- [ ] Create `experiments/db-observability/`
- [ ] Key metrics:
  - [ ] Query throughput (QPS)
  - [ ] Query latency (p50, p99)
  - [ ] Connection count
  - [ ] Cache hit ratios
- [ ] PostgreSQL statistics:
  - [ ] pg_stat_user_tables
  - [ ] pg_stat_user_indexes
  - [ ] pg_stat_activity
  - [ ] pg_stat_statements
- [ ] pg_stat_statements:
  - [ ] Query fingerprinting
  - [ ] Execution statistics
  - [ ] Top queries by time
  - [ ] Query plan changes
- [ ] Wait events:
  - [ ] pg_stat_activity.wait_event
  - [ ] Lock waits
  - [ ] I/O waits
  - [ ] CPU waits
- [ ] Lock monitoring:
  - [ ] pg_locks view
  - [ ] Lock conflicts
  - [ ] Deadlock detection
  - [ ] Blocking queries
- [ ] Replication monitoring:
  - [ ] pg_stat_replication
  - [ ] Replication lag
  - [ ] Slot status
  - [ ] WAL generation rate
- [ ] Storage monitoring:
  - [ ] Table and index sizes
  - [ ] Bloat estimation
  - [ ] Disk usage trends
  - [ ] VACUUM monitoring
- [ ] Prometheus exporters:
  - [ ] postgres_exporter
  - [ ] mysqld_exporter
  - [ ] Custom queries
- [ ] Grafana dashboards:
  - [ ] Overview dashboard
  - [ ] Query performance dashboard
  - [ ] Replication dashboard
  - [ ] Alert configuration
- [ ] Log analysis:
  - [ ] Slow query log
  - [ ] Error log patterns
  - [ ] Log aggregation
- [ ] Build comprehensive monitoring
- [ ] **ADR:** Document database monitoring strategy

---

### J.10 Distributed Database Fundamentals

**Goal:** Understand the foundational mechanisms that distributed databases use to coordinate state across nodes

**Learning objectives:**
- Understand gossip protocols, anti-entropy, and vector clocks as building blocks for distributed databases
- Evaluate consistency models and their practical impact on application behavior
- Configure quorum reads/writes and reason about the consistency spectrum

**Tasks:**
- [ ] Create `experiments/distributed-db-fundamentals/`
- [ ] Gossip protocols:
  - [ ] Epidemic-style information dissemination
  - [ ] Push, pull, and push-pull gossip
  - [ ] Convergence time and message overhead
  - [ ] Gossip in practice (Cassandra, CockroachDB, Consul)
- [ ] Anti-entropy mechanisms:
  - [ ] Merkle trees for replica divergence detection
  - [ ] Read repair (detect stale data on read)
  - [ ] Hinted handoff (buffer writes for unavailable nodes)
  - [ ] Active anti-entropy vs passive repair
- [ ] Vector clocks and causality:
  - [ ] Lamport timestamps vs vector clocks
  - [ ] Detecting concurrent writes
  - [ ] Vector clock size growth and pruning
  - [ ] Dotted version vectors (Riak optimization)
- [ ] Conflict resolution strategies:
  - [ ] Last-write-wins (LWW) and timestamp pitfalls
  - [ ] CRDTs (G-Counter, OR-Set, LWW-Register)
  - [ ] Application-level resolution callbacks
  - [ ] Sibling values and client-side merge (Riak model)
- [ ] Consistency models in databases:
  - [ ] Strong consistency (linearizable reads/writes)
  - [ ] Eventual consistency (replicas converge)
  - [ ] Causal consistency (respects causality)
  - [ ] Session guarantees (read-your-writes, monotonic reads)
- [ ] Quorum reads and writes:
  - [ ] R + W > N rule for strong consistency
  - [ ] Tunable consistency (Cassandra consistency levels)
  - [ ] Sloppy quorums and availability trade-off
  - [ ] Quorum latency impact (wait for slowest replica)
- [ ] Dynamo-style architecture:
  - [ ] Consistent hashing with virtual nodes
  - [ ] Preference list and coordinator node
  - [ ] Put/Get paths through coordinator
  - [ ] Comparison: DynamoDB, Cassandra, Riak, ScyllaDB
- [ ] Consistency spectrum decision framework:
  - [ ] Mapping use cases to consistency levels (payments vs likes)
  - [ ] PACELC trade-offs (latency vs consistency during normal operation)
  - [ ] Benchmarking latency impact of stronger consistency
  - [ ] Cost of coordination: throughput at each consistency level
- [ ] Benchmark consistency trade-offs:
  - [ ] Deploy Cassandra on Kubernetes with different consistency levels
  - [ ] Measure throughput and latency at ONE, QUORUM, ALL
  - [ ] Simulate node failure and observe read/write behavior
- [ ] **ADR:** Document consistency model selection for workload types

---

### J.11 Scaling Topologies: Replicas, Shards & Replicas of Shards

**Goal:** Understand the four fundamental scaling topologies and when to apply each based on workload characteristics

**Learning objectives:**
- Distinguish between replication-only, sharding-only, and combined topologies
- Apply a decision framework based on read/write ratio, data volume, and availability requirements
- Perform capacity planning for each topology on Kubernetes

**Tasks:**
- [ ] Create `experiments/db-scaling-topologies/`
- [ ] Single primary + read replicas:
  - [ ] Architecture: one writer, N readers
  - [ ] Replication lag and read consistency trade-offs
  - [ ] Read scaling ceiling (when replicas aren't enough)
  - [ ] Kubernetes deployment: CloudNativePG with read replicas
- [ ] Multi-primary replication:
  - [ ] Architecture: multiple nodes accept writes
  - [ ] Conflict detection and resolution mechanisms
  - [ ] Use cases: multi-region active-active, low-latency writes
  - [ ] Examples: MySQL Group Replication, Galera Cluster, CockroachDB
- [ ] Horizontal sharding without replicas:
  - [ ] Architecture: data partitioned across nodes, no redundancy per shard
  - [ ] When acceptable (batch processing, rebuildable data)
  - [ ] Risk profile: single shard failure loses data
  - [ ] Examples: Vitess single-replica shards, application-level sharding
- [ ] Replicas of shards (combined pattern):
  - [ ] Architecture: each shard has its own replica set
  - [ ] Read scaling AND write scaling simultaneously
  - [ ] Operational complexity: N shards x M replicas
  - [ ] Examples: MongoDB sharded cluster, Vitess with replicas, CockroachDB ranges
- [ ] Topology decision framework:
  - [ ] Read-heavy workloads (>10:1 read/write ratio) → replicas first
  - [ ] Write-heavy or large datasets → sharding first
  - [ ] Both → replicas of shards
  - [ ] Decision tree: data size, QPS, availability SLA, operational maturity
- [ ] Capacity planning:
  - [ ] Estimating storage per shard (data size / shard count)
  - [ ] Estimating connections (pods × pool size × shards)
  - [ ] IOPS requirements per node (storage class dependent)
  - [ ] Network bandwidth between replicas
- [ ] Kubernetes resource modeling:
  - [ ] CPU and memory per database pod
  - [ ] PVC sizing and storage class selection
  - [ ] Node affinity and anti-affinity for HA
  - [ ] Pod disruption budgets for maintenance
- [ ] Topology migration paths:
  - [ ] Single node → replicas (add streaming replication)
  - [ ] Replicas → sharded replicas (introduce Vitess/Citus)
  - [ ] Sharding after the fact: data migration strategies
  - [ ] Online resharding without downtime
- [ ] Benchmark scaling topologies:
  - [ ] Deploy single primary, then add replicas, measure read throughput scaling
  - [ ] Deploy sharded setup, measure write throughput scaling
  - [ ] Compare operational overhead (failover time, backup complexity)
- [ ] **ADR:** Document scaling topology selection and migration plan

---

### J.12 Saga Patterns & Distributed Transactions

**Goal:** Implement and benchmark distributed transaction patterns across databases on Kubernetes

**Learning objectives:**
- Understand 2PC/3PC limitations and why sagas are preferred in microservices
- Implement saga orchestration and choreography with compensating transactions
- Compare saga patterns against distributed database transactions for correctness and performance

**Tasks:**
- [ ] Create `experiments/saga-patterns/`
- [ ] Two-Phase Commit (2PC) in databases:
  - [ ] Prepare and commit phases
  - [ ] Coordinator as single point of failure
  - [ ] Blocking problem: participant in-doubt state
  - [ ] Performance cost: lock duration across prepare/commit
- [ ] Three-Phase Commit (3PC):
  - [ ] Pre-commit phase to reduce blocking
  - [ ] Non-blocking under certain failure assumptions
  - [ ] Why rarely used: network partitions violate assumptions
  - [ ] Comparison: 2PC blocking vs 3PC complexity
- [ ] The cost of 2PC — latency and throughput analysis:
  - [ ] Network round-trips: minimum 2 RTTs (prepare + commit) per participant
  - [ ] Lock hold duration: rows locked from prepare until commit/abort across all participants
  - [ ] Coordinator log fsync: forced WAL flush at prepare and commit (2 fsyncs on coordinator)
  - [ ] Participant log fsync: forced WAL flush at prepare (1 fsync per participant)
  - [ ] Tail latency amplification: transaction latency = slowest participant + coordinator
  - [ ] Throughput ceiling: lock contention grows non-linearly with participant count
  - [ ] Failure penalty: in-doubt state holds locks until coordinator recovers (minutes to hours)
  - [ ] Cross-datacenter 2PC: RTT penalty makes prepare phase 10-100x slower than local
- [ ] Benchmarking 2PC cost on Kubernetes:
  - [ ] Deploy PostgreSQL with `postgres_fdw` for cross-database 2PC (`PREPARE TRANSACTION`)
  - [ ] Measure single-node transaction vs 2-participant vs 3-participant latency
  - [ ] Measure throughput degradation as participant count increases (2, 3, 5 nodes)
  - [ ] Simulate coordinator crash mid-prepare: measure lock hold duration and recovery time
  - [ ] Compare 2PC latency: same-node pods vs cross-node vs cross-cluster
  - [ ] Quantify the "2PC tax": overhead ratio vs local single-database transaction
- [ ] Why systems avoid 2PC:
  - [ ] Google Spanner: uses TrueTime + Paxos instead of 2PC for most transactions
  - [ ] CockroachDB: parallel commits optimization to reduce 2PC to 1 RTT in common case
  - [ ] Kafka transactions: optimistic 2PC with async commit (different trade-off)
  - [ ] Microservices: sagas preferred because 2PC couples availability of all participants
- [ ] Saga orchestration:
  - [ ] Central saga execution coordinator (SEC)
  - [ ] Explicit state machine defining step sequence
  - [ ] Command/reply communication with participants
  - [ ] Timeout handling and step retries
- [ ] Saga choreography:
  - [ ] Event-driven step triggering (no central coordinator)
  - [ ] Each service publishes events, next service reacts
  - [ ] Decentralized control flow
  - [ ] Harder to trace and debug end-to-end
- [ ] Compensating transactions:
  - [ ] Semantic undo (not always exact inverse)
  - [ ] Ordering: compensate in reverse order of execution
  - [ ] Idempotent compensations (retries must be safe)
  - [ ] Non-compensable steps: pivot transactions and countermeasures
- [ ] Outbox pattern for reliable messaging:
  - [ ] Write business data + outbox record in same transaction
  - [ ] Separate publisher reads outbox and emits events
  - [ ] Debezium CDC as alternative to polling
  - [ ] Guarantees: at-least-once delivery with consumer idempotency
- [ ] Idempotency implementation:
  - [ ] Idempotency keys (client-generated unique IDs)
  - [ ] Server-side deduplication table
  - [ ] Natural idempotency (PUT vs POST semantics)
  - [ ] Time-windowed deduplication and key expiration
- [ ] Failure modes and recovery:
  - [ ] Participant failure during forward execution
  - [ ] Coordinator failure during compensation
  - [ ] Network partition mid-saga
  - [ ] Manual intervention and dead-letter handling
- [ ] Distributed database transactions comparison:
  - [ ] CockroachDB serializable transactions (Raft-based)
  - [ ] TiDB distributed transactions (Percolator model)
  - [ ] When a distributed database eliminates saga need
  - [ ] Trade-off: operational complexity vs application complexity
- [ ] Benchmark saga vs distributed transaction:
  - [ ] Deploy saga-based order flow (PostgreSQL + message broker)
  - [ ] Deploy same flow on CockroachDB with distributed transactions
  - [ ] Measure latency, throughput, and failure recovery time
  - [ ] Compare operational overhead on Kubernetes
- [ ] **ADR:** Document distributed transaction strategy selection

---

### J.13 Database Category Selection & Comparison

**Goal:** Evaluate database categories by deploying representative systems on Kubernetes and benchmarking their strengths

**Learning objectives:**
- Understand the strengths and trade-offs of each major database category
- Deploy and benchmark representative databases from each category on Kubernetes
- Apply a decision framework to select the right database for a given workload

**Tasks:**
- [ ] Create `experiments/db-category-comparison/`
- [ ] Relational databases:
  - [ ] PostgreSQL: extensibility, JSONB, strong ecosystem
  - [ ] MySQL/InnoDB: replication maturity, wide adoption
  - [ ] When relational fits: ACID needs, complex queries, joins
  - [ ] Kubernetes deployment: CloudNativePG, Percona Operator
- [ ] Document databases:
  - [ ] MongoDB: flexible schema, horizontal scaling, aggregation pipeline
  - [ ] FerretDB: MongoDB-compatible wire protocol on PostgreSQL
  - [ ] When document fits: varied schemas, embedded data, rapid iteration
  - [ ] Trade-offs: no joins, denormalization cost, consistency tuning
- [ ] Wide-column stores:
  - [ ] Cassandra: tunable consistency, linear write scaling
  - [ ] ScyllaDB: Cassandra-compatible with C++ performance
  - [ ] When wide-column fits: high write throughput, time-series-like access
  - [ ] Trade-offs: limited query model, no joins, compaction overhead
- [ ] Key-value stores:
  - [ ] Redis: in-memory speed, data structures, pub/sub
  - [ ] DragonflyDB: Redis-compatible with multi-threaded architecture
  - [ ] etcd: Kubernetes backing store, strong consistency via Raft
  - [ ] When key-value fits: caching, sessions, counters, leaderboards
- [ ] Time-series databases:
  - [ ] TimescaleDB: PostgreSQL extension, continuous aggregates
  - [ ] InfluxDB: purpose-built, Flux query language
  - [ ] VictoriaMetrics: Prometheus-compatible, efficient compression
  - [ ] When time-series fits: metrics, IoT, financial ticks
- [ ] Graph databases:
  - [ ] Neo4j: native graph storage, Cypher query language
  - [ ] Apache AGE: graph extension for PostgreSQL
  - [ ] When graph fits: relationships are the query (social, fraud, knowledge)
  - [ ] Trade-offs: scaling challenges, limited ecosystem on Kubernetes
- [ ] Embedded and analytical databases:
  - [ ] DuckDB: in-process OLAP, columnar, zero-dependency
  - [ ] SQLite: embedded relational, single-file, edge/mobile
  - [ ] When embedded fits: analytics sidecars, batch processing, edge
  - [ ] Kubernetes pattern: DuckDB as sidecar for local analytics
- [ ] Vector databases:
  - [ ] pgvector: PostgreSQL extension for vector similarity search
  - [ ] Milvus: purpose-built, GPU-accelerated ANN search
  - [ ] Qdrant: Rust-based, filtering with vector search
  - [ ] When vector fits: embeddings, RAG, recommendation, image search
- [ ] Selection decision framework:
  - [ ] Query pattern analysis (OLTP, OLAP, search, graph traversal)
  - [ ] Consistency and durability requirements
  - [ ] Scale dimensions (data volume, read QPS, write QPS)
  - [ ] Operational maturity and Kubernetes operator availability
- [ ] Benchmark representative databases:
  - [ ] Deploy PostgreSQL, MongoDB, Cassandra, Redis on same Kubernetes cluster
  - [ ] Run comparable workload (insert, point read, range scan, update)
  - [ ] Measure throughput, latency, resource consumption per database
  - [ ] Document where each excels and where each struggles
- [ ] **ADR:** Document database category selection criteria and decision matrix

---

### J.14 NewSQL & Distributed SQL

**Goal:** Evaluate NewSQL databases on Kubernetes and understand when distributed SQL justifies its complexity over single-node PostgreSQL

**Learning objectives:**
- Understand the internal architecture of CockroachDB, TiDB, and YugabyteDB
- Benchmark NewSQL databases against single-node PostgreSQL on Kubernetes
- Determine when NewSQL is worth the operational overhead

**Tasks:**
- [ ] Create `experiments/newsql-distributed-sql/`
- [ ] CockroachDB architecture:
  - [ ] Ranges: automatic range-based sharding of tables
  - [ ] Raft consensus per range (leaseholder + replicas)
  - [ ] Distributed SQL layer: query planning across ranges
  - [ ] Transaction model: serializable isolation via timestamp ordering
- [ ] CockroachDB on Kubernetes:
  - [ ] CockroachDB Operator deployment
  - [ ] Node locality and topology-aware replication
  - [ ] Rolling upgrades and decommissioning nodes
  - [ ] Backup and restore with BACKUP/RESTORE commands
- [ ] TiDB architecture:
  - [ ] TiDB server (stateless SQL layer)
  - [ ] TiKV (distributed key-value storage, Raft groups)
  - [ ] PD (Placement Driver) for scheduling and metadata
  - [ ] TiFlash: columnar replica for HTAP workloads
- [ ] TiDB on Kubernetes:
  - [ ] TiDB Operator deployment and scaling
  - [ ] Separate scaling of TiDB, TiKV, and PD components
  - [ ] Resource requirements: minimum viable cluster
  - [ ] Monitoring with TiDB Dashboard and Grafana
- [ ] YugabyteDB architecture:
  - [ ] DocDB: document store built on RocksDB
  - [ ] Raft consensus per tablet (shard)
  - [ ] YSQL (PostgreSQL-compatible) and YCQL (Cassandra-compatible) APIs
  - [ ] Tablet splitting and load balancing
- [ ] YugabyteDB on Kubernetes:
  - [ ] YugabyteDB Operator or Helm deployment
  - [ ] Master and TServer pod topology
  - [ ] Multi-zone and multi-region configuration
  - [ ] xCluster replication for disaster recovery
- [ ] Spanner-inspired design patterns:
  - [ ] TrueTime and bounded clock uncertainty (Spanner)
  - [ ] Hybrid Logical Clocks as alternative (CockroachDB)
  - [ ] External consistency vs serializable isolation
  - [ ] Why open-source systems approximate but don't replicate Spanner
- [ ] Comparison with traditional sharding:
  - [ ] Application-transparent sharding (NewSQL) vs explicit shard routing (Vitess/Citus)
  - [ ] Cross-shard transactions: automatic (NewSQL) vs saga/2PC (traditional)
  - [ ] Operational complexity: one system vs shard-aware application
  - [ ] Migration path: PostgreSQL → Citus vs PostgreSQL → CockroachDB
- [ ] Trade-offs vs single-node PostgreSQL:
  - [ ] Latency overhead: consensus round-trips per write
  - [ ] Query compatibility: PostgreSQL feature coverage gaps
  - [ ] Resource consumption: 3-node minimum vs single pod
  - [ ] When single-node PostgreSQL with replicas is enough
- [ ] Benchmark NewSQL databases:
  - [ ] Deploy CockroachDB (3-node), TiDB, and single-node PostgreSQL
  - [ ] Run pgbench TPC-B equivalent on each
  - [ ] Measure write latency, read throughput, and tail latencies
  - [ ] Compare resource cost per transaction on Kubernetes
- [ ] **ADR:** Document NewSQL evaluation and adoption criteria

---

### J.15 Database Operators on Kubernetes

**Goal:** Evaluate Kubernetes database operators by deploying each and benchmarking their operational capabilities

**Learning objectives:**
- Deploy and compare major PostgreSQL, MySQL, and MongoDB operators
- Evaluate operators across HA, backup, scaling, monitoring, and upgrade dimensions
- Identify operator anti-patterns that lead to data loss or operational fragility

**Tasks:**
- [ ] Create `experiments/db-operators/`
- [ ] CloudNativePG:
  - [ ] Architecture: single operator, no sidecar, direct PostgreSQL management
  - [ ] HA: automated failover with pg_rewind, fencing
  - [ ] Backup: Barman-based continuous backup to object storage
  - [ ] Features: declarative clusters, rolling updates, connection pooling (PgBouncer)
- [ ] Zalando Postgres Operator:
  - [ ] Architecture: Patroni-based HA within pods
  - [ ] HA: Patroni leader election via Kubernetes endpoints or etcd
  - [ ] Backup: WAL-E/WAL-G to S3/GCS
  - [ ] Features: logical backups, connection pooling, team API integration
- [ ] CrunchyData PGO (Postgres Operator):
  - [ ] Architecture: pgBackRest for backup, Patroni for HA
  - [ ] HA: distributed topology support, proxy routing via PgBouncer
  - [ ] Backup: pgBackRest with full/incremental/differential
  - [ ] Features: monitoring integration (pgMonitor), PostGIS support
- [ ] Percona Operators:
  - [ ] Percona Operator for PostgreSQL (based on CrunchyData PGO)
  - [ ] Percona Operator for MySQL (Percona XtraDB Cluster, Group Replication)
  - [ ] Percona Operator for MongoDB (replica sets, sharded clusters)
  - [ ] Common patterns: backup to S3, monitoring, TLS automation
- [ ] Vitess operator:
  - [ ] Architecture: vtgate, vttablet, vtctld components
  - [ ] Horizontal sharding with transparent query routing
  - [ ] Online DDL and schema management
  - [ ] Monitoring and Grafana dashboards
- [ ] Operator comparison matrix:
  - [ ] HA capabilities: failover time, fencing, split-brain prevention
  - [ ] Backup: frequency, PITR granularity, restore time
  - [ ] Scaling: vertical (CPU/memory), horizontal (read replicas, shards)
  - [ ] Monitoring: built-in dashboards, Prometheus metrics, alerting
- [ ] Operator deployment benchmarks:
  - [ ] Time to provision a 3-node HA cluster for each operator
  - [ ] Failover time: kill primary pod, measure recovery duration
  - [ ] Backup and restore: time to backup, time to restore
  - [ ] Rolling upgrade: measure downtime during PostgreSQL minor version upgrade
- [ ] Operator anti-patterns:
  - [ ] Running stateful databases on ephemeral storage
  - [ ] Ignoring PodDisruptionBudgets for database pods
  - [ ] Skipping backup validation (untested backups are not backups)
  - [ ] Over-relying on operator magic: understand what happens underneath
- [ ] Deploy and compare two PostgreSQL operators:
  - [ ] Deploy CloudNativePG and Zalando Operator side by side
  - [ ] Run identical workloads, compare failover and backup behavior
  - [ ] Evaluate operational experience: CRD ergonomics, troubleshooting
- [ ] **ADR:** Document database operator selection and configuration standards

---

### J.16 Managed vs Self-Hosted Databases

**Goal:** Build a cost and operational comparison framework for managed vs self-hosted databases on Kubernetes

**Learning objectives:**
- Model total cost of ownership for managed services vs operator-managed databases
- Evaluate latency and performance trade-offs of in-cluster vs external databases
- Apply a decision framework accounting for team size, SLA requirements, and workload characteristics

**Tasks:**
- [ ] Create `experiments/db-managed-vs-selfhosted/`
- [ ] Cloud managed database services:
  - [ ] AWS RDS and Aurora (PostgreSQL, MySQL)
  - [ ] GCP Cloud SQL and AlloyDB
  - [ ] Azure Database for PostgreSQL Flexible Server
  - [ ] Cosmos DB (multi-model, globally distributed)
- [ ] Cost modeling dimensions:
  - [ ] Compute: instance size, reserved vs on-demand pricing
  - [ ] Storage: provisioned IOPS, throughput, allocated vs consumed
  - [ ] Network: cross-AZ transfer, VPC peering, egress charges
  - [ ] Licensing: open-source vs commercial (Oracle, SQL Server)
- [ ] Operational overhead comparison:
  - [ ] Managed: patching, backups, HA handled by provider
  - [ ] Self-hosted: operator handles HA/backup but team owns troubleshooting
  - [ ] Staff cost: DBA time for self-hosted vs managed service learning curve
  - [ ] Incident response: managed service SLA vs in-house on-call
- [ ] Latency considerations:
  - [ ] In-cluster database: sub-millisecond network latency
  - [ ] External managed database: cross-network hop, VPC peering latency
  - [ ] Connection pooling impact with external databases
  - [ ] Benchmarking: measure p50/p99 query latency in-cluster vs external
- [ ] Availability and durability:
  - [ ] Managed service SLAs (99.95%–99.999%)
  - [ ] Self-hosted HA with operators (depends on configuration)
  - [ ] Multi-AZ managed vs pod anti-affinity self-hosted
  - [ ] Backup durability: managed cross-region vs self-hosted to object storage
- [ ] Hybrid approaches:
  - [ ] Development/staging self-hosted, production managed
  - [ ] Read replicas in-cluster connected to managed primary
  - [ ] Crossplane provisioning managed databases alongside Kubernetes workloads
  - [ ] ExternalName services for transparent database endpoint switching
- [ ] Vendor lock-in considerations:
  - [ ] Managed-only features (Aurora Serverless, Cosmos DB multi-model)
  - [ ] Data portability: export/import paths
  - [ ] API compatibility: RDS PostgreSQL vs vanilla PostgreSQL
  - [ ] Exit cost: data transfer fees, migration effort
- [ ] Decision framework:
  - [ ] Team size and DBA expertise
  - [ ] SLA requirements and compliance constraints
  - [ ] Workload latency sensitivity
  - [ ] Cost sensitivity and budget predictability
- [ ] Build cost comparison model:
  - [ ] Model a reference workload (100GB, 5000 QPS, 3 replicas)
  - [ ] Price on RDS, Cloud SQL, Azure Database, and self-hosted (operator + nodes)
  - [ ] Include hidden costs: monitoring, backup storage, network transfer
  - [ ] Produce 1-year and 3-year TCO comparison
- [ ] **ADR:** Document managed vs self-hosted decision and criteria

---

### J.17 Database Benchmarking Methodology

**Goal:** Design and execute reproducible database benchmarks on Kubernetes with statistically valid results

**Learning objectives:**
- Use standard benchmarking tools (pgbench, sysbench, YCSB) correctly
- Account for Kubernetes-specific factors that affect benchmark results
- Design benchmark experiments with proper warm-up, steady-state, and percentile reporting

**Tasks:**
- [ ] Create `experiments/db-benchmarking/`
- [ ] pgbench (PostgreSQL):
  - [ ] Built-in TPC-B-like workload
  - [ ] Custom script mode for realistic workloads
  - [ ] Scaling factor and client count tuning
  - [ ] Interpreting results: TPS, latency distribution
- [ ] sysbench (MySQL and PostgreSQL):
  - [ ] OLTP read-only, read-write, and write-only workloads
  - [ ] Table count and size configuration
  - [ ] Thread scaling and concurrency testing
  - [ ] Lua scripting for custom workloads
- [ ] YCSB (Yahoo Cloud Serving Benchmark):
  - [ ] Core workloads A–F (update-heavy to scan-heavy)
  - [ ] Pluggable database bindings (Cassandra, MongoDB, Redis, PostgreSQL)
  - [ ] Zipfian vs uniform key distribution
  - [ ] Comparing databases with identical workload profiles
- [ ] TPC benchmarks:
  - [ ] TPC-C: OLTP order-processing benchmark
  - [ ] TPC-H: decision-support (analytical) benchmark
  - [ ] HammerDB as TPC-C/TPC-H runner
  - [ ] Interpreting results: tpmC, QphH, price/performance
- [ ] Kubernetes-specific benchmark factors:
  - [ ] Storage class impact: local-path vs Ceph vs EBS vs local NVMe
  - [ ] Network overhead: CNI plugin, service mesh sidecar latency
  - [ ] Resource limits: CPU throttling artifacts in latency measurements
  - [ ] Node placement: co-located vs cross-node client/server
- [ ] Benchmark experiment design:
  - [ ] Workload profiles: define read/write ratio, key distribution, record size
  - [ ] Warm-up phase: fill caches, stabilize buffer pool before measuring
  - [ ] Steady-state measurement: fixed duration after warm-up
  - [ ] Cool-down: capture background effects (compaction, checkpoints)
- [ ] Statistical rigor:
  - [ ] Multiple runs (minimum 3, ideally 5+)
  - [ ] Report p50, p95, p99, p99.9 latencies (not just averages)
  - [ ] Coefficient of variation to assess run-to-run stability
  - [ ] Confidence intervals and outlier identification
- [ ] Common benchmarking pitfalls:
  - [ ] Benchmarking cold caches (results not representative)
  - [ ] Insufficient data volume (fits entirely in memory)
  - [ ] Client bottleneck (benchmark tool saturates before database)
  - [ ] Ignoring background operations (compaction, WAL archiving, VACUUM)
- [ ] Reproducible benchmark experiments:
  - [ ] Containerized benchmark clients as Kubernetes Jobs
  - [ ] Parameterized Argo Workflows for benchmark execution
  - [ ] Results collection to Prometheus/VictoriaMetrics for comparison
  - [ ] Version-controlled benchmark configurations in experiment directory
- [ ] Execute reference benchmark suite:
  - [ ] Deploy PostgreSQL with CloudNativePG, run pgbench at multiple concurrency levels
  - [ ] Vary storage class and measure IOPS impact on TPS
  - [ ] Capture resource utilization (CPU, memory, disk I/O) during benchmark
  - [ ] Produce benchmark report with comparison charts in Grafana
- [ ] **ADR:** Document database benchmarking standards and reporting format

---
